package middlewares

import (
	"github.com/morozoffnor/go-url-shortener/internal/types"
	"net/http"
	"strings"
)

func Compress(h http.Handler) http.Handler {
	//compress := func(w http.ResponseWriter, r *http.Request) {
	//	if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
	//		h.ServeHTTP(w, r)
	//		return
	//	}
	//	//gz, err := gzip.NewWriterLevel(w, gzip.BestSpeed)
	//	//if err != nil {
	//	//	Logger.Error(w, err.Error())
	//	//	return
	//	//}
	//	//defer gz.Close()
	//	w.Header().Set("Content-Encoding", "gzip")
	//	h.ServeHTTP(types.NewGzipWriter(w), r)
	//}
	//
	//return http.HandlerFunc(compress)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			gzipWriter := types.NewGzipWriter(w)
			w = gzipWriter
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

		h.ServeHTTP(w, r)

	})
}
