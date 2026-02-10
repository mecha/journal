package main

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
	"strings"
	"syscall"
	"time"

	"github.com/mecha/journal/utils"

	"github.com/farmergreg/rfsnotify"
	"github.com/hashicorp/go-version"
	"gopkg.in/fsnotify.v1"
)

const minGoCryptFSVersion = "2.6.1"
const MountWaitTime = time.Millisecond * 150

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
	onFSEventFunc func(ev fsnotify.Event)
	watcher       *rfsnotify.RWatcher
}

func NewJournal(cipherPath, mountPath string, idleTimeout string) (*Journal, error) {
	journal := &Journal{
		cipherPath:    strings.TrimSuffix(cipherPath, "/"),
		mountPath:     strings.TrimSuffix(mountPath, "/"),
		idleTimeout:   idleTimeout,
		command:       nil,
		isMounted:     false,
		errorChan:     make(chan error),
		onUnmountFunc: nil,
		onFSEventFunc: nil,
	}

	watcher, err := rfsnotify.NewWatcher()
	if err != nil {
		return journal, err
	}
	journal.watcher = watcher

	return journal, nil
}

func (j *Journal) IsMounted() bool {
	return j.isMounted
}

func (j *Journal) OnUnmount(callback func()) {
	j.onUnmountFunc = callback
}

func (j *Journal) OnFSEvent(callback func(ev fsnotify.Event)) {
	j.onFSEventFunc = callback
}

func (j *Journal) Mount(password string) error {
	if j.isMounted {
		return errors.New("journal is already mounted")
	}

	os.MkdirAll(j.mountPath, 0755)

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

	var execError error = nil
	var hasReturned = false
	go func() {
		execError = j.command.Wait()
		if execError != nil && hasReturned {
			log.Printf("journal locked; %s", execError.Error())
		}
		j.isMounted = false
		j.watcher.Close()
		select {
		case j.errorChan <- execError:
		default:
		}
		if hasReturned && j.onUnmountFunc != nil {
			j.onUnmountFunc()
		}
	}()

	err = nil
	select {
	case err = <-j.errorChan:
	case <-time.NewTimer(MountWaitTime).C:
	}

	hasReturned = true
	if err != nil {
		return err
	}

	err = j.watcher.AddRecursive(j.mountPath)
	if err != nil {
		log.Println(err)
	}

	go func() {
		for {
			select {
			case ev, ok := <-j.watcher.Events:
				if !ok {
					return
				}
				if j.onFSEventFunc != nil {
					j.onFSEventFunc(ev)
				}
			case err, ok := <-j.watcher.Errors:
				if !ok {
					return
				}
				log.Println(err)
			}
		}
	}()

	return nil
}

func (j *Journal) Unmount() error {
	if !j.isMounted || j.command == nil {
		return nil
	}

	j.command.Process.Signal(syscall.SIGTERM)
	j.watcher.Close()

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

func (j *Journal) EntryPath(date time.Time) string {
	day, month, year := date.Day(), int(date.Month()), date.Year()
	return fmt.Sprintf("%s/%02d/%02d/%02d.md", j.mountPath, year, month, day)
}

func (j *Journal) HasEntry(date time.Time) (bool, error) {
	if !j.isMounted {
		return false, nil
	}

	filepath := j.EntryPath(date)
	_, err := os.Stat(filepath)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func (j *Journal) GetEntry(date time.Time) (string, bool, error) {
	if !j.isMounted {
		return "", false, nil
	}

	filepath := j.EntryPath(date)
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

func (j *Journal) CreateEntry(date time.Time) (string, error) {
	if !j.isMounted {
		return "", errors.New("journal is not mounted")
	}

	filepath := j.EntryPath(date)
	dirpath := path.Dir(filepath)

	err := os.MkdirAll(dirpath, 0740)
	if err != nil {
		return "", err
	}

	err = j.watcher.Add(dirpath)
	if err != nil {
		log.Println(err)
	}

	file, err := os.Create(filepath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	title := j.EntryTitle(date)
	_, err = file.WriteString("# " + title + "\n\n")
	if err != nil {
		return "", err
	}

	err = os.Chmod(filepath, 0740)
	return filepath, err
}

func (j *Journal) EditEntry(date time.Time) error {
	if !j.isMounted {
		return errors.New("journal is not mounted")
	}

	filepath := j.EntryPath(date)
	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		_, err := j.CreateEntry(date)
		if err != nil {
			return err
		}
		log.Printf("created new entry: %s", filepath)
	}

	winTitle := date.Format("02 Jan 2006")
	cmd := exec.Command("tmux", "neww", "-n", winTitle, "nvim", filepath)
	cmd.Run()
	// log.Printf("opened entry for editing: %s", filepath)

	return nil
}

func (j *Journal) EntryTitle(date time.Time) string {
	return date.Format("Mon - 02 Jan 2006")
}

func (j *Journal) GetEntryAtPath(path string) (day, month, year int) {
	if !strings.HasPrefix(path, j.mountPath) {
		return 0, 0, 0
	}

	relpath := path[len(j.mountPath):]
	dateStr := strings.TrimSuffix(strings.TrimPrefix(relpath, "/"), ".md")

	year, month, day, err := utils.ParseDayMonthYear(dateStr)
	if err != nil {
		return 0, 0, 0
	}
	return day, month, year
}

func (j *Journal) DeleteEntry(date time.Time) error {
	path := j.EntryPath(date)
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

func (j *Journal) handleFSEvent(ev fsnotify.Event) {
	if j.onFSEventFunc != nil {
		j.onFSEventFunc(ev)
	}
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
