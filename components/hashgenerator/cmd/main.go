package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	hg "github.com/De-cROMPOS/pastebin/hashgenerator/internal/hashgen"

	pb "github.com/De-cROMPOS/pastebin/hashgenerator/proto"

	"google.golang.org/grpc"
)

func main() {
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	log.Printf("initializing DB...")
	conDB := hg.InitDB()
	defer func() {
		if sqlDB, err := conDB.DB(); err == nil {
			sqlDB.Close()
			log.Printf("DB connection closed")
		}
	}()
	log.Printf("DB initialized!")

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	log.Printf("starting grpc server...")
	s := grpc.NewServer(
		grpc.MaxConcurrentStreams(1000),
		grpc.NumStreamWorkers(20),
		grpc.ConnectionTimeout(10*time.Second),
	)
	pb.RegisterHasherServer(s, &hg.HgProtoServer{
		DB: conDB,
	})

	log.Printf("serving grpc...")
	go func() {
		if err := s.Serve(lis); err != nil {
			log.Fatal(err)
		}
	}()

	<-shutdown
	log.Printf("got shutdown signal")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	stopped := make(chan struct{})
	go func() {
		s.GracefulStop()
		close(stopped)
	}()

	select {
	case <-stopped:
		log.Printf("gRPC server stopped gracefully")
	case <-ctx.Done():
		s.Stop()
		log.Printf("gRPC server force stopped ")
	}
}
