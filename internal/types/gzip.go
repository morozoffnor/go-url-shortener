package types

import (
	"compress/gzip"
	"io"
	"net/http"
)

type GzipWriter struct {
	ResponseWriter http.ResponseWriter
	Writer         *gzip.Writer
}

func NewGzipWriter(w http.ResponseWriter) *GzipWriter {
	return &GzipWriter{ResponseWriter: w,
		Writer: gzip.NewWriter(w)}
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

type GzipReader struct {
	r      io.ReadCloser
	reader *gzip.Reader
}

func NewGzipReader(r io.ReadCloser) (*GzipReader, error) {
	zr, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}

	return &GzipReader{
		r:      r,
		reader: zr,
	}, nil
}

func (g GzipReader) Read(b []byte) (int, error) {
	return g.r.Read(b)
}

func (g GzipReader) Close() error {
	err := g.r.Close()
	if err != nil {
		return err
	}
	return err
}
