package storage

import (
	"context"
	"github.com/morozoffnor/go-url-shortener/internal/config"
	"log"
)

type url struct {
	UUID        string `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

type Storage interface {
	AddNewURL(ctx context.Context, full string) (string, error)
	GetFullURL(ctx context.Context, shortURL string) (string, error)
	Ping(ctx context.Context) bool
}

func NewStorage(cfg *config.Config) Storage {
	if cfg.DatabaseDSN != "" {
		log.Print("Using database storage")
		return NewDatabase(cfg)
	}
	if cfg.FileStoragePath != "" {
		log.Print("Using file storage")
		return NewFileStorage(cfg)
	}
	log.Print("Using memory storage")
	return NewMemoryStorage(cfg)
}
