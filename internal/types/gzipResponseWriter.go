package types

import (
	"io"
	"net/http"
)

type GzipWriter struct {
	http.ResponseWriter
	Writer io.Writer
}

func (w GzipWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

func (w GzipWriter) Header() http.Header {
	return w.Header()
}

func (w GzipWriter) WriteHeader(statusCode int) {
	w.WriteHeader(statusCode)
}

func (w GzipWriter) Close() error {
	return w.Close()
}
