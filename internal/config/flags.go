package config

import "flag"

type ServerConfigFlags struct {
	ServerAddr      string
	ResultAddr      string
	FileStoragePath string
}

var Flags = NewServerConfigFlags()

func (scf *ServerConfigFlags) initFlags() {
	flag.StringVar(&scf.ServerAddr, "a", "localhost:8080", "server address")
	flag.StringVar(&scf.ResultAddr, "b", "http://localhost:8080", "result base url")
	flag.StringVar(&scf.FileStoragePath, "f", "/tmp/short-url-db.json", "file storage path")
}

func NewServerConfigFlags() *ServerConfigFlags {
	scf := &ServerConfigFlags{}
	scf.initFlags()

	return scf
}
