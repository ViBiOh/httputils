package main

import (
	"flag"
	"log/slog"
	"os"

	"github.com/ViBiOh/flags"
	"github.com/ViBiOh/httputils/v4/pkg/alcotest"
)

func main() {
	fs := flag.NewFlagSet("alcotest", flag.ExitOnError)
	fs.Usage = flags.Usage(fs)

	alcotestConfig := alcotest.Flags(fs, "")

	if err := fs.Parse(os.Args[1:]); err != nil {
		slog.Error("parse flag", "err", err)
		os.Exit(1)
	}

	alcotest.DoAndExit(alcotestConfig)
}
