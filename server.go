package main

import (
	"context"
	"errors"
	"fmt"
	pb "github.com/acubed-tm/authentication-service/protofiles"
	"golang.org/x/crypto/bcrypt"
	"log"
	"time"
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

func (s *server) GetInvites(ctx context.Context, req *pb.GetInvitesRequest) (*pb.GetInvitesReply, error) {
	accountUuid := req.AccountUuid
	emails, err := GetAllEmailsByUuid(ctx, accountUuid)
	if err != nil {
		return nil, err
	}

	var ret []string
	for _, email := range emails {
		invites, err := GetInviteOrganizationsByEmail(ctx, email)
		if err != nil {
			return nil, err
		}
		for _, invite := range invites {
			ret = append(ret, invite)
		}
	}

	return &pb.GetInvitesReply{
		OrganizationUuids: ret,
	}, nil
}

func (*server) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterReply, error) {
	log.Printf("Starting registration")
	email := req.Email

	// check if already has account
	// could reduce db calls, but we're students so who cares
	pass, err := GetPasswordByEmail(ctx, email)
	if pass != "" {
		return nil, errors.New("user already has a password set")
	}

	// could save this request
	invitations, err := GetInviteOrganizationsByEmail(ctx, email)
	if len(invitations) == 0 {
		return nil, errors.New("user is not invited by any organizations")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(pass), passwordCost)
	if err != nil {
		return nil, err
	}
	err = ChangePasswordForEmail(ctx, email, string(hashedPassword))

	if err != nil {
		return nil, err
	} else {
		return &pb.RegisterReply{}, nil
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

	uuid, err := GetUuidByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}

	tokenString, err := CreateToken(uuid)
	if err != nil {
		return nil, err
	}

	err = AddJwtTokenToUser(ctx, uuid, tokenString)
	if err != nil {
		return nil, err
	}

	return &pb.LoginReply{Token: tokenString}, nil
}

func (s *server) ActivateEmail(ctx context.Context, req *pb.ActivateEmailRequest) (*pb.ActivateEmailReply, error) {
	verificationToken := req.Token
	err := VerifyEmailByToken(ctx, verificationToken, time.Now())
	if err != nil {
		return nil, err
	}
	return &pb.ActivateEmailReply{}, nil
}

func (s *server) DropSingleToken(ctx context.Context, req *pb.DropSingleTokenRequest) (*pb.DropSingleTokenReply, error) {
	jwtToken := req.Token
	return &pb.DropSingleTokenReply{}, RemoveJwtTokenByToken(ctx, jwtToken, true)
}

func (s *server) DropAllTokens(ctx context.Context, req *pb.DropAllTokensRequest) (*pb.DropAllTokensReply, error) {
	jwtToken := req.Token
	return &pb.DropAllTokensReply{}, RemoveJwtTokenByToken(ctx, jwtToken, true)
}
