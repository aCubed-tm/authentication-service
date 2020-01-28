package main

import (
	"context"
	"errors"
	"fmt"
	pb "github.com/acubed-tm/authentication-service/protofiles"
	"golang.org/x/crypto/bcrypt"
	"log"
	"strings"
)

const passwordCost = 12

type server struct{}

func (*server) IsEmailRegistered(ctx context.Context, req *pb.IsEmailRegisteredRequest) (*pb.IsEmailRegisteredReply, error) {
	_, err := GetEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	} else {
		return &pb.IsEmailRegisteredReply{IsRegistered: true}, nil
	}
}

func (*server) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterReply, error) {
	log.Printf("Starting registration")
	email, err := GetEmailByVerificationToken(ctx, req.VerificationToken)
	if err != nil {
		return nil, err
	}
	log.Printf("Found email %v", email)

	if !strings.EqualFold(email, req.Email) {
		return nil, errors.New("email did not match verification token")
	}

	// check if already has account
	// could reduce db calls, but we're students so who cares
	pass, err := GetPasswordByEmail(ctx, email)
	if pass != "" {
		return nil, errors.New("user already has a password set")
	}

	// NOTE: could also remove verification token
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(pass), passwordCost)
	if err != nil {
		return nil, err
	}
	err = ChangePasswordForEmail(ctx, email, string(hashedPassword))

	if err != nil {
		return nil, err
	} else {
		return &pb.RegisterReply{Success: true}, nil
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
