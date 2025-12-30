package middleware

import (
	"bytes"
	"io"
	"log"
	"net/http"
)

type responseWriter struct {
	http.ResponseWriter
	statusCode int
	body       *bytes.Buffer
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	if rw.body != nil {
		rw.body.Write(b)
	}
	return rw.ResponseWriter.Write(b)
}

// Logging middleware logs HTTP requests
func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Read request body
		var bodyBytes []byte
		if r.Body != nil {
			bodyBytes, _ = io.ReadAll(r.Body)
			r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}

		// Create response writer wrapper
		rw := &responseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
			body:           nil, // Don't capture response body for logging
		}

		// Log request
		log.Printf("%s %s - Body: %s, Query: %s", r.Method, r.URL.Path, string(bodyBytes), r.URL.RawQuery)

		// Call next handler
		next.ServeHTTP(rw, r)

		// Log response
		log.Printf("%s %s - Status: %d", r.Method, r.URL.Path, rw.statusCode)
	})
}

