package main

import (
	"github.com/morozoffnor/go-url-shortener/internal/config"
	"github.com/morozoffnor/go-url-shortener/internal/server"
)

func main() {
	if err := server.RunServer(config.Server.ServerAddr, config.Server.ResultAddr); err != nil {
		panic(err)
	}
}
