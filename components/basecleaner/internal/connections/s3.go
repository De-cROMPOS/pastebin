package connections

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/segmentio/kafka-go"
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

func (sc *S3Client) DeleteData(ctx context.Context, kafkaMsg <-chan []kafka.Message, pgMsg chan<- string) error {
	bucket := "text-bin"

	exists, err := sc.minioClient.BucketExists(ctx, bucket)
	if err != nil {
		return fmt.Errorf("failed to check bucket existence: %v", err)
	}
	if !exists {
		return fmt.Errorf("bucket %s does not exist", bucket)
	}

	s3Objects := make(chan string, 100)
	defer close(s3Objects)

	s3Errors := make(chan minio.RemoveObjectError, 100)
	success := make(chan bool, 1)

	// deletion
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		defer close(s3Errors)
		defer close(success)
		opts := minio.RemoveObjectsOptions{
			GovernanceBypass: true,
		}
		batchSize := 1000
		var batch []string

		flushBatch := func() {
			if len(batch) == 0 {
				return
			}

			deletionCounter := len(batch)

			// Making chans for s3 storage
			objectsCh := make(chan minio.ObjectInfo, len(batch))
			go func() {
				defer close(objectsCh)
				for _, obj := range batch {
					objectsCh <- minio.ObjectInfo{Key: obj}
				}
			}()



			for err := range sc.minioClient.RemoveObjects(ctx, bucket, objectsCh, opts) {
				if err.Err != nil {
					log.Printf("problem with hash (%v):%v ", err.ObjectName, err.Err)
					deletionCounter--
				}
			}
			success <- true
			batch = nil
			log.Printf("deleted from s3 %v values", deletionCounter)
		}

		for {
			select {
			case <-ctx.Done():
				flushBatch()
				return
			case <-ticker.C:
				if len(batch) >= 0 {
					flushBatch()
				}
			default:
				select {
				case obj, ok := <-s3Objects:
					if !ok {
						flushBatch()
						return
					}
					batch = append(batch, obj)
					if len(batch) >= batchSize {
						flushBatch()
					}
				default:
					time.Sleep(100 * time.Millisecond)
				}
			}
		}
	}()

	// reading kafka msgs
	for {
		select {
		case <-ctx.Done():
			return nil
		case msgs, ok := <-kafkaMsg:
			if !ok {
				return nil
			}

			msgsSlice := make([]string,0,len(msgs))

			for _, msg := range msgs {
				msgsSlice = append(msgsSlice, string(msg.Value))
			}

			for _, msg := range msgsSlice {
				s3Objects <- msg
			}

			// todo: формируем мапу непрошедших, ждем сигнала от success и кидаем в постгрю за исключением непрошедших

			if ok := <-success; ok {
				for _,msg :=range msgsSlice {
					pgMsg <- msg
				}
			}
		}
	}
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

	return s3Client, nil
}
