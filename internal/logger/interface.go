package logger

import "net/http"

// BaseLogger used logger interface definition.
type BaseLogger interface {
	// Sync Method for flushing data in stream.
	Sync()

	// Info posts message on log 'info' level.
	Info(msg string, fields ...interface{})
	// LogHandler implements middleware for logging requests.
	LogHandler(h http.Handler) http.Handler
}

// CreateLogger logger implementation should be replaced here.
func CreateLogger(level string) BaseLogger {
	log, err := CreateZapLogger(level)
	if err != nil {
		panic(err)
	}

	return log
}
