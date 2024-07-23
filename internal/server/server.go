package server

import (
	"github.com/go-chi/chi/v5"
	"github.com/morozoffnor/go-url-shortener/internal/config"
	"github.com/morozoffnor/go-url-shortener/internal/handlers"
	"github.com/morozoffnor/go-url-shortener/internal/storage"
	"github.com/morozoffnor/go-url-shortener/pkg/middlewares"
	"net/http"
)

func newRouter(h *handlers.Handlers) *chi.Mux {
	r := chi.NewRouter()
	r.Use(middlewares.Log)
	r.Get("/ping", h.PingHandler)
	r.Get("/{id}", h.FullURLHandler)
	r.Post("/", middlewares.Compress(h.ShortURLHandler))
	r.Post("/api/shorten/batch", middlewares.Compress(h.BatchHandler))
	r.Post("/api/shorten", middlewares.Compress(h.ShortenHandler))

	return r
}

func New(cfg *config.Config, strg storage.Storage, h *handlers.Handlers) *http.Server {
	s := &http.Server{
		Addr:    cfg.ServerAddr,
		Handler: newRouter(h),
	}
	return s
}
