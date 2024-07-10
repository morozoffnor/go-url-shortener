package server

import (
	"github.com/go-chi/chi/v5"
	"github.com/morozoffnor/go-url-shortener/internal/config"
	"github.com/morozoffnor/go-url-shortener/internal/handlers"
	"github.com/morozoffnor/go-url-shortener/internal/storage"
	"github.com/morozoffnor/go-url-shortener/pkg/middlewares"
	"log"
	"net/http"
)

func newRouter(cfg *config.ServerConfig, s *storage.URLStorage) *chi.Mux {
	r := chi.NewRouter()
	r.Use(middlewares.Log)
	r.Get("/{id}", handlers.NewFullURLHandler(cfg, s))
	r.Post("/", middlewares.Compress(handlers.NewShortURLHandler(cfg, s)))
	r.Post("/api/shorten", middlewares.Compress(handlers.NewShortenHandler(cfg, s)))
	return r
}

func RunServer(cfg *config.ServerConfig) error {
	s := storage.New(cfg)
	r := newRouter(cfg, s)

	log.Print("The server is listening on " + cfg.ServerAddr)
	return http.ListenAndServe(cfg.ServerAddr, r)
}
