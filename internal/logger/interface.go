package logger

import "net/http"

type BaseLogger interface {
	Create(level string) (*InfoLogger, error)
	Sync()

	RequestLogger(h http.HandlerFunc) http.Handler
}
