package main

import (
	"github.com/ViBiOh/httputils/v4/pkg/server"
)

type services struct {
	server *server.Server
}

func newServices(config configuration) services {
	return services{
		server: server.New(config.server),
	}
}
