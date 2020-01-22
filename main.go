package main

import (
	"context"
	"log"
	"net"
	"os"
	"time"

	pb "github.com/acubed-tm/authentication-service/proto"
	grpc "google.golang.org/grpc"
)

const port = ":50551"

type server struct {
	pb.UnimplementedTestServiceServer
}

func main() {
	is_client := len(os.Args[1:]) == 1 && os.Args[1:][0] == "client"

	if is_client {
		run_server()
	} else {
		run_client()
	}
}

func run_client() {
	conn, err := grpc.Dial("localhost" + port, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect: %v", err)

	}
	defer conn.Close()
	c := pb.NewTestServiceClient(conn)

	// Contact the server and print out its response.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	_, err = c.Test(ctx, &pb.TestRequest{})
	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}
	log.Printf("Greeting")
}

func run_server() {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterTestServiceServer(s, &server{})
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
