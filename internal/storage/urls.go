package storage

import (
	"encoding/json"
	"errors"
	"github.com/google/uuid"
	"github.com/morozoffnor/go-url-shortener/internal/config"
	"github.com/morozoffnor/go-url-shortener/pkg/chargen"
	"github.com/morozoffnor/go-url-shortener/pkg/middlewares"
	"os"
	"sync"
)

type URLStorage struct {
	mu   *sync.Mutex
	cfg  *config.Config
	List []*url
}

type url struct {
	UUID        string `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

//var URLs = newURLStorage()

func NewURLStorage(cfg *config.Config) *URLStorage {
	u := &URLStorage{
		List: []*url{},
		mu:   &sync.Mutex{},
		cfg:  cfg,
	}
	err := u.LoadFromFile()
	if err != nil {
		middlewares.Logger.Error("error loading from file", err)
	}
	return u
}

func (s *URLStorage) AddNewURL(full string) (string, error) {
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

func (s *URLStorage) GetFullURL(shortURL string) (string, error) {
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

func (s *URLStorage) SaveToFile(URLToSave *url) error {
	file, err := os.OpenFile(s.cfg.FileStoragePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		middlewares.Logger.Error("error opening file "+s.cfg.FileStoragePath, err)
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

func (s *URLStorage) LoadFromFile() error {
	file, err := os.OpenFile(s.cfg.FileStoragePath, os.O_RDONLY, 0666)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		middlewares.Logger.Error("error opening file "+s.cfg.FileStoragePath, err)
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
