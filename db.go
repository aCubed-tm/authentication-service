package main

import (
	"time"

	"github.com/neo4j/neo4j-go-driver/neo4j"
)

const databaseUrl = "bolt://neo4j-public.default:7687"
const username = "neo4j"
const password = ""

var driver neo4j.Driver = nil

func newSession(accessMode neo4j.AccessMode) neo4j.Session {
	var (
		err     error
		session neo4j.Session
	)

	if driver == nil {
		driver, err = neo4j.NewDriver(databaseUrl, neo4j.BasicAuth(username, password, ""))
		if err != nil {
			panic(err.Error())
		}
	}

	session, err = driver.Session(accessMode)
	if err != nil {
		panic(err.Error())
	}

	return session
}

func _(email string) (bool, error) { // CheckEmailExists
	query := "MATCH (e:Email {emailAddress: {email}}) RETURN COUNT(e)"
	variables := map[string]interface{}{"email": email}
	count, err := FetchSingle(query, variables)
	if err != nil {
		return false, err
	}
	return count.(int64) > 0, nil
}

func VerifyEmailByToken(token string, verificationTime time.Time) error {
	query := "MATCH (e:Email {verificationToken: {token}}) SET e.verifiedAt = {date}"
	variables := map[string]interface{}{"token": token, "date": verificationTime.Unix()}
	return Write(query, variables)
}

func GetAllEmailsByUuid(uuid string) ([]string, error) {
	query := "MATCH (:Account {uuid: {uuid}})--(e:Email) RETURN e.emailAddress"
	variables := map[string]interface{}{"uuid": uuid}
	return FetchStringArray(query, variables)
}

func GetPasswordByEmail(email string) (string, error) {
	query := "MATCH (:Email {emailAddress: {email}})--(a:Account) RETURN a.password"
	variables := map[string]interface{}{"email": email}
	password, err := FetchSingle(query, variables)
	if err != nil {
		return "", err
	}
	if password == nil { // can't return nil strings
		return "", nil
	}
	return password.(string), nil
}

func GetInviteOrganizationsByEmail(email string) ([]string, error) {
	query := "MATCH (:Email{emailAddress: {email}})<-[:INVITED]-(o:Organisation) RETURN o.uuid"
	variables := map[string]interface{}{"email": email}
	return FetchStringArray(query, variables)
}

func GetUuidByEmail(email string) (string, error) {
	query := "MATCH (:Email{emailAddress: {email}})<-[:HAS_EMAIL]-(a:Account) RETURN a.uuid"
	variables := map[string]interface{}{"email": email}
	uuid, err := FetchSingle(query, variables)
	if err != nil {
		return "", err
	}
	if uuid == nil {
		return "", nil
	}
	return uuid.(string), nil
}

func CreateAccount(email, password, uuid string) error {
	query := "MATCH (x:Email{emailAddress:{email}}) CREATE (x)<-[:HAS_EMAIL]-(:Account{password: {password}, active: true, uuid: {uuid}})"
	variables := map[string]interface{}{"email": email, "password": password, "uuid": uuid}
	return Write(query, variables)
}

func _(email string, password string) error { // ChangePasswordForEmail
	query := "MATCH (:Email {emailAddress: {email}})<-[:HAS_EMAIL]-(a:Account) SET a.password = {password}"
	variables := map[string]interface{}{"email": email, "password": password}
	return Write(query, variables)
}

func DropJwtToken(token string) error {
	query := "MATCH (t:JWTToken{token: {token}}) DETACH DELETE t"
	variables := map[string]interface{}{"token": token}
	return Write(query, variables)
}

func DropAllTokensForUuid(uuid string) error {
	query := "MATCH (:Account {uuid: {uuid}})-[:HAS_TOKEN]->(t:JWTToken) DETACH DELETE t"
	variables := map[string]interface{}{"uuid": uuid}
	return Write(query, variables)
}

func AddJwtTokenToUser(uuid string, token string) error {
	query := "MATCH (a:Account {uuid: {uuid}}) CREATE (a)-[:HAS_TOKEN]->(t:JWTToken{token: {token}})"
	variables := map[string]interface{}{"token": token, "uuid": uuid}
	return Write(query, variables)
}
