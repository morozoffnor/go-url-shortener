package config

import "flag"

type ServerConfigFlags struct {
	ServerAddr      string
	ResultAddr      string
	FileStoragePath string
}

var Flags = ServerConfigFlags{}

func init() {
	flag.StringVar(&Flags.ServerAddr, "a", "localhost:8080", "server address")
	flag.StringVar(&Flags.ResultAddr, "b", "http://localhost:8080", "result base url")
	flag.StringVar(&Flags.FileStoragePath, "f", "/tmp/short-url-db.json", "file storage path")
}
