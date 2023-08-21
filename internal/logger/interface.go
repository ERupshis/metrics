package logger

import "net/http"

type BaseLogger interface {
	Sync()

	Info(msg string, fields ...interface{})
	LogHandler(h http.Handler) http.Handler
}

func CreateLogger(level string) BaseLogger {
	log, err := CreateZapLogger(level)
	if err != nil {
		panic(err)
	}

	return log
}
