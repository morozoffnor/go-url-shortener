package app

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

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
			s := &URLStorage{
				list: test.list,
			}
			for _, full := range test.urls {
				result, _ := s.addNewURL(full)
				assert.IsType(t, "", result)
				if len(lastResult) > 0 {
					assert.Equal(t, result, lastResult)
				}
			}
		})
	}
}

func TestUrlStorage_getFullUrl(t *testing.T) {
	type url struct {
		short string
		full  string
	}
	tests := []struct {
		name    string
		list    map[string]string
		urls    []url
		wantErr bool
	}{
		{
			name: "Get full url",
			list: map[string]string{
				"http://test.com": "GhydF",
			},
			urls: []url{
				{short: "GhydF", full: "http://test.com"},
			},
			wantErr: false,
		},
		{
			name: "Get url that does not exist",
			list: map[string]string{},
			urls: []url{
				{short: "Test", full: "http://test.com"},
			},
			wantErr: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			storage := &URLStorage{
				list: test.list,
			}
			for _, url := range test.urls {
				full, err := storage.getFullURL(url.short)
				if !test.wantErr {
					require.Equal(t, url.full, full)
				} else {
					assert.Error(t, err)
				}
			}
		})
	}
}
