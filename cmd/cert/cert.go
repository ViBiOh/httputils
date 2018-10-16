package main

import (
	"flag"
	"io/ioutil"
	"strings"

	"github.com/ViBiOh/httputils/pkg/cert"
	"github.com/ViBiOh/httputils/pkg/logger"
)

func saveFile(filename, defaultFilename string, content []byte) {
	if filename == `` {
		filename = defaultFilename
	}

	if err := ioutil.WriteFile(filename, content, 0600); err != nil {
		logger.Fatal(`error while writing %s: %v`, filename, err)
	}

	logger.Info(`File %s created`, filename)
}

func main() {
	certConfig := cert.Flags(``)
	flag.Parse()

	cert, key, err := cert.GenerateFromConfig(certConfig)
	if err != nil {
		logger.Fatal(`error while generating certificate: %v`, err)
	}

	saveFile(strings.TrimSpace(*certConfig[`cert`]), `public.crt`, cert)
	saveFile(strings.TrimSpace(*certConfig[`key`]), `private.key`, key)
}
