package main

import (
	"github.com/ViBiOh/httputils/v4/pkg/cors"
	"github.com/ViBiOh/httputils/v4/pkg/owasp"
	"github.com/ViBiOh/httputils/v4/pkg/server"
)

type services struct {
	server *server.Server
	cors   cors.Service
	owasp  owasp.Service
}

func newServices(config configuration) services {
	var output services

	output.server = server.New(config.server)
	output.cors = cors.New(config.cors)
	output.owasp = owasp.New(config.owasp)

	return output
}
