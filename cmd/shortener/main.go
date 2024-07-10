package main

import (
	"flag"
	"github.com/morozoffnor/go-url-shortener/internal/config"
	"github.com/morozoffnor/go-url-shortener/internal/server"
)

func main() {
	flag.Parse()
	cfg := config.New()

	if err := server.RunServer(cfg); err != nil {
		panic(err)
	}
}
