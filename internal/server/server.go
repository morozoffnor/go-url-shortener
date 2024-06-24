package server

import (
	"github.com/go-chi/chi/v5"
	"github.com/morozoffnor/go-url-shortener/internal/handlers"
	"github.com/morozoffnor/go-url-shortener/internal/middlewares"
	"log"
	"net/http"
)

func RunServer(addr string, respAddr string) error {
	handlers.ResponseAddr = respAddr
	r := chi.NewRouter()
	r.Use(middlewares.Compress)
	r.Use(middlewares.Log)

	r.Get("/{id}", handlers.FullURL)
	r.Post("/", handlers.ShortURL)
	r.Post("/api/shorten", handlers.Shorten)

	log.Print("The server is listening on " + addr)
	return http.ListenAndServe(addr, r)
}
