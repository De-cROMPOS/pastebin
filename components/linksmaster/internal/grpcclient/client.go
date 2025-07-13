package grpcclient

import (
	"fmt"
	"net/http"

	pb "github.com/De-cROMPOS/pastebin/hashgenerator/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// TODO: init, get hash == http serv, graceful shutdown to close conn...
type GrpcClient struct {
	c pb.HasherClient
	conn *grpc.ClientConn
}

func InitGrpcClient() (*GrpcClient, error) {
	clientConn, err := grpc.NewClient("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("didn't connect. %v", err)
	}
	hasherClient := pb.NewHasherClient(clientConn)

	return &GrpcClient{
		c: hasherClient,
		conn: clientConn,
	}, nil
}

//todo: подумать бы:
// нужно вытащить из запроса ttl, текст и выдать сразу хэш и параллельно походить по бдшкам и позаписывать хм
func (grpcClient *GrpcClient) HashHandler (w http.ResponseWriter, r *http.Request) {
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



} 

func (grpcClient *GrpcClient) Close () {
	grpcClient.conn.Close()
}