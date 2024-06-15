package internal

import (
	"errors"
	"math/rand"
)

type URLStorage struct {
	list map[string]string
}

func (s *URLStorage) addNewURL(full string) (string, error) {
	if f, ok := s.list[full]; ok {
		return f, nil
	}
	randChars := createRandomCharSeq()
	s.list[full] = randChars
	return randChars, nil
}

func (s *URLStorage) getFullURL(shortURL string) (string, error) {
	for i, val := range s.list {
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
