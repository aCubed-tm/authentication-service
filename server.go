package main

import (
	"context"
	"fmt"
	pb "github.com/acubed-tm/authentication-service/protofiles"
	"golang.org/x/crypto/bcrypt"
	"log"
)

type server struct{}

func (*server) IsEmailRegistered(ctx context.Context, req *pb.IsEmailRegisteredRequest) (*pb.IsEmailRegisteredReply, error) {
	_, err := GetEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	} else {
		return &pb.IsEmailRegisteredReply{IsRegistered: true}, nil
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
