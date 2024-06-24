package types

import "net/http"

type (
	ResponseData struct {
		Status int
		Size   int
	}

	ResponseWriterWithLog struct {
		http.ResponseWriter
		ResponseData *ResponseData
	}
)

func (r *ResponseWriterWithLog) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.ResponseData.Size += size
	return size, err
}

func (r *ResponseWriterWithLog) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.ResponseData.Status = statusCode
}
