package main

import (
	"flag"

	"github.com/ViBiOh/httputils/pkg/alcotest"
)

func main() {
	fs := flag.NewFlagSet(`alcotest`, flag.ExitOnError)

	alcotestConfig := alcotest.Flags(fs, ``)
	flag.Parse()

	alcotest.DoAndExit(alcotestConfig)
}
