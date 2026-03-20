package server

import (
	"context"
	"net/http"

	khttp "github.com/go-kratos/kratos/v2/transport/http"
)

func registerHealthEndpoints(srv *khttp.Server) {
	if srv == nil {
		return
	}
	r := srv.Route("/")
	r.GET("/health/ready", wrapOperation(operationHealthReady, func(ctx context.Context, httpCtx khttp.Context) error {
		return httpCtx.Result(http.StatusOK, map[string]any{"status": "ready"})
	}))
	r.GET("/health/live", wrapOperation(operationHealthLive, func(ctx context.Context, httpCtx khttp.Context) error {
		return httpCtx.Result(http.StatusOK, map[string]any{"status": "alive"})
	}))
	r.GET("/health/startup", wrapOperation(operationHealthStartup, func(ctx context.Context, httpCtx khttp.Context) error {
		return httpCtx.Result(http.StatusOK, map[string]any{"status": "started"})
	}))
}
