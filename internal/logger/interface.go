package logger

import "net/http"

type BaseLogger interface {
	Sync()

	Log(h http.HandlerFunc) http.HandlerFunc
}
