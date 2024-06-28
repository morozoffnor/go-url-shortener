package main

import (
	"flag"
	"github.com/morozoffnor/go-url-shortener/internal/config"
	"github.com/morozoffnor/go-url-shortener/internal/server"
)

func main() {
	flag.Parse()
	config.Server = config.NewConfig()

	config.Server.UpdateByOptions(config.Flags)
	config.Server.PopulateConfigFromEnv()
	if err := server.RunServer(config.Server.ServerAddr, config.Server.ResultAddr); err != nil {
		panic(err)
	}
}
