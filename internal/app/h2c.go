package app

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

// StartH2CServer starts a cleartext HTTP/2 server with the same Gin routes as
// the main HTTP server.
func StartH2CServer(addr string, router func(*gin.Engine)) (*http.Server, error) {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("listen h2c server %s: %w", addr, err)
	}

	engine := gin.New()
	engine.Use(gin.LoggerWithConfig(gin.LoggerConfig{Output: io.Discard}))
	engine.Use(gin.Recovery())
	if router != nil {
		router(engine)
	}

	server := &http.Server{
		Addr:              listener.Addr().String(),
		Handler:           h2c.NewHandler(engine, &http2.Server{}),
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		slog.Info("h2c server listening", "addr", listener.Addr().String())
		if err := server.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("h2c server stopped", "addr", listener.Addr().String(), "err", err)
		}
	}()

	return server, nil
}
