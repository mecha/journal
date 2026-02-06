package journal

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"maps"
	"os"
	"os/exec"
	"path"
	"slices"
	"strconv"
	"strings"
	"syscall"
	"time"

	version "github.com/hashicorp/go-version"
)

// TODO: move to main?

const minGoCryptFSVersion = "2.6.1"

func init() {
	err := checkMinGoCryptFSVersion()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

type Journal struct {
	cipherPath    string
	mountPath     string
	idleTimeout   string
	command       *exec.Cmd
	isMounted     bool
	errorChan     chan error
	onUnmountFunc func()
}

func NewJournal(cipherPath, mountPath string, idleTimeout string) *Journal {
	return &Journal{
		cipherPath:    strings.TrimSuffix(cipherPath, "/"),
		mountPath:     strings.TrimSuffix(mountPath, "/"),
		idleTimeout:   idleTimeout,
		command:       nil,
		isMounted:     false,
		errorChan:     make(chan error),
		onUnmountFunc: func() {},
	}
}

func (j *Journal) IsMounted() bool {
	return j.isMounted
}

func (j *Journal) OnUnmount(callback func()) {
	j.onUnmountFunc = callback
}

func (j *Journal) Mount(password string) error {
	command := exec.Command(
		"gocryptfs",
		"-fg",
		"-q",
		"-idle",
		j.idleTimeout,
		j.cipherPath,
		j.mountPath,
	)

	stdin, err := command.StdinPipe()
	if err != nil {
		return err
	}
	defer stdin.Close()

	err = command.Start()
	if err != nil {
		return err
	}

	_, err = io.WriteString(stdin, password+"\n")
	if err != nil {
		return err
	}

	j.isMounted = true
	j.command = command

	go func() {
		err := j.command.Wait()
		log.Printf("gocryptfs exited, journal has been unmounted: %s", err.Error())
		j.isMounted = false
		select {
		case j.errorChan <- err:
			// error sent to receiver
		default:
			// channel has no receiver, but we don't care so do nothing
		}
		j.onUnmountFunc()
	}()

	return nil
}

func (j *Journal) Unmount() error {
	if !j.isMounted || j.command == nil {
		return nil
	}

	_ = j.command.Process.Signal(syscall.SIGTERM)

	select {
	case err := <-j.errorChan:
		if err != nil && j.command.ProcessState.ExitCode() != 15 {
			log.Println(err)
		}
	case <-time.After(3 * time.Second):
		log.Println("timed out waiting for gocryptfs to exit")
	}

	return nil
}

func (j *Journal) EntryPath(day int, month int, year int) string {
	return fmt.Sprintf("%s/%02d/%02d/%02d.md", j.mountPath, year, month, day)
}

func (j *Journal) HasEntry(day int, month int, year int) (bool, error) {
	if !j.isMounted {
		return false, nil
	}

	filepath := j.EntryPath(day, month, year)
	_, err := os.Stat(filepath)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func (j *Journal) GetEntry(day, month, year int) (string, bool, error) {
	if !j.isMounted {
		return "", false, nil
	}

	filepath := j.EntryPath(day, month, year)
	content, err := os.ReadFile(filepath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", false, nil
		} else {
			return "", false, nil
		}
	}

	return string(content), true, nil
}

func (j *Journal) CreateEntry(day int, month int, year int) (string, error) {
	if !j.isMounted {
		return "", errors.New("journal is not mounted")
	}

	filepath := j.EntryPath(day, month, year)
	dirpath := path.Dir(filepath)

	err := os.MkdirAll(dirpath, 0740)
	if err != nil {
		return "", err
	}

	file, err := os.Create(filepath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	title := j.EntryTitle(day, month, year)
	_, err = file.WriteString("# " + title + "\n\n")
	if err != nil {
		return "", err
	}

	err = os.Chmod(filepath, 0740)
	return filepath, err
}

func (j *Journal) EditEntry(day, month, year int) error {
	if !j.isMounted {
		return errors.New("journal is not mounted")
	}

	filepath := j.EntryPath(day, month, year)
	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		_, err := j.CreateEntry(day, month, year)
		if err != nil {
			return err
		}
	}

	date := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.Local)
	winTitle := date.Format("02 Jan 2006")
	cmd := exec.Command("tmux", "neww", "-n", winTitle, "nvim", filepath)
	cmd.Run()

	return nil
}

func (j *Journal) EntryTitle(day, month, year int) string {
	date := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.Local)
	title := date.Format("Mon - 02 Jan 2006")
	return title
}

func (j *Journal) GetEntryAtPath(path string) (day, month, year int) {
	if !strings.HasPrefix(path, j.mountPath) {
		return 0, 0, 0
	}

	relpath := path[len(j.mountPath):]

	parts := strings.Split(relpath, "/")
	if len(parts) < 3 {
		return 0, 0, 0
	}
	parts = parts[len(parts)-3:]

	var err error
	year, err = strconv.Atoi(parts[0])
	month, err = strconv.Atoi(parts[1])
	day, err = strconv.Atoi(strings.TrimSuffix(parts[2], ".md"))

	if err != nil {
		return 0, 0, 0
	}
	return day, month, year
}

func (j *Journal) DeleteEntry(day, month, year int) error {
	path := j.EntryPath(day, month, year)
	err := os.Remove(path)
	return err
}

func (j *Journal) Tags() ([]string, error) {
	if !j.isMounted {
		return []string{}, errors.New("journal is not mounted")
	}

	cmd := exec.Command("rg", "--only-matching", "--no-filename", "--no-line-number", "@[^\\W]+", j.mountPath)
	cmd.Stderr = os.Stderr
	output, err := cmd.Output()
	if err != nil {
		return []string{}, fmt.Errorf("rg error: %w", err)
	}

	tags := map[string]bool{}
	scanner := bufio.NewScanner(bytes.NewBuffer(output))
	for scanner.Scan() {
		tags[scanner.Text()] = true
	}

	return slices.Collect(maps.Keys(tags)), nil
}

func (j *Journal) SearchTag(tag string) ([]string, error) {
	if !j.isMounted {
		return []string{}, errors.New("journal is not mounted")
	}

	cmd := exec.Command("rg", "-l", "-w", tag, j.mountPath)
	cmd.Stderr = os.Stderr
	output, err := cmd.Output()
	if err != nil {
		return []string{}, fmt.Errorf("rg error: %w", err)
	}

	fileMap := map[string]bool{}
	scanner := bufio.NewScanner(bytes.NewBuffer(output))
	for scanner.Scan() {
		fileMap[scanner.Text()] = true
	}

	files := slices.Collect(maps.Keys(fileMap))
	slices.Sort(files)

	return files, nil
}

func checkMinGoCryptFSVersion() error {
	cmd := exec.Command("gocryptfs", "-version")
	output, err := cmd.Output()
	if err != nil {
		return err
	}

	parts := strings.Split(string(output), ";")
	if len(parts) < 1 {
		return errors.New("unexpected output from 'gocryptfs -version': " + string(output))
	}

	versionStr, found := strings.CutPrefix(parts[0], "gocryptfs v")
	if !found {
		return errors.New("invalid gocryptfs version: " + parts[0])
	}

	actualVersion, err := version.NewVersion(versionStr)
	if err != nil {
		return err
	}

	constraint, _ := version.NewConstraint(">= " + minGoCryptFSVersion)
	if !constraint.Check(actualVersion) {
		return errors.New("gocryptfs version " + minGoCryptFSVersion + " is required, found v" + versionStr)
	}

	return nil
}
