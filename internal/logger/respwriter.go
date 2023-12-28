package logger

import "net/http"

// responseData extra data for logging.
type responseData struct {
	status int
	size   int
}

// loggingResponseWriter http.ResponseWriter's deocrator.
type loggingResponseWriter struct {
	http.ResponseWriter
	responseData *responseData
}

// createResponseWriter create method.
func createResponseWriter(w http.ResponseWriter) *loggingResponseWriter {
	return &loggingResponseWriter{w, &responseData{200, 0}}
}

// getResponseData returns extra data for logging.
func (r *loggingResponseWriter) getResponseData() *responseData {
	return r.responseData
}

// Write decoration method to extract extra data from http.ResponseWriter.
func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size
	return size, err
}

// WriteHeader decoration method to extract extra data from http.ResponseWriter.
func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode
}
