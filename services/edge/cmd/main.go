// Copyright 2026 Invariant Solutions LLC
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/invariant-solutions-ai/microservice-reference-platform/services/edge/internal/config"
	"github.com/invariant-solutions-ai/microservice-reference-platform/services/edge/internal/server"
	"github.com/joho/godotenv"
)

const shutdownTimeout = 10 * time.Second

func main() {
	// read configuration
	if err := godotenv.Load(); err != nil {
		log.Println(".env file not loaded")
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %s", err)
		os.Exit(1)
	}

	handler := server.NewHandler(cfg.OrdersURL)

	// initialize edge service
	s := &http.Server{
		Addr:           cfg.Addr,
		Handler:        handler,
		ReadTimeout:    5 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1 MB
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// start edge service
	go func() {
		if err := s.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()

	// gracefully shutdown edge service
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	if err := s.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %s", err)
		os.Exit(1)
	}

	log.Println("Server exiting")
}
