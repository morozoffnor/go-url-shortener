package body

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
)

func GetBody(r *http.Request) ([]byte, error) {
	if strings.Contains(r.Header.Get("Content-Encoding"), "gzip") {
		return compressedBody(r)
	}
	return uncompressedBody(r)
}

func compressedBody(r *http.Request) ([]byte, error) {
	zip, err := gzip.NewReader(r.Body)
	if err != nil {
		return []byte{}, nil
	}
	defer zip.Close()

	body, err := io.ReadAll(zip)
	if err != nil {
		return nil, err
	}
	return body, nil
}

func uncompressedBody(r *http.Request) ([]byte, error) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()

	return body, err
}
