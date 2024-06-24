package types

import (
	"compress/gzip"
	"net/http"
)

type GzipWriter struct {
	ResponseWriter http.ResponseWriter
	Writer         *gzip.Writer
}

func (w GzipWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

func (w GzipWriter) Header() http.Header {
	return w.ResponseWriter.Header()
}

func (w GzipWriter) WriteHeader(statusCode int) {
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w GzipWriter) Close() error {
	return w.Writer.Close()
}

func NewGzipWriter(w http.ResponseWriter) *GzipWriter {
	return &GzipWriter{ResponseWriter: w,
		Writer: gzip.NewWriter(w)}
}
