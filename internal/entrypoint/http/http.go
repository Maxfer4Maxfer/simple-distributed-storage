package http

import (
	"context"
	"log"
	"net/http"
	"simple-storage/internal/utils"
	"time"
)

type middlewareFunc func(http.Handler) http.Handler

// Middleware is a representation of middleware.
type Middleware []middlewareFunc

// Serve wraps a provided Handler with middleware.
func (mw Middleware) Serve(next http.Handler) http.Handler {
	chain := next

	for i := len(mw) - 1; i >= 0; i-- {
		chain = mw[i](chain)
	}

	return chain
}

// MiddlewareSwitch declareas witch middleware will be available.
type MiddlewareSwitch struct {
	CorrelationID bool
	Logging       bool
	Prometheus    bool
}

// Config declares configuration for the HTTP server.
type Config struct {
	Address          string
	MiddlewareSwitch MiddlewareSwitch

	// https://golang.org/pkg/net/http/#Server
	// https://medium.com/@simonfrey/go-as-in-golang-standard-net-http-config-will-break-your-production-environment-1360871cb72b
	ReadTimeout       time.Duration
	ReadHeaderTimeout time.Duration
	WriteTimeout      time.Duration
	IdleTimeout       time.Duration
	MaxHeaderBytes    int
}

// ServerHTTP is a wraper on http.Server.
type ServerHTTP struct {
	log    *log.Logger
	config Config
	server *http.Server
}

// New returns a HTTP server.
func New(
	log *log.Logger,
	config Config,
	router http.Handler,
) *ServerHTTP {
	log = utils.LoggerExtendWithPrefix(log, "http-server ->")

	s := &ServerHTTP{
		log:    log,
		config: config,
		server: &http.Server{
			Addr:              config.Address,
			ReadTimeout:       config.ReadTimeout,
			ReadHeaderTimeout: config.ReadHeaderTimeout,
			WriteTimeout:      config.WriteTimeout,
			IdleTimeout:       config.IdleTimeout,
			MaxHeaderBytes:    config.MaxHeaderBytes,
		},
	}

	middleware := Middleware{
		s.middlewareCORS(),
		s.middlewareLogging(),
	}

	s.server.Handler = middleware.Serve(router)

	return s
}

// Start starts HTTP Server.
func (s *ServerHTTP) Start() chan error {
	serverErrors := make(chan error, 1)

	go func() {
		s.log.Printf("start HTTP API Listening %s", s.config.Address)
		serverErrors <- s.server.ListenAndServe()
	}()

	return serverErrors
}

// Shutdown stops HTTP Server.
func (s *ServerHTTP) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}
