package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/De-cROMPOS/pastebin/basecleaner/internal/connections"
)

func main() {
	sc := connections.StorageConnections{}
	if err := sc.Init(); err != nil {
		log.Fatalf("error while initializing connections: %v", err)
	}
	defer sc.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, syscall.SIGINT, syscall.SIGTERM)
	log.Printf("server started working...")
	go func() {
		if err := sc.ServiceServe(ctx); err != nil {
			log.Printf("Service stopped with error: %v", err)
			cancel()
		}
	}()

	<-stopChan
    log.Println("Received shutdown signal")

    cancel()

}
