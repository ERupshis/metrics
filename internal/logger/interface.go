package logger

import "net/http"

type BaseLogger interface {
	Sync()

	Info(msg string, fields ...interface{})
	LogHandler(h http.Handler) http.Handler
}
