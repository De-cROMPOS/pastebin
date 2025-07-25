package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/De-cROMPOS/pastebin/linksmaker/internal/connectorclient"
)

func main() {
	var c connectorclient.ConnectorClient

	log.Println("Initializing connections to storages...")
	if err := c.Init(); err != nil {
		log.Fatalf("Failed to initialize: %v", err)
	}
	defer c.Close()
	log.Printf("connected, starting serving")

	server := &http.Server{
		Addr:    ":1234",
		Handler: http.HandlerFunc(c.HashHandler),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Printf("Server started")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	<-done
	log.Println("Server is shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Server shutdown failed: %v", err)
	}

	log.Println("Server stopped")
}
