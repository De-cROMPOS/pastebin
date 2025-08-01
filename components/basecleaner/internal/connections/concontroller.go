package connections

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/segmentio/kafka-go"
)

type StorageConnections struct {
	PGClient
	S3Client
	KafkaCLient
}

func (sc *StorageConnections) Init() error {
	err := sc.S3Client.S3Init()
	if err != nil {
		return fmt.Errorf("error while initializing s3")
	}

	err = sc.PGClient.PGInit()
	if err != nil {
		return fmt.Errorf("error while initializing pg")
	}

	err = sc.KafkaInit()
	if err != nil {
		sc.PGClient.Close()
		return fmt.Errorf("error while initializing kafka: %v", err)
	}

	return nil
}

func (sc *StorageConnections) Close() error {
	err := sc.PGClient.Close()
	err1 := sc.KafkaCLient.Close()
	if err != nil && err1 != nil {
		return fmt.Errorf("error while closing connections: \n1: %s\n2: %s", err, err1)
	}
	if err != nil {
		return fmt.Errorf("error while closing pg: %v", err)
	}
	if err1 != nil {
		return fmt.Errorf("error while closing kafka: %v", err1)
	}

	return nil
}

func (sc *StorageConnections) ServiceServe(ctx context.Context) error {

	kafkaMsgs := make(chan []kafka.Message, 10)
	pgMsg := make(chan string, 100)

	var wg sync.WaitGroup
	wg.Add(3)
	//todo: add more graceful shutdown && chanel closures
	go func() {
		defer close(kafkaMsgs)
		log.Printf("kafka getter started")
		sc.KafkaCLient.GetMsgs(ctx, kafkaMsgs)
		wg.Done()
	}()

	go func() {
		log.Printf("s3 deleter started")
		sc.S3Client.DeleteData(ctx, kafkaMsgs, pgMsg)
		wg.Done()
	}()

	go func() {
		defer close(pgMsg)
		log.Printf("pg deleter started")
		sc.PGClient.DeleteRows(ctx, pgMsg)
		log.Printf("pg deleter stopped working")
		wg.Done()
	}()

	wg.Wait()
	return nil
}
