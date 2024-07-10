package gzip

import (
	"compress/gzip"
	"io"
	"net/http"
)

type Writer struct {
	ResponseWriter http.ResponseWriter
	Writer         *gzip.Writer
}

func NewWriter(w http.ResponseWriter) *Writer {
	return &Writer{ResponseWriter: w,
		Writer: gzip.NewWriter(w)}
}

func (w *Writer) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

func (w *Writer) Header() http.Header {
	return w.ResponseWriter.Header()
}

func (w *Writer) WriteHeader(statusCode int) {
	w.ResponseWriter.Header().Set("Content-Encoding", "gzip")
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *Writer) Close() error {
	return w.Writer.Close()
}

type Reader struct {
	r      io.ReadCloser
	reader *gzip.Reader
}

func NewReader(r io.ReadCloser) (*Reader, error) {
	zr, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}

	return &Reader{
		r:      r,
		reader: zr,
	}, nil
}

func (g *Reader) Read(b []byte) (int, error) {
	return g.r.Read(b)
}

func (g *Reader) Close() error {
	err := g.r.Close()
	if err != nil {
		return err
	}
	return err
}
