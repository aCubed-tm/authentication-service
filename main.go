package main

import (
	"log"
	"net"

	pb "github.com/acubed-tm/authentication-service/protofiles"
	"google.golang.org/grpc"
)

const port = ":50551"

func main() {
	log.Print("Starting server")
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterAuthServiceServer(s, &server{})
	for {
		if err := s.Serve(lis); err != nil {
			log.Fatalf("Failed to serve: %v", err)
		}
	}
}
