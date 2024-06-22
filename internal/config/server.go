package config

import (
	"flag"
	"github.com/caarlos0/env/v6"
)

type ServerType struct {
	ServerAddr string `env:"SERVER_ADDRESS"`
	ResultAddr string `env:"BASE_URL"`
}

var Server ServerType

func init() {
	Server = NewServerConfig()
}

func NewServerConfig() ServerType {
	c := ServerType{}
	err := env.Parse(&c)
	if err != nil || c.ServerAddr == "" {
		flag.StringVar(&c.ServerAddr, "a", ":8080", "server port")
		flag.StringVar(&c.ResultAddr, "b", "http://localhost:8080", "port for short links")
		flag.Parse()
	}
	return c
}
