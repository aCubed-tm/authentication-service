package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	pb "github.com/acubed-tm/authentication-service/protofiles"
	"golang.org/x/crypto/bcrypt"
)

const passwordCost = 12

type server struct{}

func (*server) IsEmailRegistered(_ context.Context, req *pb.IsEmailRegisteredRequest) (*pb.IsEmailRegisteredReply, error) {
	uuid, err := GetUuidByEmail(req.Email)
	if err != nil {
		return nil, err
	} else {
		return &pb.IsEmailRegisteredReply{IsRegistered: true, AccountUuid: uuid}, nil
	}
}

func (s *server) GetInvites(_ context.Context, req *pb.GetInvitesRequest) (*pb.GetInvitesReply, error) {
	accountUuid := req.AccountUuid
	emails, err := GetAllEmailsByUuid(accountUuid)
	if err != nil {
		return nil, err
	}

	var ret []string
	for _, email := range emails {
		invites, err := GetInviteOrganizationsByEmail(email)
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

func (*server) Register(_ context.Context, req *pb.RegisterRequest) (*pb.RegisterReply, error) {
	log.Printf("Starting registration")
	email := req.Email

	// check if already has account
	// could reduce db calls, but we're students so who cares
	pass, err := GetPasswordByEmail(email)
	if pass != "" {
		return nil, errors.New("user already has a password set")
	}

	// could save this request
	invitations, err := GetInviteOrganizationsByEmail(email)
	if len(invitations) == 0 {
		return nil, errors.New("user is not invited by any organizations")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(pass), passwordCost)
	if err != nil {
		return nil, err
	}
	err = ChangePasswordForEmail(email, string(hashedPassword))

	if err != nil {
		return nil, err
	} else {
		return &pb.RegisterReply{}, nil
	}
}

func (*server) Login(_ context.Context, req *pb.LoginRequest) (*pb.LoginReply, error) {
	ret, err := GetPasswordByEmail(req.Email)
	if err != nil {
		return nil, err
	} else {
		log.Printf("ret: %v", ret)
	}

	bcryptErr := bcrypt.CompareHashAndPassword([]byte(ret), []byte(req.Password))
	if bcryptErr != nil {
		return nil, errors.New(fmt.Sprintf("incorrect password, %v", bcryptErr))
	}

	uuid, err := GetUuidByEmail(req.Email)
	if err != nil {
		return nil, err
	}

	tokenString, err := CreateToken(uuid)
	if err != nil {
		return nil, err
	}

	err = AddJwtTokenToUser(uuid, tokenString)
	if err != nil {
		return nil, err
	}

	return &pb.LoginReply{Token: tokenString}, nil
}

func (s *server) ActivateEmail(_ context.Context, req *pb.ActivateEmailRequest) (*pb.ActivateEmailReply, error) {
	verificationToken := req.Token
	err := VerifyEmailByToken(verificationToken, time.Now())
	if err != nil {
		return nil, err
	}
	return &pb.ActivateEmailReply{}, nil
}

func (s *server) DropSingleToken(_ context.Context, req *pb.DropSingleTokenRequest) (*pb.DropSingleTokenReply, error) {
	jwtToken := req.Token
	return &pb.DropSingleTokenReply{}, DropJwtToken(jwtToken)
}

func (s *server) DropAllTokens(_ context.Context, req *pb.DropAllTokensRequest) (*pb.DropAllTokensReply, error) {
	jwtToken := req.Token
	parsedToken, err := DecodeToken(jwtToken)
	if err != nil {
		return nil, err
	}
	return &pb.DropAllTokensReply{}, DropAllTokensForUuid(parsedToken.Uuid)
}
