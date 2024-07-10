package config

import (
	"os"
	"strings"
)

type ServerConfig struct {
	ResultAddr      string
	ServerAddr      string
	FileStoragePath string
}

func (config *ServerConfig) UpdateByOptions(o *ServerConfigFlags) {
	config.ServerAddr = o.ServerAddr
	config.ResultAddr = strings.Trim(o.ResultAddr, "/")
	config.FileStoragePath = o.FileStoragePath
}

func (config *ServerConfig) PopulateConfigFromEnv() {

	sa := os.Getenv("SERVER_ADDRESS")
	if sa != "" {
		config.ServerAddr = sa
	}
	ra := os.Getenv("BASE_URL")
	if ra != "" {
		config.ResultAddr = ra
	}
	fsp := os.Getenv("FILE_STORAGE_PATH")
	if fsp != "" {
		config.FileStoragePath = fsp
	}
}

func New() *ServerConfig {
	c := &ServerConfig{
		ServerAddr:      ":8080",
		ResultAddr:      "http://localhost:8080",
		FileStoragePath: "/tmp/test.json",
	}
	c.UpdateByOptions(Flags)
	c.PopulateConfigFromEnv()
	return c
}
