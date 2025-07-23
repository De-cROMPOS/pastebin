package connectorclient

import "time"

type reqData struct {
	Text string `json:"text"`
	TTL  string `json:"ttl"`
}

type metaTable struct {
    Hash       string    `gorm:"primaryKey"`
    S3URL      string    `gorm:"column:s3_url"`
    Expiration time.Time `gorm:"primaryKey"`
    CreatedAt  time.Time `gorm:"default:now()"`
}
