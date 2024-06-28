package storage

import (
	"encoding/json"
	"errors"
	"github.com/google/uuid"
	"github.com/morozoffnor/go-url-shortener/internal/config"
	"github.com/morozoffnor/go-url-shortener/internal/middlewares"
	"math/rand"
	"os"
	"sync"
)

type URLStorage struct {
	mu   *sync.Mutex
	List []*url
}

type url struct {
	UUID        string `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

var URLs = newURLStorage()

func newURLStorage() URLStorage {
	u := URLStorage{
		List: []*url{},
		mu:   &sync.Mutex{},
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
	for _, v := range URLs.List {
		if v.OriginalURL == full {
			return v.ShortURL, nil
		}
	}
	newURL := &url{
		UUID:        uuid.NewString(),
		ShortURL:    createRandomCharSeq(),
		OriginalURL: full,
	}
	URLs.mu.Lock()
	URLs.List = append(URLs.List, newURL)
	_ = URLs.SaveToFile(newURL)
	URLs.mu.Unlock()
	return newURL.ShortURL, nil
}

func (s *URLStorage) GetFullURL(shortURL string) (string, error) {
	if len(shortURL) < 1 {
		return "", errors.New("no short URL provided")
	}
	for _, v := range URLs.List {
		if v.ShortURL == shortURL {
			return v.OriginalURL, nil
		}
	}
	return "", errors.New("there is no such URL")
}

func (s *URLStorage) SaveToFile(URLToSave *url) error {
	file, err := os.OpenFile(config.Server.FileStoragePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		middlewares.Logger.Error("error opening file "+config.Server.FileStoragePath, err)
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
	file, err := os.OpenFile(config.Server.FileStoragePath, os.O_RDONLY, 0666)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		middlewares.Logger.Error("error opening file "+config.Server.FileStoragePath, err)
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

func createRandomCharSeq() string {
	letterRunes := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	chars := make([]rune, 10)
	for i := range 10 {
		chars[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(chars)
}
