package main

import (
	"github.com/morozoffnor/go-url-shortener/internal/config"
	"github.com/morozoffnor/go-url-shortener/internal/server"
)

func main() {

	cfg := config.NewConfig()

	if err := server.RunServer(cfg); err != nil {
		panic(err)
	}
}
