package storage

import (
	"context"
	"github.com/morozoffnor/go-url-shortener/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"log"
	"os"
	"testing"
	"time"
)

func TestUrlStorage_addNewUrl(t *testing.T) {
	cfg := &config.Config{
		ServerAddr:      ":8080",
		ResultAddr:      "http://localhost:8080",
		FileStoragePath: "/tmp/test.json",
	}
	strg := NewMemoryStorage(cfg)
	tmpFile, err := os.CreateTemp(os.TempDir(), "dbtest*.json")
	require.Nil(t, err)
	defer tmpFile.Close()
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

			for _, full := range test.urls {
				ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				result, err := strg.AddNewURL(ctx, full)
				defer cancel()
				require.NoError(t, err)
				assert.IsType(t, "", result)
				log.Print(result)
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
			name: "Get url that does not exist",
			URLs: []*url{
				{UUID: "2", ShortURL: "Test", OriginalURL: ""},
			},
			shortURL: "Test",
			wantErr:  true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cfg := &config.Config{
				ServerAddr:      ":8080",
				ResultAddr:      "http://localhost:8080",
				FileStoragePath: "/tmp/test.json",
			}
			strg := NewMemoryStorage(cfg)
			tmpFile, err := os.CreateTemp(os.TempDir(), "dbtest*.json")
			require.Nil(t, err)
			defer tmpFile.Close()

			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			shortURL, _ := strg.AddNewURL(ctx, test.URLs[0].OriginalURL)
			full, _, err := strg.GetFullURL(ctx, shortURL)
			defer cancel()
			if !test.wantErr {
				require.Equal(t, "http://test.com", full)
			} else {
				assert.Error(t, err)
			}

		})
	}
}
