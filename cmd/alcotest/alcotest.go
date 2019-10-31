package main

import (
	"flag"
	"log"
	"os"

	"github.com/ViBiOh/httputils/v3/pkg/alcotest"
)

func main() {
	fs := flag.NewFlagSet("alcotest", flag.ExitOnError)

	alcotestConfig := alcotest.Flags(fs, "")

	if err := fs.Parse(os.Args[1:]); err != nil {
		log.Fatalf("%s", err)
	}

	alcotest.DoAndExit(alcotestConfig)
}
