package middlewares

import (
	"github.com/morozoffnor/go-url-shortener/internal/types"
	"go.uber.org/zap"
	"net/http"
	"time"
)

var Logger = NewLogger()

func NewLogger() *zap.SugaredLogger {
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	sugar := logger.Sugar()
	return sugar
}

func Log(h http.Handler) http.Handler {
	log := func(w http.ResponseWriter, r *http.Request) {
		defer Logger.Sync()

		start := time.Now()

		respData := &types.ResponseData{}

		rwl := types.ResponseWriterWithLog{
			ResponseWriter: w,
			ResponseData:   respData,
		}
		h.ServeHTTP(&rwl, r)

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
	}

	return http.HandlerFunc(log)
}
