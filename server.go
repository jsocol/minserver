package minserver

import (
	"context"
	"errors"
	"net/http"
	"sync/atomic"
)

var (
	ErrNotStarted     = errors.New("not started")
	ErrNotRunning     = errors.New("not running")
	ErrAlreadyStarted = errors.New("already running")
)

type Middleware func(next http.HandlerFunc) http.HandlerFunc

type Server struct {
	mux        *http.ServeMux
	srv        *http.Server
	middleware []Middleware
	running    atomic.Bool
}

func New() *Server {
	return &Server{
		mux: http.NewServeMux(),
	}
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
		Addr:    ":8000",
	}
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
	return s.srv.Shutdown(ctx)
}
