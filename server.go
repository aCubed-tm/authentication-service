package main

import (
	"context"
	"errors"
	"fmt"
	googleUuid "github.com/google/uuid"
	"log"
	"time"

	pb "github.com/acubed-tm/authentication-service/protofiles"
	"golang.org/x/crypto/bcrypt"
)

const passwordCost = 12

type server struct{}

func (s *server) GetUuidFromToken(_ context.Context, req *pb.GetUuidFromTokenRequest) (*pb.GetUuidFromTokenReply, error) {
	decoded, err := DecodeToken(req.Token)
	if err != nil {
		return nil, err
	}
	return &pb.GetUuidFromTokenReply{Uuid: decoded.Uuid}, nil
}

func (*server) IsEmailRegistered(_ context.Context, req *pb.IsEmailRegisteredRequest) (*pb.IsEmailRegisteredReply, error) {
	uuid, err := GetUuidByEmail(req.Email)
	if err != nil {
		return nil, err
	} else {
		return &pb.IsEmailRegisteredReply{IsRegistered: uuid != "", AccountUuid: uuid}, nil
	}
}

func (s *server) GetInvitesByEmail(_ context.Context, req *pb.GetInvitesByEmailRequest) (*pb.GetInvitesByEmailReply, error) {
	email := req.Email
	ret, err := GetInviteOrganizationsByEmail(email)
	if err != nil {
		return nil, err
	}

	return &pb.GetInvitesByEmailReply{
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

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), passwordCost)
	if err != nil {
		return nil, err
	}

	err = CreateAccount(email, string(hashedPassword), googleUuid.New().String())

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

func (s *server) MakeEmailPrimary(_ context.Context, req *pb.MakeEmailPrimaryRequest) (*pb.MakeEmailPrimaryReply, error) {
	emailUuid := req.EmailUuid
	err := SetNewPrimaryEmail(emailUuid)
	if err != nil {
		return nil, err
	}
	return &pb.MakeEmailPrimaryReply{}, nil
}

func (s *server) AddEmail(_ context.Context, req *pb.AddEmailRequest) (*pb.AddEmailReply, error) {
	email := req.Email
	exists, err := CheckEmailExists(email)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("email already exists")
	}

	// actually add mail
	newEmailUuid := googleUuid.New().String()
	verificationToken := googleUuid.New().String()
	err = AddEmailToUser(req.AccountUuid, newEmailUuid, email, verificationToken)
	if err != nil {
		return nil, err
	}
	return &pb.AddEmailReply{
		EmailUuid:         newEmailUuid,
		VerificationToken: verificationToken,
	}, nil
}

func (s *server) DeleteEmail(_ context.Context, req *pb.DeleteEmailRequest) (*pb.DeleteEmailReply, error) {
	isPrimary, err := CheckEmailPrimary(req.Uuid)
	if err != nil {
		return nil, err
	}
	if isPrimary {
		return nil, errors.New("cannot delete primary email")
	}

	err = DeleteEmail(req.Uuid)
	if err != nil {
		return nil, err
	}

	return &pb.DeleteEmailReply{}, nil
}
