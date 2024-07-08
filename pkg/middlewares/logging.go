package middlewares

import (
	"go.uber.org/zap"
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

var Logger = NewLogger()

func NewLogger() *zap.SugaredLogger {
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	sugar := logger.Sugar()
	return sugar
}

func Log(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer Logger.Sync()

		start := time.Now()

		respData := &ResponseData{}

		rwl := ResponseWriterWithLog{
			ResponseWriter: w,
			ResponseData:   respData,
		}
		next.ServeHTTP(&rwl, r)

		duration := time.Since(start)

		Logger.Infoln(
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
