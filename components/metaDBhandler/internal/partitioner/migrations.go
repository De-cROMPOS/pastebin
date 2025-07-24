package partitioner

import (
	"fmt"
	"log"
	"time"

	"gorm.io/gorm"
)

func InitMainTable(db *gorm.DB) {
	db.Exec(`
		CREATE TABLE IF NOT EXISTS meta_table (
			hash VARCHAR(255),
			s3_url TEXT,
			expiration TIMESTAMPTZ,
			created_at TIMESTAMPTZ DEFAULT NOW(),
			PRIMARY KEY (hash, expiration)
		)
		PARTITION BY RANGE (expiration)
	`)
}

func CreateNewPartition(db *gorm.DB, timeFrom time.Time) error {
	UTCtime := timeFrom.UTC()
	starterTime := time.Date(
		UTCtime.Year(),
		UTCtime.Month(),
		UTCtime.Day(),
		UTCtime.Hour(),
		0,
		0,
		0,
		time.UTC,
	)

	endTime := starterTime.Add(1 * time.Hour)

	partName := fmt.Sprintf("meta_table_%s", starterTime.Format("20060102_15"))

	query := fmt.Sprintf(`
        CREATE TABLE IF NOT EXISTS %s PARTITION OF meta_table
        FOR VALUES FROM ('%s') TO ('%s')
    `,
		partName,
		starterTime.Format("2006-01-02 15:04:05"),
		endTime.Format("2006-01-02 15:04:05"),
	)

	if err := db.Exec(query).Error; err != nil {
		return fmt.Errorf("failed to create partition: %v", err)
	}

	log.Printf("partition created: %s", partName)
	return nil
}

func DropOldPartition(db *gorm.DB) error {
	timeNow := time.Now().UTC()
	timeToDrop := timeNow.Add(-1 * time.Hour)

	starterTime := time.Date(
		timeToDrop.Year(),
		timeToDrop.Month(),
		timeToDrop.Day(),
		timeToDrop.Hour(),
		0,
		0,
		0,
		time.UTC,
	)
// todo: add kafka with query data
	oldPartName := fmt.Sprintf("meta_table_%s", starterTime.Format("20060102_15"))
	dropQuery := fmt.Sprintf(`DROP TABLE IF EXISTS %s`, oldPartName)
	if err := db.Exec(dropQuery).Error; err != nil {
		return fmt.Errorf("failed to drop old partition %s: %v", oldPartName, err)
	}

	log.Printf("partition deleted: %s", oldPartName)
	return nil
}
