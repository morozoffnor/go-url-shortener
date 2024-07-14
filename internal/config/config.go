package config

import (
	"flag"
	"os"
	"strings"
)

type Config struct {
	ResultAddr      string
	ServerAddr      string
	FileStoragePath string
	DatabaseDSN     string
}

func (c *Config) UpdateByOptions(o *ServerConfigFlags) {
	flag.Parse()
	c.ServerAddr = o.ServerAddr
	c.ResultAddr = strings.Trim(o.ResultAddr, "/")
	c.FileStoragePath = o.FileStoragePath
}

func (c *Config) PopulateConfigFromEnv() {

	sa := os.Getenv("SERVER_ADDRESS")
	if sa != "" {
		c.ServerAddr = sa
	}
	ra := os.Getenv("BASE_URL")
	if ra != "" {
		c.ResultAddr = ra
	}
	fsp := os.Getenv("FILE_STORAGE_PATH")
	if fsp != "" {
		c.FileStoragePath = fsp
	}
	ddsn := os.Getenv("DATABASE_DSN")
	if ddsn != "" {
		c.DatabaseDSN = ddsn
	}
}

func New() *Config {
	c := &Config{
		ServerAddr:      ":8080",
		ResultAddr:      "http://localhost:8080",
		FileStoragePath: "/tmp/test.json",
	}
	c.UpdateByOptions(Flags)
	c.PopulateConfigFromEnv()
	return c
}
