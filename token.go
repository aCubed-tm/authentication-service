package main

import (
	"github.com/dgrijalva/jwt-go"
	"log"
	"time"
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

func _(token string) bool { // CheckToken, temporarily renamed to mute compiler warning
	parsedToken, err := jwt.ParseWithClaims(token, &JwtClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(jwtSecret), nil
	})

	if err != nil {
		log.Printf("Failed to parse token: %v", err)
		return false
	}

	if _, ok := parsedToken.Claims.(*JwtClaims); ok {
		return parsedToken.Valid && parsedToken.Method.Alg() == "HS256"
	} else {
		return false
	}
}
