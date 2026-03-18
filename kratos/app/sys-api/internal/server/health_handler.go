package server

import (
	"net/http"

	khttp "github.com/go-kratos/kratos/v2/transport/http"
)

func registerHealthEndpoints(srv *khttp.Server) {
	if srv == nil {
		return
	}
	r := srv.Route("/")
	r.GET("/health/ready", func(ctx khttp.Context) error {
		return ctx.Result(http.StatusOK, map[string]any{"status": "ready"})
	})
	r.GET("/health/live", func(ctx khttp.Context) error {
		return ctx.Result(http.StatusOK, map[string]any{"status": "alive"})
	})
	r.GET("/health/startup", func(ctx khttp.Context) error {
		return ctx.Result(http.StatusOK, map[string]any{"status": "started"})
	})
}
