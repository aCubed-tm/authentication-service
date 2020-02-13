package main

import (
	"errors"
	"time"

	"github.com/dgrijalva/jwt-go"
)

const jwtSecret = "change me!" // TODO: move to secret!

type JwtClaims struct {
	Uuid string `json:"uuid"`
	jwt.StandardClaims
}

func CreateToken(uuid string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, JwtClaims{
		Uuid: uuid,
		StandardClaims: jwt.StandardClaims{
			IssuedAt: time.Now().Unix(),
		},
	})
	return token.SignedString([]byte(jwtSecret))
}

func DecodeToken(token string) (*JwtClaims, error) {
	parsedToken, err := jwt.ParseWithClaims(token, &JwtClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(jwtSecret), nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := parsedToken.Claims.(*JwtClaims)
	if !ok {
		return nil, errors.New("failed to parse token into JwtClaims")
	}

	if !parsedToken.Valid || parsedToken.Method.Alg() != "HS256" {
		return nil, errors.New("failed to verify token")
	}

	return claims, nil
}
