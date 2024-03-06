package minserver

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"sync/atomic"
)

var (
	ErrNotStarted     = errors.New("not started")
	ErrNotRunning     = errors.New("not running")
	ErrAlreadyStarted = errors.New("already running")
)

// Middleware functions are used to wrap or modify an incoming HTTP request or
// outgoing response. See [middleware.Logging] for an example.
type Middleware func(next http.HandlerFunc) http.HandlerFunc

// Options configure the server to use specific settings.
type Option func(*Server)

// WithAddr sets an address for ListenAndServe. The default is ":8000".
func WithAddr(addr string) Option {
	return func(s *Server) {
		s.bindAddr = addr
	}
}

type Server struct {
	mux        *http.ServeMux
	srv        *http.Server
	bindAddr   string
	middleware []Middleware
	running    atomic.Bool
}

func New(opts ...Option) *Server {
	s := &Server{
		mux:      http.NewServeMux(),
		bindAddr: ":8000",
	}

	for _, o := range opts {
		o(s)
	}

	return s
}

func (s *Server) AddMiddleware(mw Middleware) {
	s.middleware = append(s.middleware, mw)
}

func (s *Server) AddRoute(pattern string, handler http.Handler) {
	s.mux.Handle(pattern, handler)
}

func (s *Server) AddRouteFunc(pattern string, handler http.HandlerFunc) {
	s.AddRoute(pattern, http.HandlerFunc(handler))
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	next := s.mux.ServeHTTP
	for _, mw := range s.middleware {
		next = mw(next)
	}
	next(w, r)
}

func (s *Server) Start() error {
	if s.running.Load() {
		return ErrAlreadyStarted
	}
	s.running.Store(true)
	s.srv = &http.Server{
		Handler: s,
		Addr:    s.bindAddr,
	}
	slog.LogAttrs(context.Background(), slog.LevelInfo, fmt.Sprintf("minserver: starting on %s", s.bindAddr), slog.String("addr", s.bindAddr))
	return s.srv.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	if s.srv == nil {
		return ErrNotStarted
	}
	if !s.running.Load() {
		return ErrNotRunning
	}
	s.running.Store(false)
	slog.LogAttrs(ctx, slog.LevelInfo, "minserver: shutting down")
	return s.srv.Shutdown(ctx)
}
