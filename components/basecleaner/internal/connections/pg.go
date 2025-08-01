package connections

import (
	"context"
	"fmt"
	"log"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type PGClient struct {
	pgConn *gorm.DB
}

func (c *PGClient) PGInit() error {
	var err error

	dsn := "host=localhost user=admin password=loh dbname=hash_db port=5432"

	c.pgConn, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("cant connect to pg")
	}

	sqlDB, _ := c.pgConn.DB()
	sqlDB.SetMaxIdleConns(1)
	sqlDB.SetMaxOpenConns(1)
	sqlDB.SetConnMaxLifetime(time.Hour)

	return nil
}

func (c *PGClient) Close() error {
	sqlDB, _ := c.pgConn.DB()

	err := sqlDB.Close()
	if err != nil {
		return fmt.Errorf("error while closing pg")
	}

	return nil
}

func (c *PGClient) DeleteRows(ctx context.Context, pgMsg <-chan string) error {

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	hashes := []string{}
	batchSize := 100

	for {
		select {
		case <-ctx.Done():
			if len(hashes) > 0 {
				if err := c.Delete(hashes); err != nil {
					log.Printf("%v", err)
				}
			}
			return nil
		case hash, ok := <-pgMsg:
			if !ok && len(hashes) > 0 {
				if err := c.Delete(hashes); err != nil {
					log.Printf("%v", err)
				}
				return nil
			}

			hashes = append(hashes, hash)
			if len(hashes) >= batchSize {
				if err := c.Delete(hashes); err != nil {
					log.Printf("%v", err)
				}
				hashes = hashes[:0]
			}
		case <-ticker.C:
			if len(hashes) > 0 {
				if err := c.Delete(hashes); err != nil {
					log.Printf("%v", err)
				}
				hashes = hashes[:0]
			}
		}
	}
}

func (c *PGClient) Delete(hashes []string) error {
	result := c.pgConn.Table("hash_table").Where("hash in ?", hashes).Delete(nil)
	if result.Error != nil {
		return fmt.Errorf("failed to delete batch: %v", result.Error)
	}
	log.Printf("Deleted %d hashes from PG", result.RowsAffected)
	return nil
}
