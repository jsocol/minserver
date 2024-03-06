package handlers

import (
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"strings"
)

type staticOption func(*Static)

func WithPaths(prefix, basePath string) staticOption {
	return func(s *Static) {
		s.prefix = prefix
		s.basePath = basePath
	}
}

func WithPrefix(prefix string) staticOption {
	return func(s *Static) {
		s.prefix = prefix
	}
}

func WithBasePath(basePath string) staticOption {
	return func(s *Static) {
		s.basePath = basePath
	}
}

func WithFS(fsys fs.FS) staticOption {
	return func(s *Static) {
		s.fsys = fsys
	}
}

type Static struct {
	prefix   string
	basePath string
	fsys     fs.FS
}

func NewStatic(opts ...staticOption) http.Handler {
	s := &Static{}
	for _, o := range opts {
		o(s)
	}

	// default to the current working directory
	if s.basePath == "" {
		cwd, err := os.Getwd()
		if err != nil {
			panic(err)
		}
		s.basePath = cwd
	}

	// default to "/"
	if s.prefix == "" {
		s.prefix = "/"
	}

	// set the default FS to the basePath
	if s.fsys == nil {
		s.fsys = os.DirFS(s.basePath)
	}

	return s
}

func (s *Static) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodOptions {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	urlPath := r.URL.Path
	if strings.Contains(urlPath, "..") {
		w.WriteHeader(http.StatusBadRequest)
		attrs := []slog.Attr{
			slog.String("path", urlPath),
		}
		slog.LogAttrs(r.Context(), slog.LevelWarn, "minserver: static file request with ..", attrs...)
		return
	}
	relativePath := strings.TrimPrefix(urlPath, s.prefix)

	http.ServeFileFS(w, r, s.fsys, relativePath)
}
