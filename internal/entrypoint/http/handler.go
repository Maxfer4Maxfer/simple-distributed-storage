package http

import (
	"encoding/json"
	"log"
	"net/http"
)

// Handler is a wraper on http.Server.
type Handler struct {
	log *log.Logger
}

// NewHandler returns a common handler.
func NewHandler(log *log.Logger) *Handler {
	return &Handler{
		log: log,
	}
}

// HandleEmpty defines empty handler.
func (han *Handler) HandleEmpty() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})
}

// HandleOK defines handler for OK.
func (han *Handler) HandleOK() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
}

// HandleCORS defines cors.
func (han *Handler) HandleCORS(next http.Handler) http.HandlerFunc {
	origins := map[string]struct{}{}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")

		if _, ok := origins[r.Header.Get("origin")]; ok {
			w.Header().Set("Access-Control-Allow-Origin", r.Header.Get("origin"))
		}

		w.Header().Set("Access-Control-Allow-Headers",
			`X-Requested-With, X-SupplierId, X-UserId, X-Debug-Mode,
			X-Debug-Supplier-ID, Accept, Content-Type, Content-Length,
			Accept-Encoding, X-CSRF-Token, Authorization,
			Access-Control-Request-Headers, Access-Control-Request-Method,
			Connection, Host, Origin, User-Agent, Referer, Cache-Control,
			X-header, Wb-AppType, Wb-AppVersion`,
		)
		w.Header().Set(
			"Access-Control-Allow-Methods",
			"GET, POST, PUT, PATCH, DELETE, OPTIONS",
		)
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Content-Type", "text/plain; charset=UTF-8")

		next.ServeHTTP(w, r)
	})
}

// ResponseWithError helps to form the right response in case of error.
func (han *Handler) ResponseWithError(
	w http.ResponseWriter, r *http.Request, err error, statusCode int,
) {
	res := struct {
		Error string `json:"error"`
	}{}

	res.Error = err.Error()

	w.WriteHeader(statusCode)

	w.Header().Set("Content-Type", "application/json")

	err = json.NewEncoder(w).Encode(&res)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (han *Handler) ResponseWithData(
	w http.ResponseWriter, r *http.Request,
	buf []byte,
) {
	w.Header().Set("Content-Type", "application/octet-stream")
	w.WriteHeader(http.StatusOK)

	w.Write(buf)
}
