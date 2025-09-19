package middleware

import (
	"log"
	"net/http"
)

type spyWriter struct {
	status int
	http.ResponseWriter
}

// WriteHeader implements http.ResponseWriter.
func (s *spyWriter) WriteHeader(statusCode int) {
	s.status = statusCode
	s.ResponseWriter.WriteHeader(statusCode)
}

func RequestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		spy := &spyWriter{ResponseWriter: w}
		log.Printf("INFO: processing %v %v\n", r.Method, r.URL)
		next.ServeHTTP(spy, r)
		log.Printf("INFO: answered with %v\n", spy.status)
	})
}
