package main

import (
	"flag"
	"log"
	"os"
	"strings"
)

var Flags struct {
	path        string
	mntPath     string
	idleTimeout string
}

func parseFlags() {
	flag.StringVar(&Flags.mntPath, "m", "/tmp/journal", "The path to the directory where the journal will be mounted.")
	flag.StringVar(&Flags.idleTimeout, "idle", "30m", "The journal will be unmounted after some time without any operations. Examples: 30s, 5m, 1h")
	flag.Parse()

	Flags.path = strings.TrimSpace(flag.Arg(0))
	if len(Flags.path) == 0 {
		path, hasEnv := os.LookupEnv("JOURNAL_ENC_DIR")
		if hasEnv {
			Flags.path = path
		} else {
			log.Fatal("no path argument specified and the env variable is not set")
		}
	}
}
