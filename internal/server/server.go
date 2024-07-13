package server

import (
	"github.com/go-chi/chi/v5"
	"github.com/morozoffnor/go-url-shortener/internal/config"
	"github.com/morozoffnor/go-url-shortener/internal/handlers"
	"github.com/morozoffnor/go-url-shortener/internal/storage"
	"github.com/morozoffnor/go-url-shortener/pkg/middlewares"
	"net/http"
)

func newRouter(cfg *config.Config, s *storage.URLStorage) *chi.Mux {
	r := chi.NewRouter()
	r.Use(middlewares.Log)
	r.Get("/{id}", handlers.NewFullURLHandler(cfg, s))
	r.Post("/", middlewares.Compress(handlers.NewShortURLHandler(cfg, s)))
	r.Post("/api/shorten", middlewares.Compress(handlers.NewShortenHandler(cfg, s)))
	return r
}

func New(cfg *config.Config) *http.Server {
	strg := storage.New(cfg)
	s := &http.Server{
		Addr:    cfg.ServerAddr,
		Handler: newRouter(cfg, strg),
	}
	return s
}

//func RunServer(cfg *config.Config) error {
//	s := New(cfg)
//	log.Print("The server is listening on " + cfg.ServerAddr)
//	return http.ListenAndServe(cfg.ServerAddr, r)
//}
