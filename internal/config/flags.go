package config

import "flag"

type ServerConfigFlags struct {
	ServerAddr      string
	ResultAddr      string
	FileStoragePath string
	DatabaseDSN     string
	JWTSecret       string
}

var Flags = NewServerConfigFlags()

func (scf *ServerConfigFlags) initFlags() {
	flag.StringVar(&scf.ServerAddr, "a", "localhost:8080", "server address")
	flag.StringVar(&scf.ResultAddr, "b", "http://localhost:8080", "result base url")
	flag.StringVar(&scf.FileStoragePath, "f", "", "file storage path")
	flag.StringVar(&scf.DatabaseDSN, "d", "", "postgres connection string")
	flag.StringVar(&scf.JWTSecret, "j", "secret", "jwt secret")
}

func NewServerConfigFlags() *ServerConfigFlags {
	scf := &ServerConfigFlags{}
	scf.initFlags()

	return scf
}
