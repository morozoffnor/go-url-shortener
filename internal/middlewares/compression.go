package middlewares

import (
	"compress/gzip"
	"github.com/morozoffnor/go-url-shortener/internal/types"
	"net/http"
	"strings"
)

func Compress(h http.Handler) http.Handler {
	compress := func(w http.ResponseWriter, r *http.Request) {

		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			h.ServeHTTP(w, r)
			return
		}
		gz, err := gzip.NewWriterLevel(w, gzip.BestSpeed)
		if err != nil {
			Logger.Error(w, err.Error())
			return
		}
		defer gz.Close()
		w.Header().Set("Content-Encoding", "gzip")
		h.ServeHTTP(types.GzipWriter{ResponseWriter: w, Writer: gz}, r)
	}

	return http.HandlerFunc(compress)
}
