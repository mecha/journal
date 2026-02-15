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
	"os/signal"
	"path"
	"slices"
	"strings"
	"syscall"
	"time"

	"github.com/farmergreg/rfsnotify"
	"github.com/hashicorp/go-version"
	"github.com/mecha/journal/utils"
	"gopkg.in/fsnotify.v1"
)

var ErrIncorrectPassword = errors.New("Incorrect password")
var ErrMountNotEmpty = errors.New("Mount point is not empty")

type Journal struct {
	cipherPath  string
	mountPath   string
	idleTimeout string
	command     *exec.Cmd
	signals     chan os.Signal
	isMounted   bool
	errorChan   chan error
	onUnmount   func()
	onFSEvent   func(ev fsnotify.Event)
	watcher     *rfsnotify.RWatcher
}

func NewJournal(cipherPath, mountPath string, idleTimeout string) (*Journal, error) {
	journal := &Journal{
		cipherPath:  strings.TrimSuffix(cipherPath, "/"),
		mountPath:   strings.TrimSuffix(mountPath, "/"),
		idleTimeout: idleTimeout,
		command:     nil,
		signals:     make(chan os.Signal, 1),
		isMounted:   false,
		errorChan:   make(chan error),
		onUnmount:   nil,
		onFSEvent:   nil,
	}

	watcher, err := rfsnotify.NewWatcher()
	if err != nil {
		return journal, err
	}
	journal.watcher = watcher

	return journal, nil
}

func (j *Journal) Mount(password string) error {
	if j.isMounted {
		return errors.New("journal is already mounted")
	}

	os.MkdirAll(j.mountPath, 0755)

	j.command = exec.Command(
		"gocryptfs",
		"-fg",
		"-notifypid",
		fmt.Sprintf("%d", os.Getpid()),
		"-idle",
		j.idleTimeout,
		j.cipherPath,
		j.mountPath,
	)

	// for writing the password to the command over its STDIN
	stdin, err := j.command.StdinPipe()
	if err != nil {
		return err
	}
	defer stdin.Close()

	// get signal from gocryptfs when it has mounted
	signal.Notify(j.signals, syscall.SIGUSR1)
	defer signal.Stop(j.signals)

	err = j.command.Start()
	if err != nil {
		return err
	}

	_, err = io.WriteString(stdin, password+"\n")
	if err != nil {
		return err
	}

	go func() {
		j.errorChan <- j.command.Wait()
	}()

	select {
	// timeout, abort mission
	case <-time.NewTimer(3 * time.Second).C:
		return errors.New("timed out waiting for journal to mount")

	// got error, command has exited
	case err := <-j.errorChan:
		switch err := err.(type) {
		case *exec.ExitError:
			switch err.ExitCode() {
			case 12:
				return ErrIncorrectPassword
			case 10:
				return ErrMountNotEmpty
			}
		}
		return err

	// got signal, has mounted successfully
	case <-j.signals:
		j.isMounted = true

		// watch mounted path for fs events
		err := j.watcher.AddRecursive(j.mountPath)
		if err != nil {
			log.Println(err)
		}
		go j.handleWatcherEvents()

		// listen for errors from the command to unmount
		go func() {
			err := <-j.errorChan
			if err != nil {
				log.Printf("journal locked; %s", err.Error())
			}
			j.isMounted = false
			j.watcher.Close()
			if j.onUnmount != nil {
				j.onUnmount()
			}
			j.errorChan <- err
		}()

		return nil
	}
}

func (j *Journal) handleWatcherEvents() {
	for {
		select {
		case ev, ok := <-j.watcher.Events:
			if !ok {
				return
			}
			if j.onFSEvent != nil {
				j.onFSEvent(ev)
			}
		case err, ok := <-j.watcher.Errors:
			if !ok {
				return
			}
			log.Println(err)
		}
	}
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

	title := date.Format("Mon - 02 Jan 2006")
	_, err = file.WriteString("# " + title + "\n\n")
	if err != nil {
		return "", err
	}

	err = os.Chmod(filepath, 0740)
	return filepath, err
}

func (j *Journal) EditEntry(date time.Time, window bool) error {
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

	title := date.Format("02 Jan 2006")
	err := openEditor(filepath, title, window)

	return err
}

func (j *Journal) GetEntryAtPath(path string) (time.Time, error) {
	if !strings.HasPrefix(path, j.mountPath) {
		return time.Time{}, errors.New("not a path to a journal entry")
	}

	relpath := path[len(j.mountPath):]
	dateStr := strings.TrimSuffix(strings.TrimPrefix(relpath, "/"), ".md")

	date, err := utils.ParseYearMonthDay(dateStr)
	if err != nil {
		return date, err
	}
	return date, nil
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

func (j *Journal) SearchTag(tag string) ([]time.Time, error) {
	if !j.isMounted {
		return []time.Time{}, errors.New("journal is not mounted")
	}

	cmd := exec.Command("rg", "-l", "-w", tag, j.mountPath)
	cmd.Stderr = os.Stderr
	output, err := cmd.Output()
	if err != nil {
		return []time.Time{}, fmt.Errorf("rg error: %w", err)
	}

	fileMap := map[time.Time]bool{}
	scanner := bufio.NewScanner(bytes.NewBuffer(output))
	for scanner.Scan() {
		date, err := j.GetEntryAtPath(scanner.Text())
		if err != nil {
			continue
		}
		fileMap[date] = true
	}

	entries := slices.Collect(maps.Keys(fileMap))
	slices.SortFunc(entries, func(a, b time.Time) int { return a.Compare(b) })

	return entries, nil
}

func checkGCFSVersion(minVersion string) error {
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

	constraint, _ := version.NewConstraint(">= " + minVersion)
	if !constraint.Check(actualVersion) {
		return errors.New("gocryptfs version " + minVersion + " is required, found v" + versionStr)
	}

	return nil
}
