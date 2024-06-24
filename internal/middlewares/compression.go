package middlewares

import (
	"github.com/morozoffnor/go-url-shortener/internal/types"
	"net/http"
	"strings"
)

func Compress(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nw := w
		if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			gzipWriter := types.NewGzipWriter(w)
			nw = gzipWriter
			defer gzipWriter.Close()
		}

		if strings.Contains(r.Header.Get("Content-Encoding"), "gzip") {
			gzipReader, err := types.NewGzipReader(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			r.Body = gzipReader
			defer gzipReader.Close()
		}

		h.ServeHTTP(nw, r)

	})
}
