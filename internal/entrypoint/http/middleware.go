package http

import (
	"net/http"
)

// middlewareLogging log start and end of a http session.
func (s *ServerHTTP) middlewareLogging() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			s.log.Printf("Incomming http request:"+
				"r.URL.Path: %s r.Method: %s r.RemoteAddr: %s ",
				r.URL.Path, r.Method, r.RemoteAddr)

			next.ServeHTTP(w, r)
		})
	}
}

// middlewareCORS log start and end of http session.
func (s *ServerHTTP) middlewareCORS() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set(
				"Access-Control-Allow-Headers",
				`Accept, Content-Type, Content-Length, Accept-Encoding,
				X-CSRF-Token, Authorization, Access-Control-Request-Headers,
				Access-Control-Request-Method, Connection, Host, Origin,
				User-Agent, Referer, Cache-Control`)
			w.Header().Set(
				"Access-Control-Allow-Methods",
				"GET, POST, PUT, PATCH, DELETE, OPTIONS",
			)
			w.Header().Set("Content-Type", "text/plain; charset=UTF-8")

			next.ServeHTTP(w, r)
		})
	}
}
