package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	// connection to postgres + kafka

	// че то типо миграций должно быть
	// Основная таблица
	// CREATE TABLE hash_table (
	//     hash VARCHAR(255),
	//     s3_url TEXT,
	//     expiration TIMESTAMPTZ,
	//     created_at TIMESTAMPTZ DEFAULT NOW(),
	//     PRIMARY KEY (hash, expiration)
	// ) PARTITION BY RANGE (expiration);

	// Партиция за текущий час
	// CREATE TABLE hash_table_20240620_15 PARTITION OF hash_table
	//     FOR VALUES FROM ('2024-06-20 15:00:00') TO ('2024-06-20 16:00:00');



	// creating partitions (now - now + 1h; now+1h - now + 2h; ... ; now + 25h - now +26h )
	// now в часах без минут

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	checkInterval := 1 * time.Hour
	ticker := time.NewTicker(checkInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// go managePartitions добавляем новую партицию (now+24h+1 : now+24h+2) и удаленяем старую
		case sig := <-sigChan:
			log.Printf("Received signal: %v. Shutting down...", sig)
			return
		}
	}

}
