package main

import (
	"flag"
	"log"
	"os"

	"github.com/ViBiOh/httputils/pkg/alcotest"
)

func main() {
	fs := flag.NewFlagSet("alcotest", flag.ExitOnError)

	alcotestConfig := alcotest.Flags(fs, "")

	if err := fs.Parse(os.Args[1:]); err != nil {
		log.Fatalf("%#v", err)
	}

	alcotest.DoAndExit(alcotestConfig)
}
