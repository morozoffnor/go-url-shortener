package main

import (
	config "github.com/morozoffnor/go-url-shortener/config"
	"github.com/morozoffnor/go-url-shortener/internal/app"
)

func main() {
	if err := app.RunServer(config.Server.ServerAddr, config.Server.ResultAddr); err != nil {
		panic(err)
	}
}
