package storage

import (
	"context"
	"github.com/morozoffnor/go-url-shortener/internal/config"
	"log"
)

//go:generate mockgen -source=storage.go -destination=mock/storage.go -package=mock

type url struct {
	UUID        string `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

type BatchInput struct {
	OriginalURL   string `json:"original_url"`
	CorrelationID string `json:"correlation_id"`
}

type BatchOutput struct {
	ShortURL      string `json:"short_url"`
	CorrelationID string `json:"correlation_id"`
}

type Storage interface {
	AddNewURL(ctx context.Context, full string) (string, error)
	GetFullURL(ctx context.Context, shortURL string) (string, error)
	AddBatch(ctx context.Context, urls []BatchInput) ([]BatchOutput, error)
}

type Pingable interface {
	Ping(ctx context.Context) bool
}

func NewStorage(cfg *config.Config, ctx context.Context) Storage {
	if cfg.DatabaseDSN != "" {
		log.Print("Using database storage")
		return NewDatabase(cfg, ctx)
	}
	if cfg.FileStoragePath != "" {
		log.Print("Using file storage")
		return NewFileStorage(cfg)
	}
	log.Print("Using memory storage")
	return NewMemoryStorage(cfg)
}
