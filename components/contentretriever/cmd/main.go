package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/De-cROMPOS/pastebin/contentretriever/internal/content"
)

func main() {

	log.Printf("initializing conftroller...")
	var controller content.ContentController
	if err := controller.Initialize(); err != nil {
		log.Fatalf("smth went wrong while initializing controller %s", err)
	}
	log.Printf("controller initialized...")

	server := &http.Server{
		Addr:         ":4321",
		Handler:      http.HandlerFunc(controller.GetHandler),
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
