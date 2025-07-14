package connectorclient

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"

	pb "github.com/De-cROMPOS/pastebin/hashgenerator/proto"
	"google.golang.org/grpc"
)

type reqData struct {
	Text string `json:"text"`
	TTL  string `json:"ttl"`
}

// TODO: init, get hash == http serv, graceful shutdown to close conn...
type ConnectorClient struct {
	hc       pb.HasherClient
	grpcConn *grpc.ClientConn

	s3Conn *minio.Client

	// TODO:
	// redis conn
	// postgres conn
	// s3 conn
}

// TODO: хэлфчеки + переподнятие упавших
func Initconnectorclient() (*ConnectorClient, error) {
	grpcConn, err := getGrpcConn()
	if err != nil {
		return nil, fmt.Errorf("grpc didn't connect: %v", err)
	}

	defer func() {
		if err != nil {
			grpcConn.Close()
		}
	}()

	hasherClient := pb.NewHasherClient(grpcConn)

	s3Conn, err := getS3Conn()
	if err != nil {
		return nil, fmt.Errorf("s3 didn't  connect: %v", err)
	}

	// TODO:
	// redis conn
	// postgre conn

	return &ConnectorClient{
		hc:       hasherClient,
		grpcConn: grpcConn,
		s3Conn:   s3Conn,
	}, nil
}

func (connectorClient *ConnectorClient) asyncLoader(hash, text string, ttl time.Duration) (string, error) {

	ctx := context.Background()

	// loading text into s3
	_, err := connectorClient.s3Conn.PutObject(
		ctx,
		"text-bin",
		hash,
		strings.NewReader(text),
		int64(len(text)),
		minio.PutObjectOptions{
			ContentType: "text/plain",
		},
	)
	if err != nil {
		return "", fmt.Errorf("failed to upload to storage: %w", err)
	}

	// link gen
	url, err := connectorClient.s3Conn.PresignedGetObject(
		ctx,
		"text-bin",
		hash,
		ttl,
		nil,
	)
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	fmt.Println("We got an url:", url.String())
	return url.String(), nil
}

// todo: подумать бы:
func (connectorclient *ConnectorClient) HashHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte(`{"error": "method not allowed"}`))
		return
	}

	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error": "bad content type, expected application/json"}`))
		return
	}

	var data reqData
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error": "bad content"}`))
		return
	}

	resp, err := connectorclient.hc.GetHash(r.Context(), &pb.HashRequest{Text: data.Text})
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte(`{"error": "grpc service error"}`))
	}

	response := map[string]string{
		"hash": resp.GetHash(),
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)

	go func() {
		duration, err := time.ParseDuration(data.TTL)
		if err != nil {
			duration = 24 * time.Hour
		}

		connectorclient.asyncLoader(resp.GetHash(), data.Text, duration)
	}()

	// TODO:
	// grpc req
	// async writing to DB's
	// http response

}

func (connectorclient *ConnectorClient) Close() {
	// TODO:
	// redis closure
	// postrge closure
	// s3 closure
	connectorclient.grpcConn.Close()
}
