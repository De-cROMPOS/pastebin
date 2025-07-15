package connectorclient

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/minio/minio-go/v7/pkg/lifecycle"
)

type S3Client struct {
	minioClient *minio.Client
}

func (sc *S3Client) S3Init() error {
	var err error
	sc.minioClient, err = getS3Conn()
	if err != nil {
		return fmt.Errorf("failed to initialize MinIO client: %v", err)
	}

	return nil
}

func getS3Conn() (*minio.Client, error) {
	endpoint := "localhost:9000"
	accessKey := "admin"
	secretKey := "lohlohloh"
	useSSL := false

	s3Client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to initialize MinIO client: %v", err)
	}

	bucketExists, err := s3Client.BucketExists(context.Background(), "text-bin")
	if err != nil {
		return nil, fmt.Errorf("failed to connect to S3: %v", err)
	}

	if !bucketExists {
		err = s3Client.MakeBucket(context.Background(), "text-bin", minio.MakeBucketOptions{})
		if err != nil {
			return nil, fmt.Errorf("failed to create bucket in S3: %v", err)
		}
		err = SetAutoDeletePolicy(s3Client, "text-bin", 1)
		if err != nil {
			return nil, fmt.Errorf("failed to set autodelete in S3: %v", err)
		}
	}

	return s3Client, nil
}

func (sc *S3Client) AddTextToS3(hash, text *string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// loading text into s3
	_, err := sc.minioClient.PutObject(
		ctx,
		"text-bin",
		*hash,
		strings.NewReader(*text),
		int64(len(*text)),
		minio.PutObjectOptions{
			ContentType: "text/plain",
		},
	)
	if err != nil {
		return fmt.Errorf("failed to upload to storage: %w", err)
	}

	return nil
}

func (sc *S3Client) GetLinkFromS3(hash *string, ttl *time.Duration) (*url.URL, error) {
	ctx := context.Background()

	// link gen
	url, err := sc.minioClient.PresignedGetObject(
		ctx,
		"text-bin",
		*hash,
		*ttl,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	return url, nil
}

func SetAutoDeletePolicy(minioClient *minio.Client, bucketName string, daysToExpire int) error {
	config := lifecycle.NewConfiguration()
	expDays := lifecycle.ExpirationDays(daysToExpire)
	config.Rules = []lifecycle.Rule{
		{
			ID:     "autodelete-after-days",
			Status: "Enabled",
			Expiration: lifecycle.Expiration{
				Days: expDays,
			},
		},
	}

	err := minioClient.SetBucketLifecycle(context.Background(), bucketName, config)
	return err
}
