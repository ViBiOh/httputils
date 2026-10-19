package main

import (
	"flag"
	"os"

	"github.com/ViBiOh/flags"
	"github.com/ViBiOh/httputils/v4/pkg/alcotest"
)

func main() {
	flagSet := flag.NewFlagSet("alcotest", flag.ExitOnError)
	flagSet.Usage = flags.Usage(flagSet)

	alcotestConfig := alcotest.Flags(flagSet, "")

	_ = flagSet.Parse(os.Args[1:])

	alcotest.DoAndExit(alcotestConfig)
}
