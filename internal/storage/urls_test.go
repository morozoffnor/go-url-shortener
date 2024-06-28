package storage

import (
	"github.com/morozoffnor/go-url-shortener/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

func init() {
	tmpFile, _ := os.CreateTemp(os.TempDir(), "dbtest*.json")
	tmpFile.Close()
	config.Server.FileStoragePath = tmpFile.Name()
}

func TestUrlStorage_addNewUrl(t *testing.T) {
	tests := []struct {
		name string
		list map[string]string
		urls []string
	}{
		{
			name: "Add new url",
			list: make(map[string]string),
			urls: []string{
				"http://test.com",
			},
		},
		{
			name: "Add the same url twice",
			list: make(map[string]string),
			urls: []string{
				"http://test.com", "http://test.com",
			},
		},
	}
	var lastResult string
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			storage := newURLStorage()
			for _, full := range test.urls {
				result, err := storage.AddNewURL(full)
				require.NoError(t, err)
				assert.IsType(t, "", result)
				if len(lastResult) > 0 {
					assert.Equal(t, result, lastResult)
				}
			}
		})
	}
}

func TestUrlStorage_getFullUrl(t *testing.T) {
	tests := []struct {
		name     string
		URLs     []*url
		shortURL string
		wantErr  bool
	}{
		{
			name: "Get full url",
			URLs: []*url{
				{UUID: "1", ShortURL: "Test", OriginalURL: "http://test.com"},
			},
			shortURL: "Test",
			wantErr:  false,
		},
		{
			name:     "Get url that does not exist",
			URLs:     []*url{},
			shortURL: "Test",
			wantErr:  true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			URLs = newURLStorage()

			URLs.mu.Lock()
			URLs.List = append(URLs.List, test.URLs...)
			URLs.mu.Unlock()
			full, err := URLs.GetFullURL(test.shortURL)
			if !test.wantErr {
				require.Equal(t, "http://test.com", full)
			} else {
				assert.Error(t, err)
			}

		})
	}
}
