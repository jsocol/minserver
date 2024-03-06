package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/jsocol/minserver"
)

// NewDeadline returns a middleware function that applies a default
// timeout to incoming requests that don't have one set.
func NewDeadline(defaultTimeout time.Duration) minserver.Middleware {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			if _, ok := ctx.Deadline(); !ok {
				var cancel context.CancelFunc
				ctx, cancel = context.WithTimeout(ctx, defaultTimeout)
				defer cancel()

				r = r.WithContext(ctx)
			}
			next(w, r)
		}
	}
}
