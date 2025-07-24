package connections

import "time"

type metaRecord struct {
	Hash       string    `gorm:"column:hash"`
	S3URL      string    `gorm:"column:s3_url"`
	Expiration time.Time `gorm:"column:expiration"`
	CreatedAt  time.Time `gorm:"column:created_at"`
}
