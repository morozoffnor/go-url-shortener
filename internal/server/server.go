package server

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/morozoffnor/go-url-shortener/internal/handlers"
	"log"
	"net/http"
)

func RunServer(addr string, respAddr string) error {
	handlers.ResponseAddr = respAddr
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Get("/{id}", handlers.FullURL)
	r.Post("/", handlers.ShortURL)

	log.Print("The server is listening on " + addr)
	return http.ListenAndServe(addr, r)
}
