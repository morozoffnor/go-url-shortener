package middlewares

import (
	"github.com/morozoffnor/go-url-shortener/pkg/gzip"
	"net/http"
	"strings"
)

func Compress(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		nw := w
		if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			gzipWriter := gzip.NewWriter(w)
			nw = gzipWriter
			defer gzipWriter.Close()
		}

		next.ServeHTTP(nw, r)

	}

}
