package connectorclient

import (
	"fmt"

	pb "github.com/De-cROMPOS/pastebin/hashgenerator/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type GrpcClient struct {
	grpcConn     *grpc.ClientConn
	hasherClient pb.HasherClient
}

func (gc *GrpcClient) GrpcInit() error {
	var err error
	gc.grpcConn, err = getGrpcConn()
	if err != nil {
		return fmt.Errorf("grpc didn't connect. %v", err)
	}

	gc.hasherClient = pb.NewHasherClient(gc.grpcConn)

	return nil
}

func getGrpcConn() (*grpc.ClientConn, error) {
	clientConn, err := grpc.NewClient("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("grpc didn't connect. %v", err)
	}

	return clientConn, nil
}
