package pg

import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/De-cROMPOS/pastebin/metadbhandler/internal/partitioner"
)

func InitDB() *gorm.DB {
	dsn := "host=localhost user=admin password=loh dbname=meta_db port=5433"

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	sqlDB, _ := db.DB()
	sqlDB.SetMaxIdleConns(1)
	sqlDB.SetMaxOpenConns(1)
	sqlDB.SetConnMaxLifetime(0)

	partitioner.InitMainTable(db)


	return db
}
