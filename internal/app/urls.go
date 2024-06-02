package app

import (
	"errors"
	"math/rand"
)

type UrlStorage struct {
	list map[string]string
}

func (s *UrlStorage) addNewUrl(full string) (string, error) {
	if f, ok := s.list[full]; ok {
		return f, nil
	}
	randChars := createRandomCharSeq()
	s.list[full] = randChars
	return randChars, nil
}

func (s *UrlStorage) getFullUrl(shortUrl string) (string, error) {
	value, ok := s.list[shortUrl]
	if !ok {
		return value, errors.New("there is no such url")
	}
	return value, nil
}

func createRandomCharSeq() string {
	letterRunes := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	chars := make([]rune, 10)
	for i := range 10 {
		chars[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(chars)
}
