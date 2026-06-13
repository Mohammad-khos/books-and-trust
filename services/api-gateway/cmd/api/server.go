package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func (app *Application) Run(ctx context.Context) error {
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	srv := &http.Server{
		Addr:         app.Config.App.Addr,
		Handler:      app.Router,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	errChan := make(chan error, 1)

	go func() {
		app.Logger.Infow("HTTP server has started", "port", srv.Addr)
		errChan <- srv.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		app.Logger.Infow("shutdown signal received")
	case err := <-errChan:
		return err
	}
	// graceful shutdown context
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	app.Logger.Infow("starting graceful shutdown")

	if err := srv.Shutdown(ctx); err != nil {
		app.Logger.Warnw("could not shutdow server gracefully", "error", err)
		_ = srv.Close()
		return err
	}
	app.Logger.Infow("server stopped gracefully")
	return nil
}
