package main

import (
	"log"
	"net"

	hg "hashgenerator/internal/hashgen"

	pb "hashgenerator/proto"

	"google.golang.org/grpc"
)

func main() {

	log.Printf("initializing DB...")
	conDB := hg.InitDB()

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	log.Printf("starting grpc server...")
	s := grpc.NewServer()
	pb.RegisterHasherServer(s, &hg.HgProtoServer{
		DB: conDB,
	})

	log.Printf("starting serving grpc...")
	if err := s.Serve(lis); err != nil {
		log.Fatal(err)
	}
}
