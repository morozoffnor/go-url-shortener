package middlewares

import (
	"github.com/morozoffnor/go-url-shortener/internal/types"
	"net/http"
	"strings"
)

func Compress(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		nw := w
		if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			gzipWriter := types.NewGzipWriter(w)
			nw = gzipWriter
			defer gzipWriter.Close()
		}

		h.ServeHTTP(nw, r)

	}

}
