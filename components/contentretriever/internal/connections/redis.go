package connections

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
	er "github.com/De-cROMPOS/pastebin/contentretriever/internal/errorhandler"
)

type RClient struct {
	rConn *redis.Client
	ctx   context.Context
}

func (rc *RClient) Close() error {
	err := rc.rConn.Close()
	if err != nil {
		return fmt.Errorf("error while closing redis conn: %v", err)
	}
	return nil
}

func (rc *RClient) RedisInit() error {
	addr := "localhost:6379"
	pswd := "lohlohloh"

	rc.ctx = context.Background()

	rc.rConn = redis.NewClient(&redis.Options{
		Addr:         addr,
		Password:     pswd,
		DB:           0,
		PoolSize:     100,
		MinIdleConns: 10,
	})

	_, err := rc.rConn.Ping(rc.ctx).Result()
	if err != nil {
		return fmt.Errorf("failed to connect to redis: %s", err)
	}

	return nil
}

func (rc *RClient) GetLink(hash string) (string, error) {
	data, err := rc.rConn.HGet(rc.ctx, hash, "data").Result()
	if err != nil {
		if err == redis.Nil {
			return "", er.ErrRecordNotFound
		}
		return "", fmt.Errorf("failed to get data: %v", err)
	}

	// updating lru
	go func() {
		rc.rConn.ZAdd(rc.ctx, "lru:meta_records", redis.Z{
			Score:  float64(time.Now().UnixNano()),
			Member: hash,
		})
	}()

	var record metaRecord
	if err := json.Unmarshal([]byte(data), &record); err != nil {
		return "", fmt.Errorf("failed to unmarshal data: %v", err)
	}

	log.Printf("meta was given from redis")
	return record.S3URL, nil
}

func (rc *RClient) AddMeta(record metaRecord) error {
	data, err := json.Marshal(record)
	if err != nil {
		return err
	}

	ttl := time.Until(record.Expiration)

	pipe := rc.rConn.Pipeline()
	pipe.HSet(rc.ctx, record.Hash, "data", data)
	pipe.Expire(rc.ctx, record.Hash, ttl)
	pipe.ZAdd(rc.ctx, "lru:meta_records", redis.Z{
		Score:  float64(time.Now().UnixNano()),
		Member: record.Hash,
	})
	_, err = pipe.Exec(rc.ctx)

	log.Printf("meta was inserted")
	return err
}
