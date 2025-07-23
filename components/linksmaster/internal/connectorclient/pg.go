package connectorclient

import (
	"fmt"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type PGClient struct {
	pgConn *gorm.DB
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

func (pgc *PGClient) InsertPGData(hash, s3URL string, expiration time.Time) error {
	utcExpiration := expiration.UTC()
	record := metaTable{
		Hash:       hash,
		S3URL:      s3URL,
		Expiration: utcExpiration,
	}

	partName := fmt.Sprintf("meta_table_%s", utcExpiration.Format("20060102_15"))

	result := pgc.pgConn.Table(partName).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "hash"}, {Name: "expiration"}},
		DoNothing: true,
	}).Create(&record)

	return result.Error
}
