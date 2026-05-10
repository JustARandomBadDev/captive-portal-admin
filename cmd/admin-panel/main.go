package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/JustARandomBadDev/captive-portal-admin/internal/app"
	"github.com/JustARandomBadDev/captive-portal-admin/internal/config"
)

func main() {
	cfg := config.Load()
	ctx := context.Background()

	application, err := app.New(ctx, cfg)
	if err != nil {
		log.Fatalf("init app: %v", err)
	}
	defer application.Close()

	log.Printf("admin panel listening on %s", cfg.AppAddr)

	errCh := make(chan error, 1)
	go func() {
		errCh <- application.Server.ListenAndServe()
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-errCh:
		log.Fatalf("server stopped: %v", err)
	case <-stop:
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := application.Server.Shutdown(ctx); err != nil {
			log.Fatalf("shutdown server: %v", err)
		}
	}
}
