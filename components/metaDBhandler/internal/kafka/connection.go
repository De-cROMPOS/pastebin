package kafka

import (
	"context"
	"fmt"
	"log"
	"net"
	"strconv"
	"time"

	"github.com/segmentio/kafka-go"
)

type KafkaClient struct {
	conn   *kafka.Conn
	writer *kafka.Writer
}

func (kc *KafkaClient) KafkaInit() error {
	var err error
	kc.conn, err = kafka.Dial("tcp", "localhost:9092") // установка соединения
	if err != nil {
		return fmt.Errorf("error while connecting to kafka: %v", err)
	}

	if err := kc.deleteTopicIfExists("cleaner-topic"); err != nil {
		log.Printf("Warning: failed to delete topic: %v", err)
	}

	kc.conn.CreateTopics(kafka.TopicConfig{ // Создание топика
		Topic:             "cleaner-topic", // имя
		NumPartitions:     1,               // Количество партиций в топике
		ReplicationFactor: 1,               // На сколько других брокеров копировать инфу
	})

	kc.writer = kafka.NewWriter(kafka.WriterConfig{
		Brokers:      []string{"localhost:9092"},
		Topic:        "cleaner-topic",
		Balancer:     &kafka.Hash{},          // Хэш-балансировка по ключу
		BatchSize:    100,                    // Размер батча перед отправкой
		BatchBytes:   1048576,                // 1MB максимальный размер батча
		BatchTimeout: 100 * time.Millisecond, // Таймаут формирования батча
		RequiredAcks: 1,                      // Подтверждение от одного брокера
		Async:        false,                  // Синхронная запись
	})

	// Проверка соединения
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Пробная запись пустого сообщения (только для проверки соединения)
	err = kc.writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte("connection-test"),
		Value: []byte("test"),
	})

	if err != nil {
		return fmt.Errorf("failed to verify Kafka connection: %w", err)
	}

	return nil
}

func (kc *KafkaClient) deleteTopicIfExists(topic string) error {
	// 1. Проверка существования топика
	partitions, err := kc.conn.ReadPartitions(topic)
	if err != nil || len(partitions) == 0 {
		return nil // Топик не существует
	}

	// 2. Удаление топика
	controller, err := kc.conn.Controller()
	if err != nil {
		return fmt.Errorf("failed to get controller: %w", err)
	}

	conn, err := kafka.Dial("tcp", net.JoinHostPort(controller.Host, strconv.Itoa(controller.Port)))
	if err != nil {
		return fmt.Errorf("failed to connect to controller: %w", err)
	}
	defer conn.Close()

	return conn.DeleteTopics(topic)
}

func (kc *KafkaClient) SendHashes(hashes []string) error {
	messages := make([]kafka.Message, 0, len(hashes))

	for _, hash := range hashes {
		messages = append(messages, kafka.Message{
			Key:   []byte(hash),
			Value: []byte(hash),
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := kc.writer.WriteMessages(ctx, messages...); err != nil {
		return fmt.Errorf("failed to write messages to kafka: %w", err)
	}

	return nil
}

func (kc *KafkaClient) Close() error {
	var err error
	if kc.conn != nil {
		err = kc.conn.Close()
	}
	if kc.writer != nil {
		return kc.writer.Close()
	}
	return err
}
