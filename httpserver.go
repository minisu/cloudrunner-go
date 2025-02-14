package cloudrunner

import (
	"context"
	"fmt"
	"net/http"

	"go.einride.tech/cloudrunner/cloudserver"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.uber.org/zap"
)

// HTTPMiddleware is an HTTP middleware.
type HTTPMiddleware = func(http.Handler) http.Handler

// NewHTTPServer creates a new HTTP server preconfigured with middleware for request logging, tracing, etc.
func NewHTTPServer(ctx context.Context, handler http.Handler, middlewares ...HTTPMiddleware) *http.Server {
	if handler == nil {
		panic("cloudrunner.NewHTTPServer: handler must not be nil")
	}
	run, ok := getRunContext(ctx)
	if !ok {
		panic("cloudrunner.NewHTTPServer: must be called with a context from cloudrunner.Run")
	}
	defaultMiddlewares := []cloudserver.HTTPMiddleware{
		func(handler http.Handler) http.Handler {
			return otelhttp.NewHandler(handler, "server")
		},
		run.loggerMiddleware.HTTPServer,
		run.traceMiddleware.HTTPServer,
		run.requestLoggerMiddleware.HTTPServer,
		run.serverMiddleware.HTTPServer,
	}
	return &http.Server{
		Addr: fmt.Sprintf(":%d", run.config.Runtime.Port),
		Handler: cloudserver.ChainHTTPMiddleware(
			handler,
			append(defaultMiddlewares, middlewares...)...,
		),
		ReadTimeout:       run.serverMiddleware.Config.Timeout,
		ReadHeaderTimeout: run.serverMiddleware.Config.Timeout,
		WriteTimeout:      run.serverMiddleware.Config.Timeout,
		IdleTimeout:       run.serverMiddleware.Config.Timeout,
	}
}

// ListenHTTP binds a listener on the configured port and listens for HTTP requests.
func ListenHTTP(ctx context.Context, httpServer *http.Server) error {
	go func() {
		<-ctx.Done()
		Logger(ctx).Info("HTTPServer shutting down")
		if err := httpServer.Shutdown(context.Background()); err != nil {
			Logger(ctx).Error("HTTPServer shutdown error", zap.Error(err))
		}
	}()
	Logger(ctx).Info("HTTPServer listening", zap.String("address", httpServer.Addr))
	if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}
