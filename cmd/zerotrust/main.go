package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/ovinc/zerotrust/internal/config"
	"github.com/ovinc/zerotrust/internal/handler"
	"github.com/ovinc/zerotrust/internal/log"
	"github.com/ovinc/zerotrust/internal/otel"
	"github.com/ovinc/zerotrust/internal/store"
	"github.com/sirupsen/logrus"
)

func main() {
	// initialize logger with trace hook
	log.Init()

	// initialize opentelemetry
	defer otel.Shutdown(context.Background())

	// initialize store
	defer store.Close()

	// setup http routes
	mux := http.NewServeMux()
	mux.HandleFunc("/verify", handler.VerifyHandler)
	mux.HandleFunc("/forward-auth", handler.ForwardAuthHandler)
	mux.HandleFunc("/health", handler.HealthHandler)

	// create http server with timeouts
	cfg := config.Get()
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	server := &http.Server{
		Addr:         addr,
		Handler:      otel.Middleware(mux),
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	// start server in goroutine
	go func() {
		logrus.Infof("starting server on %s", addr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logrus.WithError(err).Fatal("server error")
		}
	}()

	// wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logrus.Info("shutting down server...")

	if err := server.Shutdown(context.Background()); err != nil {
		logrus.WithError(err).Error("server forced to shutdown")
	}

	logrus.Info("server stopped")
}
