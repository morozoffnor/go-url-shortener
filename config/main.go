package config

import "flag"

type ServerType struct {
	ServerAddr string
	ResultAddr string
}

var Server ServerType

func init() {
	Server = NewServerConfig()
}

func NewServerConfig() ServerType {
	c := ServerType{}
	flag.StringVar(&c.ServerAddr, "a", ":8080", "server port")
	flag.StringVar(&c.ResultAddr, "b", ":8080", "port for short links")
	flag.Parse()
	return c
}
