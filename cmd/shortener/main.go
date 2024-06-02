package main

import "github.com/morozoffnor/go-url-shortener/internal/app"

func main() {
	if err := app.RunServer(); err != nil {
		panic(err)
	}
}
