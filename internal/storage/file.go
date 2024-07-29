package storage

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/google/uuid"
	"github.com/morozoffnor/go-url-shortener/internal/config"
	"github.com/morozoffnor/go-url-shortener/pkg/chargen"
	"github.com/morozoffnor/go-url-shortener/pkg/logger"
	"os"
	"sync"
)

type FileStorage struct {
	mu   *sync.Mutex
	cfg  *config.Config
	List []*url
}

func NewFileStorage(cfg *config.Config) *FileStorage {
	u := &FileStorage{
		List: []*url{},
		mu:   &sync.Mutex{},
		cfg:  cfg,
	}
	err := u.LoadFromFile()
	if err != nil {
		logger.Logger.Error("error loading from file", err)
	}
	return u
}

func (s *FileStorage) AddNewURL(ctx context.Context, full string) (string, error) {
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
	_ = s.SaveToFile(newURL)
	s.mu.Unlock()
	return newURL.ShortURL, nil
}

func (s *FileStorage) GetFullURL(ctx context.Context, shortURL string) (string, error) {
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

func (s *FileStorage) SaveToFile(URLToSave *url) error {
	file, err := os.OpenFile(s.cfg.FileStoragePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		logger.Logger.Error("error opening file "+s.cfg.FileStoragePath, err)
		return err
	}
	data, err := json.MarshalIndent(&URLToSave, "", "    ")
	if err != nil {
		return err
	}
	_, err = file.Write(data)
	if err != nil {
		return err
	}
	return err
}

func (s *FileStorage) LoadFromFile() error {
	file, err := os.OpenFile(s.cfg.FileStoragePath, os.O_RDONLY, 0666)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		logger.Logger.Error("error opening file "+s.cfg.FileStoragePath, err)
		return err
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	for decoder.More() {
		var u url
		err := decoder.Decode(&u)
		if err != nil {
			return err
		}
		s.List = append(s.List, &u)
	}
	return nil
}

func (s *FileStorage) AddBatch(ctx context.Context, urls []BatchInput) ([]BatchOutput, error) {
	if len(urls) < 1 {
		return []BatchOutput{}, nil
	}
	var result []BatchOutput
	for _, v := range urls {
		shortURL, err := s.AddNewURL(ctx, v.OriginalURL)
		if err != nil {
			return nil, err
		}
		result = append(result, BatchOutput{
			ShortURL:      shortURL,
			CorrelationID: v.CorrelationID,
		})
	}
	return result, nil
}

func (s *FileStorage) GetUserURLs(ctx context.Context, userID uuid.UUID) ([]UserURLs, error) {
	if len(userID) == 0 {
		return nil, nil
	}
	var result []UserURLs
	for _, v := range s.List {
		if v.UserID == userID.String() {
			var u UserURLs
			u.ShortURL = s.cfg.ResultAddr + "/" + v.ShortURL
			u.OriginalURL = v.OriginalURL

			result = append(result, u)

		}
	}
	return result, nil
}
