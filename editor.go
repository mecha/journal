package main

import (
	"errors"
	"log"
	"os"
	"os/exec"
)

var errNoEditor = errors.New("the $EDITOR environment variable is not set")
var errNoTmux = errors.New("cannot open editor, need to be in tmux")

func openEditor(filepath string, title string, window bool) error {
	editor, hasEditor := os.LookupEnv("EDITOR")
	if !hasEditor {
		return errNoEditor
	}

	_, isInTmux := os.LookupEnv("TMUX")
	if !isInTmux {
		return errNoTmux
	}

	var cmd *exec.Cmd
	if window {
		cmd = exec.Command("tmux", "neww", "-n", title, editor, filepath)
	} else {
		cmd = exec.Command("tmux", "display-popup", "-w", "100%", "-h", "100%", "-T", title, "-EE", editor, filepath)
	}

	err := cmd.Run()
	log.Printf("opened entry for editing in %s: %s", editor, filepath)

	return err
}
