package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/De-cROMPOS/pastebin/metadbhandler/internal/pg"
	"github.com/De-cROMPOS/pastebin/metadbhandler/internal/partitioner"
	"github.com/De-cROMPOS/pastebin/metadbhandler/internal/kafka"
)

func main() {
	// connection to + kafka

	db := pg.InitDB()
	
	broker := &kafka.KafkaClient{}
	if err := broker.KafkaInit(); err != nil {
		log.Fatalf("smth went wrong while initializing kafka: %v", err)
	}

	t := time.Now()
	for range 26 {
		err := partitioner.CreateNewPartition(db, t)
		if err != nil {
			log.Printf("error while creating particion: %s", err)
		}
		t = t.Add(1 * time.Hour)
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	checkInterval := 1 * time.Hour
	ticker := time.NewTicker(checkInterval)
	defer ticker.Stop()

	log.Printf("Service is ready to use")
	for {
		select {
		case <-ticker.C:
			go func() {
				err := partitioner.DropOldPartition(db, broker)
				if err != nil {
					log.Printf("error while dropping particion: %s", err)
				}
				err = partitioner.CreateNewPartition(db, t)
				t = t.Add(1 * time.Hour)
				if err != nil {
					log.Printf("error while creating particion: %s", err)
				}
			}()
		case sig := <-sigChan:
			log.Printf("Received signal: %v. Shutting down...", sig)
			sqlDB, _ := db.DB()
			sqlDB.Close()
			return
		}
	}

}
