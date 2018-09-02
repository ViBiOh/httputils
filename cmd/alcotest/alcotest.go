package main

import (
	"flag"

	"github.com/ViBiOh/httputils/pkg/alcotest"
)

func main() {
	alcotestConfig := alcotest.Flags(``)
	flag.Parse()

	alcotest.DoAndExit(alcotestConfig)
}
