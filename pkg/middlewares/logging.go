package middlewares

import (
	"github.com/morozoffnor/go-url-shortener/pkg/logger"
	_ "github.com/morozoffnor/go-url-shortener/pkg/logger"
	"net/http"
	"time"
)

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

func Log(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer logger.Logger.Sync()

		start := time.Now()

		respData := &ResponseData{}

		rwl := ResponseWriterWithLog{
			ResponseWriter: w,
			ResponseData:   respData,
		}
		next.ServeHTTP(&rwl, r)

		duration := time.Since(start)

		logger.Logger.Infoln(
			"uri", r.RequestURI,
			"method", r.Method,
			"status", respData.Status,
			"duration", duration,
			"size", respData.Size,
			"accept-encoding", r.Header.Get("Accept-Encoding"),
			"content-encoding", r.Header.Get("Content-Encoding"),
		)
	})

}
