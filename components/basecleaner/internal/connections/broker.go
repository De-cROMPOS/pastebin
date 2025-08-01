package connections

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/segmentio/kafka-go"
)

type KafkaCLient struct {
	reader *kafka.Reader
}

func (kc *KafkaCLient) KafkaInit() error {

	kc.reader = kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{"localhost:9092"},
		Topic:   "cleaner-topic",
		StartOffset: kafka.FirstOffset,
		MinBytes:    10e3,
		MaxBytes:    10e6,
		MaxWait:     5 * time.Second,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := kc.reader.SetOffsetAt(ctx, time.Now()); err != nil {
		return fmt.Errorf("failed to verify Kafka connection: %v", err)
	}

	return nil
}

func (kc *KafkaCLient) GetMsgs(ctx context.Context, kafkaMsgs chan []kafka.Message) {
	batch := []kafka.Message{}
	maxSize := 1000

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	kafkaMsg := make(chan kafka.Message, 100)
	defer close(kafkaMsg)

	go func() {
		for {
			select {
			case <-ctx.Done():
				log.Printf("Kafka reader stopping due to context cancellation")
				return

			default:
				msg, err := kc.reader.ReadMessage(ctx)
				if err != nil {
					select{
					case <-ctx.Done():
						log.Printf("Kafka reader stopping due to context cancellation")
						return
					default:
						log.Printf("kafka read error: %v\n", err)
						continue
					}
				}
				kafkaMsg <- msg
			}
		}
	}()

	for {
		select {
		case <-ctx.Done():
			if len(batch) > 0 {
				kafkaMsgs <- batch
				log.Printf("sent %v msgs from kafka after shutdown", len(batch))
			}
			return
		case <-ticker.C:
			if len(batch) > 0 {
				kafkaMsgs <- batch
				log.Printf("sending %v msgs from kafka to s3", len(batch))
				batch = batch[:0]
			}
		case msg := <-kafkaMsg:
			log.Printf("appending in kafka msg")
			batch = append(batch, msg)
			if len(batch) >= maxSize {
				kafkaMsgs <- batch
				log.Printf("sending %v msgs from kafka to s3", len(batch))
				batch = batch[:0]
			}
		}
	}
}

func (kc *KafkaCLient) Close() error {
	err := kc.reader.Close()
	return err
}
