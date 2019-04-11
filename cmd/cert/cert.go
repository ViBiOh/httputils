package main

import (
	"flag"
	"io/ioutil"
	"os"
	"strings"

	"github.com/ViBiOh/httputils/pkg/cert"
	"github.com/ViBiOh/httputils/pkg/errors"
	"github.com/ViBiOh/httputils/pkg/logger"
)

func saveFile(filename, defaultFilename string, content []byte) {
	if filename == "" {
		filename = defaultFilename
	}

	if err := ioutil.WriteFile(filename, content, 0600); err != nil {
		logger.Fatal("%+v", errors.WithStack(err))
	}

	logger.Info("File %s created", filename)
}

func main() {
	fs := flag.NewFlagSet("cert", flag.ExitOnError)

	certConfig := cert.Flags(fs, "")

	if err := fs.Parse(os.Args[1:]); err != nil {
		logger.Fatal("%+v", err)
	}

	cert, key, err := cert.GenerateFromConfig(certConfig)
	if err != nil {
		logger.Fatal("%+v", err)
	}

	saveFile(strings.TrimSpace(*certConfig.Cert), "public.crt", cert)
	saveFile(strings.TrimSpace(*certConfig.Key), "private.key", key)
}
