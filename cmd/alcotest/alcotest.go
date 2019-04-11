package main

import (
	"flag"
	"os"

	"github.com/ViBiOh/httputils/pkg/alcotest"
	"github.com/ViBiOh/httputils/pkg/logger"
)

func main() {
	fs := flag.NewFlagSet("alcotest", flag.ExitOnError)

	alcotestConfig := alcotest.Flags(fs, "")

	if err := fs.Parse(os.Args[1:]); err != nil {
		logger.Fatal("%+v", err)
	}

	alcotest.DoAndExit(alcotestConfig)
}
