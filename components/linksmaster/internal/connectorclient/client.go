package connectorclient

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	pb "github.com/De-cROMPOS/pastebin/hashgenerator/proto"
)

// TODO: init, get hash == http serv, graceful shutdown to close conn...
type ConnectorClient struct {
	GrpcClient
	S3Client
	PGClient
}

// TODO: хэлфчеки + переподнятие упавших
func (cc *ConnectorClient) Init() error {
	err := cc.GrpcInit()
	if err != nil {
		return fmt.Errorf("grpc didn't connect: %v", err)
	}

	// In case if err will happen next
	defer func() {
		if err != nil {
			cc.grpcConn.Close()
		}
	}()

	err = cc.S3Init()
	if err != nil {
		return fmt.Errorf("s3 didn't connect: %v", err)
	}

	err = cc.PGInit()
	if err != nil {
		return fmt.Errorf("pg didn't connect: %v", err)
	}

	return nil
}

func (connectorClient *ConnectorClient) asyncLoader(hash, text string, ttl time.Duration) error {

	err := connectorClient.AddTextToS3(&hash, &text)
	if err != nil {
		return fmt.Errorf("failed to upload to s3: %v", err)
	}

	// s3 link gen
	url, err := connectorClient.GetLinkFromS3(&hash, &ttl)
	if err != nil {
		return fmt.Errorf("failed to get link from s3: %v", err)
	}

	expTime := time.Now().Add(ttl)

	err = connectorClient.InsertPGData(hash, url.String(), expTime)
	if err != nil {
		return fmt.Errorf("failed to put data into pg: %v", err)
	}

	return nil
}

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

	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()
	resp, err := connectorclient.hasherClient.GetHash(ctx, &pb.HashRequest{Text: data.Text})
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte(`{"error": "grpc service error"}`))
		return
	}

	response := map[string]string{
		"hash": resp.GetHash(),
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)

	//todo: сделать пул воркеров для асинклоадера
	go func() {
		duration, err := time.ParseDuration(data.TTL)
		if err != nil {
			duration = 24 * time.Hour
		}

		err = connectorclient.asyncLoader(resp.GetHash(), data.Text, duration)
		// todo: мб через брокер докидывать если ошибка случится
		// или пытаться переподключиться и по новой грузить и данные по идее тут будут копитсься...
		if err != nil {
			log.Printf("smth went wrong while writing to storages: %v", err)
		}
	}()


}

func (connectorclient *ConnectorClient) Close() {
	log.Printf("pg is shutting down...")
	if err := connectorclient.PGClient.Close(); err != nil {
		log.Printf("pg shutdown error: %s", err)
	}
	log.Printf("pg stopped")

	log.Printf("grpc is shutting down...")
	if err := connectorclient.grpcConn.Close(); err != nil {
		log.Printf("grpc shutdown error: %s", err)
	}
	log.Printf("grpc stopped")
}
