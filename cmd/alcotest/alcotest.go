package main

import (
	"context"
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
		slog.LogAttrs(context.Background(), slog.LevelError, "parse flag", slog.Any("error", err))
		os.Exit(1)
	}

	alcotest.DoAndExit(alcotestConfig)
}
