package main

import (
	"context"
	"log"
	"net"
	"os"
	"time"

	pb "github.com/acubed-tm/authentication-service/proto"
	"google.golang.org/grpc"
)

const port = ":50551"

type server struct{}

func (*server) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginReply, error) {
	ret, err := GetUserByEmail(ctx, "john.doe@mail.be")
	if err != nil {
		log.Printf("Error: %v", err)
	} else {
		log.Printf("ret: %v", ret)
	}
	return &pb.LoginReply{Success: true}, nil
}

func main() {
	isClient := len(os.Args[1:]) == 1 && os.Args[1:][0] == "client"

	if isClient {
		runServer()
	} else {
		runClient()
	}
}

func runClient() {
	conn, err := grpc.Dial("localhost"+port, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	//noinspection GoUnhandledErrorResult
	defer conn.Close()
	c := pb.NewLoginServiceClient(conn)

	// Contact the server and print out its response.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	resp, err := c.Login(ctx, &pb.LoginRequest{})
	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}
	log.Printf("Login success: %v", resp.Success)
}

func runServer() {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterLoginServiceServer(s, &server{})
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
