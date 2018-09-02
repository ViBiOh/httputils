package main

import (
	"flag"
	"io/ioutil"
	"log"
	"strings"

	"github.com/ViBiOh/httputils/pkg/cert"
)

func saveFile(filename, defaultFilename string, content []byte) {
	if filename == `` {
		filename = defaultFilename
	}

	if err := ioutil.WriteFile(filename, content, 0600); err != nil {
		log.Fatalf(`Error while writing %s: %v`, filename, err)
	}

	log.Printf(`File %s created`, filename)
}

func main() {
	certConfig := cert.Flags(``)
	flag.Parse()

	cert, key, err := cert.GenerateFromConfig(certConfig)
	if err != nil {
		log.Fatalf(`Error while generating certificate: %v`, err)
	}

	saveFile(strings.TrimSpace(*certConfig[`cert`]), `public.crt`, cert)
	saveFile(strings.TrimSpace(*certConfig[`key`]), `private.key`, key)
}
