package handlers

import (
	"bytes"
	"encoding/json"
	"github.com/morozoffnor/go-url-shortener/internal/config"
	"github.com/morozoffnor/go-url-shortener/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func init() {

	config.Server.FileStoragePath = ""
}

func TestShortURL(t *testing.T) {
	type want struct {
		code        int
		response    string
		contentType string
	}
	tests := []struct {
		name string
		body []string
		want want
	}{
		{
			name: "Positive test #1 (add new short url)",
			body: []string{"http://test.com/"},
			want: want{
				code:        http.StatusCreated,
				response:    "",
				contentType: "text/plain, utf-8",
			},
		},
		{
			name: "Positive test #2 (post the same full url twice)",
			body: []string{"http://test.com/", "http://test.com/"},
			want: want{
				code:        http.StatusCreated,
				response:    "",
				contentType: "text/plain, utf-8",
			},
		},
	}
	var lastRes []byte
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tmpFile, ferr := os.CreateTemp(os.TempDir(), "dbtest*.json")
			require.Nil(t, ferr)
			tmpFile.Close()
			config.Server.FileStoragePath = tmpFile.Name()
			for _, body := range test.body {
				rBody := bytes.NewBuffer([]byte(body))
				request := httptest.NewRequest(http.MethodPost, "/", rBody)

				w := httptest.NewRecorder()
				ShortURL(w, request)

				res := w.Result()

				assert.Equal(t, test.want.code, res.StatusCode)
				defer res.Body.Close()
				resBody, err := io.ReadAll(res.Body)
				log.Print("test res body ", resBody)
				require.NoError(t, err)
				assert.NotEmpty(t, resBody)
				if lastRes != nil {
					assert.Equal(t, lastRes, resBody)
				}
				lastRes = resBody
				assert.Equal(t, test.want.contentType, res.Header.Get("Content-Type"))
			}

		})
	}
}

func TestFullUrl(t *testing.T) {
	type want struct {
		code          int
		url           string
		checkLocation bool
	}
	tests := []struct {
		name     string
		shortURL string
		want     want
	}{
		{
			name:     "Positive test (get full url)",
			shortURL: "/TeSt",
			want: want{
				code:          http.StatusTemporaryRedirect,
				url:           "http://test.xyz/",
				checkLocation: true,
			},
		},
		{
			name:     "Negative test (url does not exist)",
			shortURL: "/TeSt",
			want: want{
				code:          http.StatusBadRequest,
				url:           "http://test.xyz/",
				checkLocation: false,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tmpFile, err := os.CreateTemp(os.TempDir(), "dbtest*.json")
			require.Nil(t, err)
			tmpFile.Close()
			config.Server.FileStoragePath = tmpFile.Name()
			url, _ := storage.URLs.AddNewURL("http://test.xyz/")
			if !test.want.checkLocation {
				url = "DoNotCare"
			}
			request := httptest.NewRequest(http.MethodGet, "/"+url, nil)
			request.SetPathValue("id", url)
			w := httptest.NewRecorder()
			FullURL(w, request)

			res := w.Result()
			defer res.Body.Close()
			assert.Equal(t, test.want.code, res.StatusCode)
			if test.want.checkLocation {
				assert.Equal(t, test.want.url, res.Header.Get("Location"))
			}

		})
	}
}

func TestShorten(t *testing.T) {
	type reqBody struct {
		URL string `json:"url"`
	}
	type resBody struct {
		Result string `json:"result"`
	}
	type want struct {
		code        int
		response    resBody
		contentType string
	}

	tests := []struct {
		name string
		body []reqBody
		want want
	}{
		{
			name: "Test json request",
			body: []reqBody{{URL: "http://test.com/"}},
			want: want{
				code:        http.StatusCreated,
				response:    resBody{},
				contentType: "application/json",
			},
		},
		{
			name: "Positive test #2 (post the same full url twice)",
			body: []reqBody{{URL: "http://test.com/"}, {URL: "http://test.com/"}},
			want: want{
				code:        http.StatusCreated,
				response:    resBody{},
				contentType: "application/json",
			},
		},
	}

	var lastRes string
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tmpFile, ferr := os.CreateTemp(os.TempDir(), "dbtest*.json")
			require.Nil(t, ferr)
			tmpFile.Close()
			config.Server.FileStoragePath = tmpFile.Name()
			for _, body := range test.body {
				jsonReqBody, err := json.Marshal(body)
				require.NoError(t, err)
				request := httptest.NewRequest(http.MethodPost, "/", bytes.NewBuffer(jsonReqBody))

				w := httptest.NewRecorder()
				Shorten(w, request)

				res := w.Result()

				assert.Equal(t, test.want.code, res.StatusCode)
				defer res.Body.Close()
				var rBody resBody
				var buf bytes.Buffer
				_, err = buf.ReadFrom(res.Body)
				require.NoError(t, err)
				err = json.Unmarshal(buf.Bytes(), &rBody)
				require.NoError(t, err)
				assert.NotEmpty(t, rBody)
				if lastRes != "" {
					assert.Equal(t, lastRes, rBody.Result)
				}
				lastRes = rBody.Result
				assert.Equal(t, test.want.contentType, res.Header.Get("Content-Type"))
			}
		})
	}
}
