package main

import (
	"flag"
	"os"

	"github.com/ViBiOh/flags"
	"github.com/ViBiOh/httputils/v4/pkg/alcotest"
)

func main() {
	fs := flag.NewFlagSet("alcotest", flag.ExitOnError)
	fs.Usage = flags.Usage(fs)

	alcotestConfig := alcotest.Flags(fs, "")

	_ = fs.Parse(os.Args[1:])

	alcotest.DoAndExit(alcotestConfig)
}
