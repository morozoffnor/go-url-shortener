package logger

import "go.uber.org/zap"

var Logger = NewLogger()

func NewLogger() *zap.SugaredLogger {
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	sugar := logger.Sugar()
	return sugar
}
