package connectorclient

import (
	"context"
	"fmt"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/minio/minio-go/v7/pkg/lifecycle"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func getGrpcConn() (*grpc.ClientConn, error) {
	clientConn, err := grpc.NewClient("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("grpc didn't connect. %v", err)
	}

	return clientConn, nil
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
