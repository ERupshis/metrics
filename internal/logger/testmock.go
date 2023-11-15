package logger

import (
	"net/http"
)

type logMock struct {
}

func CreateMock() BaseLogger {
	return &logMock{}
}

func (t *logMock) Info(_ string, _ ...interface{}) {
}

func (t *logMock) Printf(_ string, _ ...interface{}) {
}

func (t *logMock) Sync() {
}

func (t *logMock) LogHandler(h http.Handler) http.Handler {
	return h
}
