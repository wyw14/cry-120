package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	now := time.Now().UTC()
	app, err := assemble(os.Getenv("LAUNCHGUARD_DATA"), now)
	if err != nil {
		log.Fatal(err)
	}
	if err := app.recover(context.Background()); err != nil {
		log.Fatal(err)
	}
	initialOperation()
	address := os.Getenv("LAUNCHGUARD_ADDR")
	if address == "" {
		address = "127.0.0.1:21220"
	}
	server := &http.Server{Addr: address, Handler: app.handler, ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 15 * time.Second, WriteTimeout: 30 * time.Second, IdleTimeout: 60 * time.Second}
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-stop
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := app.checkpoint(); err != nil {
			log.Printf("checkpoint failed: %v", err)
		}
		if err := server.Shutdown(ctx); err != nil {
			log.Printf("shutdown failed: %v", err)
		}
	}()
	log.Printf("LaunchGuard listening on %s", address)
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}
}
