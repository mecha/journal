package main

import (
	"errors"
	"log"
	"os"
	"os/exec"
)

var errNoEditor = errors.New("the $EDITOR environment variable is not set")
var errNoTmux = errors.New("cannot open editor, need to be in tmux")

func openInEditor(filepath string, winTitle string) error {
	editor, hasEditor := os.LookupEnv("EDITOR")
	if !hasEditor {
		return errNoEditor
	}

	var cmd *exec.Cmd

	_, isInTmux := os.LookupEnv("TMUX")
	if !isInTmux {
		return errNoTmux
	}

	cmd = exec.Command("tmux", "neww", "-n", winTitle, editor, filepath)
	err := cmd.Run()
	log.Printf("opened entry for editing in %s: %s", editor, filepath)

	return err
}
