package main

import (
	"time"

	"github.com/neo4j/neo4j-go-driver/neo4j"
)

const databaseUrl = "neo4j-public.default:7687"
const username = "neo4j"
const password = ""

var driver neo4j.Driver = nil

func newSessionRead() neo4j.Session  { return newSession(neo4j.AccessModeRead) }
func newSessionWrite() neo4j.Session { return newSession(neo4j.AccessModeWrite) }

// e.g. neo4j.AccessModeWrite
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

func CheckEmailExists(email string) (bool, error) {
	q := "MATCH (e:Email {emailAddress: $email}) RETURN COUNT(e)"
	variables := map[string]interface{}{"email": email}

	count, err := newSessionRead().ReadTransaction(func(transaction neo4j.Transaction) (interface{}, error) {
		result, err := transaction.Run(q, variables)
		if err != nil {
			return nil, err
		}

		if result.Next() {
			return result.Record().GetByIndex(0), nil
		}

		return nil, result.Err()
	})
	if err != nil {
		return false, err
	}

	return count.(int) > 0, nil
}

func VerifyEmailByToken(token string, verificationTime time.Time) error {
	q := "MATCH (e:Email {verificationToken: $token}) SET e.verifiedAt = $date"
	variables := map[string]interface{}{"token": token, "date": verificationTime}

	_, err := newSessionWrite().WriteTransaction(func(transaction neo4j.Transaction) (interface{}, error) {
		_, err := transaction.Run(q, variables)
		return nil, err
	})
	return err
}

func GetAllEmailsByUuid(uuid string) ([]string, error) {
	q := "MATCH (:Account {uuid: $uuid})--(e:Email) RETURN e.emailAddress"
	variables := map[string]interface{}{"uuid": uuid}

	result, err := newSessionRead().ReadTransaction(func(transaction neo4j.Transaction) (interface{}, error) {
		result, err := transaction.Run(q, variables)
		if err != nil {
			return nil, err
		}

		var ret []string
		for result.Next() {
			rec := result.Record()
			ret = append(ret, rec.GetByIndex(0).(string))
		}

		return ret, nil
	})
	return result.([]string), err
}

func GetPasswordByEmail(email string) (string, error) {
	q := "MATCH (:Email {emailAddress: $email})--(a:Account) RETURN a.password"
	variables := map[string]interface{}{"email": email}

	password, err := newSessionRead().ReadTransaction(func(transaction neo4j.Transaction) (interface{}, error) {
		result, err := transaction.Run(q, variables)
		if err != nil {
			return nil, err
		}

		if result.Next() {
			return result.Record().GetByIndex(0), nil
		}

		return nil, result.Err()
	})
	return password.(string), err
}

func GetInviteOrganizationsByEmail(email string) ([]string, error) {
	q := "MATCH (:Email{emailAddress: $email})<-[:INVITED]-(o:Organisation) RETURN o.uuid"
	variables := map[string]interface{}{"email": email}
	invites, err := newSessionRead().ReadTransaction(func(transaction neo4j.Transaction) (interface{}, error) {
		result, err := transaction.Run(q, variables)
		if err != nil {
			return nil, err
		}

		var ret []string
		for result.Next() {
			rec := result.Record()
			ret = append(ret, rec.GetByIndex(0).(string))
		}

		return ret, nil
	})
	return invites.([]string), err
}

func GetUuidByEmail(email string) (string, error) {
	q := "MATCH (:Email{emailAddress: $email})<-[:HAS_EMAIL]-(a:Account) RETURN a.uuid"
	variables := map[string]interface{}{"email": email}

	uuid, err := newSessionRead().ReadTransaction(func(transaction neo4j.Transaction) (interface{}, error) {
		result, err := transaction.Run(q, variables)
		if err != nil {
			return nil, err
		}

		if result.Next() {
			return result.Record().GetByIndex(0), nil
		}

		return nil, result.Err()
	})
	return uuid.(string), err
}

func ChangePasswordForEmail(email string, password string) error {
	q := "MATCH (:Email {emailAddress: $email})<-[:HAS_EMAIL]-(a:Account) SET a.password = $password"
	variables := map[string]interface{}{"email": email, "password": password}

	_, err := newSessionWrite().WriteTransaction(func(transaction neo4j.Transaction) (interface{}, error) {
		_, err := transaction.Run(q, variables)
		return nil, err
	})
	return err
}

func DropJwtToken(token string) error {
	q := "MATCH (t:JWTToken{token: $token}) DETACH DELETE t"
	variables := map[string]interface{}{"token": token}

	_, err := newSessionWrite().WriteTransaction(func(transaction neo4j.Transaction) (interface{}, error) {
		_, err := transaction.Run(q, variables)
		return nil, err
	})
	return err
}

func DropAllTokensForUuid(uuid string) error {
	q := "MATCH (:Account {uuid: $uuid})-[:HAS_TOKEN]->(t:JWTToken) DETACH DELETE t"
	variables := map[string]interface{}{"$uuid": uuid}

	_, err := newSessionWrite().WriteTransaction(func(transaction neo4j.Transaction) (interface{}, error) {
		_, err := transaction.Run(q, variables)
		return nil, err
	})
	return err
}

func AddJwtTokenToUser(uuid string, token string) error {
	q := "MATCH (a:Account {uuid: $uuid}) CREATE (a)-[:HAS_TOKEN]->(t:JWTToken{token: $token})"
	variables := map[string]interface{}{"$token": token, "$uuid": uuid}

	_, err := newSessionWrite().WriteTransaction(func(transaction neo4j.Transaction) (interface{}, error) {
		_, err := transaction.Run(q, variables)
		return nil, err
	})
	return err
}
