package middleware

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
)

type loggingWriter struct {
	http.ResponseWriter
	req    *http.Request
	status int
}

func (lw *loggingWriter) WriteHeader(status int) {
	lw.status = status
	lw.ResponseWriter.WriteHeader(lw.status)
}

func (lw *loggingWriter) Write(data []byte) (int, error) {
	if lw.status == 0 {
		lw.status = http.StatusOK
	}
	return lw.ResponseWriter.Write(data)
}

func (lw *loggingWriter) log() {
	<-lw.req.Context().Done()
	ctxErr := lw.req.Context().Err()

	level := slog.LevelInfo
	if lw.status >= http.StatusBadRequest || lw.status == 0 || ctxErr != context.Canceled {
		level = slog.LevelError
	}

	msg := fmt.Sprintf("[%d] %s %s", lw.status, lw.req.Method, lw.req.URL.Path)
	attrs := []slog.Attr{
		slog.Int("status", lw.status),
		slog.String("path", lw.req.URL.Path),
		slog.String("method", lw.req.Method),
	}
	if ctxErr != context.Canceled {
		attrs = append(attrs, slog.Any("ctxError", ctxErr))
	}
	slog.LogAttrs(lw.req.Context(), level, msg, attrs...)
}

// Logging logs all requests using [log/slog]. If the request times
// out, the status will be set to 0.
func Logging(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		lw := &loggingWriter{
			ResponseWriter: w,
			req:            r,
		}
		go lw.log()
		w = lw
		next(w, r)
	}
}
