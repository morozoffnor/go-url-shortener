package main

import (
	"github.com/morozoffnor/go-url-shortener/config"
	"github.com/morozoffnor/go-url-shortener/internal"
)

func main() {
	if err := internal.RunServer(config.Server.ServerAddr, config.Server.ResultAddr); err != nil {
		panic(err)
	}
}
