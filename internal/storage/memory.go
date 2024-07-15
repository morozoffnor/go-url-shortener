package storage

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"github.com/morozoffnor/go-url-shortener/internal/config"
	"github.com/morozoffnor/go-url-shortener/pkg/chargen"
	"sync"
)

type MemoryStorage struct {
	mu   *sync.Mutex
	cfg  *config.Config
	List []*url
}

func NewMemoryStorage(cfg *config.Config) *MemoryStorage {
	u := &MemoryStorage{
		List: []*url{},
		mu:   &sync.Mutex{},
		cfg:  cfg,
	}
	return u
}

func (s *MemoryStorage) AddNewURL(ctx context.Context, full string) (string, error) {
	if len(full) < 1 {
		return "", errors.New("blank URL")
	}
	for _, v := range s.List {
		if v.OriginalURL == full {
			return v.ShortURL, nil
		}
	}
	newURL := &url{
		UUID:        uuid.NewString(),
		ShortURL:    chargen.CreateRandomCharSeq(),
		OriginalURL: full,
	}
	s.mu.Lock()
	s.List = append(s.List, newURL)
	s.mu.Unlock()
	return newURL.ShortURL, nil
}

func (s *MemoryStorage) GetFullURL(ctx context.Context, shortURL string) (string, error) {
	if len(shortURL) < 1 {
		return "", errors.New("no short URL provided")
	}
	for _, v := range s.List {
		if v.ShortURL == shortURL {
			return v.OriginalURL, nil
		}
	}
	return "", errors.New("there is no such URL")
}

func (s *MemoryStorage) Ping(ctx context.Context) bool {
	return true
}
