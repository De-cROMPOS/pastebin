package connections

import (
	"fmt"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type PGClient struct {
	pgConn *gorm.DB
}

func (pgc *PGClient) Close() error {
    if pgc.pgConn != nil {
        db, err := pgc.pgConn.DB()
        if err != nil {
            return fmt.Errorf("failed to get generic database object: %w", err)
        }
        return db.Close()
    }
    return nil
}

func (pgc *PGClient) PGInit() error {
	dsn := "host=localhost user=admin password=loh dbname=meta_db port=5433"

	var err error
	pgc.pgConn, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("couldn't open db connection")
	}

	sqlDB, _ := pgc.pgConn.DB()
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	return nil
}

func (pgc *PGClient) GetMeta(hash string) (metaRecord, error) {
	var meta metaRecord
	if hash == "" {
		return meta, fmt.Errorf("empty hash")
	}

	err := pgc.pgConn.Table("meta_table").
		Where("hash = ? AND expiration > ?", hash, time.Now().UTC()).
		First(&meta).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return meta, fmt.Errorf("link not found or expired")
		}
		return meta, fmt.Errorf("database error: %w", err)
	}

	return meta, nil
}

func (pgc *PGClient) GetLink(hash string) (string, error) {
	var meta metaRecord

	err := pgc.pgConn.Table("meta_table").
		Where("hash = ? AND expiration > ?", hash, time.Now().UTC()).
		First(&meta).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return "", fmt.Errorf("link not found or expired")
		}
		return "", fmt.Errorf("database error: %w", err)
	}

	return meta.S3URL, nil
}
