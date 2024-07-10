package main

import (
	"github.com/morozoffnor/go-url-shortener/internal/config"
	"github.com/morozoffnor/go-url-shortener/internal/server"
)

func main() {

	cfg := config.New()

	if err := server.RunServer(cfg); err != nil {
		panic(err)
	}
}
