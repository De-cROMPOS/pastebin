package main

import (
	"context"
	"log"
	"time"

	pb "github.com/De-cROMPOS/pastebin/hashgenerator/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	conn, err := grpc.NewClient("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewHasherClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	starterTime := time.Now()

	r, err := c.GetHash(ctx, &pb.HashRequest{Text: "Hello"})
	if err != nil {
		log.Fatalf("coulnd get grpc response: %v", err)
	}
	log.Printf("Message hash: %v, time spent: %v", r.GetHash(), time.Since(starterTime))

}
