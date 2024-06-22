package storage

import (
	"errors"
	"math/rand"
)

type URLStorage struct {
	List map[string]string
}

var URLs = &URLStorage{
	List: make(map[string]string),
}

func (s *URLStorage) AddNewURL(full string) (string, error) {
	if f, ok := s.List[full]; ok {
		return f, nil
	}
	randChars := createRandomCharSeq()
	s.List[full] = randChars
	return randChars, nil
}

func (s *URLStorage) GetFullURL(shortURL string) (string, error) {
	for i, val := range s.List {
		if val == shortURL {
			return i, nil
		}
	}
	return "", errors.New("there is no such url")
}

func createRandomCharSeq() string {
	letterRunes := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	chars := make([]rune, 10)
	for i := range 10 {
		chars[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(chars)
}
