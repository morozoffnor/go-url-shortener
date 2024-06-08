package app

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

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

			for _, body := range test.body {
				rBody := bytes.NewBuffer([]byte(body))
				request := httptest.NewRequest(http.MethodPost, "/", rBody)

				w := httptest.NewRecorder()
				ShortURL(w, request)

				res := w.Result()

				assert.Equal(t, test.want.code, res.StatusCode)
				defer res.Body.Close()
				resBody, err := io.ReadAll(res.Body)
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
		shortUrl string
		want     want
	}{
		{
			name:     "Positive test (get full url)",
			shortUrl: "/TeSt",
			want: want{
				code:          http.StatusTemporaryRedirect,
				url:           "http://test.xyz/",
				checkLocation: true,
			},
		},
		{
			name:     "Negative test (url does not exist)",
			shortUrl: "/TeSt",
			want: want{
				code:          http.StatusBadRequest,
				url:           "http://test.xyz/",
				checkLocation: false,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			url, _ := urlStorage.addNewURL("http://test.xyz/")
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
