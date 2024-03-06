package middleware_test

import (
	"time"

	"github.com/jsocol/minserver"
	"github.com/jsocol/minserver/middleware"
)

func ExampleNewDeadline() {
	srv := minserver.New()
	srv.AddMiddleware(middleware.NewDeadline(3 * time.Second))
}
