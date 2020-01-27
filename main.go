package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	pb "github.com/acubed-tm/authentication-service/protofiles"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc"
)

const port = ":50551"

type server struct{}

func (*server) IsEmailRegistered(ctx context.Context, req *pb.IsEmailRegisteredRequest) (*pb.IsEmailRegisteredReply, error) {
	_, err := GetEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	} else {
		return &pb.IsEmailRegisteredReply{IsRegistered:true}, nil
	}
}

func (*server) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginReply, error) {
	ret, err := GetPasswordByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	} else {
		log.Printf("ret: %v", ret)
	}

	bcryptErr := bcrypt.CompareHashAndPassword([]byte(ret), []byte(req.Password))
	if bcryptErr != nil {
		return nil, errors.New(fmt.Sprintf("incorrect password, %v", bcryptErr))
	}

	return &pb.LoginReply{Success: true}, nil
}

func main() {
	isClient := len(os.Args[1:]) == 1 && os.Args[1:][0] == "client"

	if isClient {
		runClient()
	} else {
		runServer()
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
	resp, err := c.Login(ctx, &pb.LoginRequest{Email: "john.doe@mail.be", Password: "test123"})
	if err != nil {
		log.Fatalf("could not log in: %v", err)
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
