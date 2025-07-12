package hashgen

import (
	"fmt"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// setting up DB connection with pool
func InitDB() *gorm.DB {
	dsn := "host=localhost user=admin password=loh dbname=hash_db port=5432"

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	sqlDB, _ := db.DB()
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	
	return db
}

func HashChecker(connDB *gorm.DB, hash string) (bool, error) {

	res := connDB.Exec(
        `INSERT INTO hash_table (hash) 
         VALUES (?) 
         ON CONFLICT (hash) DO NOTHING`,
        hash,
    )

	if res.Error != nil {
		return false, fmt.Errorf("insertion error: %v", res.Error)
	}

    return res.RowsAffected == 0, nil
}
