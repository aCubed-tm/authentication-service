package main

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/dgraph-io/dgo/v2"
	"github.com/dgraph-io/dgo/v2/protos/api"
	"google.golang.org/grpc"
	"log"
)

const databaseUrl = "my-release-dgraph-alpha.acubed:9080"

func newClient() *dgo.Dgraph {
	// Dial a gRPC connection. The address to dial to can be configured when
	// setting up the dgraph cluster.
	d, err := grpc.Dial(databaseUrl, grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}

	return dgo.NewDgraphClient(
		api.NewDgraphClient(d),
	)
}

func GetEmail(ctx context.Context, email string) (string, error) {
	c := newClient()

	variables := map[string]string{"$email": email}
	q := `
		query x($email: string){
			email(func: eq(emailAddress, $email)) {
				emailAddress
			}
		}
	`

	resp, err := c.NewTxn().QueryWithVars(ctx, q, variables)
	if err != nil {
		return "", err
	}

	var decode struct {
		All []struct {
			Address string `json:"emailAddress"`
		} `json:"email"`
	}
	log.Println("JSON: " + string(resp.GetJson()))
	if err := json.Unmarshal(resp.GetJson(), &decode); err != nil {
		return "", err
	}

	if len(decode.All) == 0 {
		return "", errors.New("couldn't find email")
	}

	return decode.All[0].Address, nil
}

func GetEmailByVerificationToken(ctx context.Context, token string) (string, error) {
	c := newClient()

	variables := map[string]string{"$token": token}
	q := `
		query x($token: string){
			email(func: eq(verificationToken, $token)) {
				emailAddress
				verificationToken
			}
		}
	`

	resp, err := c.NewTxn().QueryWithVars(ctx, q, variables)
	if err != nil {
		return "", err
	}

	var decode struct {
		All []struct {
			Address string `json:"emailAddress"`
			Token   string `json:"verificationToken"`
		} `json:"email"`
	}
	log.Println("JSON: " + string(resp.GetJson()))
	if err := json.Unmarshal(resp.GetJson(), &decode); err != nil {
		return "", err
	}

	if len(decode.All) == 0 {
		return "", errors.New("couldn't find token")
	}

	return decode.All[0].Address, nil
}

func GetPasswordByEmail(ctx context.Context, email string) (string, error) {
	c := newClient()

	variables := map[string]string{"$email": email}
	q := `
		query x($email: string){
			email(func: eq(emailAddress, $email)) {
				emailAddress
				~email {
					password
				}
			}
		}
	`

	resp, err := c.NewTxn().QueryWithVars(ctx, q, variables)
	// defer txn.Discard(ctx)
	// resp, err := txn.Query(context.Background(), q)
	if err != nil {
		return "", err
	}

	// After we get the balances, we have to decode them into structs so that
	// we can manipulate the data.
	var decode struct {
		All []struct {
			Address string `json:"emailAddress"`
			User    []struct {
				Password string `json:"password"`
			} `json:"~email"`
		} `json:"email"`
	}
	log.Println("JSON: " + string(resp.GetJson()))
	if err := json.Unmarshal(resp.GetJson(), &decode); err != nil {
		return "", err
	}

	if len(decode.All) == 0 {
		return "", errors.New("couldn't find email")
	}

	return decode.All[0].User[0].Password, nil
}

func GetUuidByEmail(ctx context.Context, email string) (string, error) {
	c := newClient()

	variables := map[string]string{"$email": email}
	q := `
		query x($email: string){
			email(func: eq(emailAddress, $email)) {
				emailAddress
				~email {
					uuid
				}
			}
		}
	`

	resp, err := c.NewTxn().QueryWithVars(ctx, q, variables)
	if err != nil {
		return "", err
	}

	var decode struct {
		All []struct {
			Address string `json:"emailAddress"`
			User    []struct {
				Uuid string `json:"uuid"`
			} `json:"~email"`
		} `json:"email"`
	}
	log.Println("JSON: " + string(resp.GetJson()))
	if err := json.Unmarshal(resp.GetJson(), &decode); err != nil {
		return "", err
	}

	if len(decode.All) == 0 {
		return "", errors.New("couldn't find email")
	}

	return decode.All[0].User[0].Uuid, nil
}

func ChangePasswordForEmail(ctx context.Context, email string, password string) error {
	c := newClient()

	variables := map[string]string{"$email": email}
	q := `
		query x($email: string){
			email(func: eq(emailAddress, $email)) {
				emailAddress
				~email {
					uid
					password
				}
			}
		}
	`

	txn := c.NewTxn()
	resp, err := txn.QueryWithVars(ctx, q, variables)
	if err != nil {
		return err
	}

	var decode struct {
		All []struct {
			Address string `json:"emailAddress"`
			User    []struct {
				Uid      string `json:"uid"`
				Password string `json:"password"`
			} `json:"~email"`
		} `json:"email"`
	}
	log.Println("JSON: " + string(resp.GetJson()))
	if err := json.Unmarshal(resp.GetJson(), &decode); err != nil {
		return err
	}

	if len(decode.All) == 0 {
		return errors.New("couldn't find email")
	}

	decode.All[0].User[0].Password = password

	out, err := json.Marshal(decode.All[0].User[0])
	if err != nil {
		return err
	}

	_, err = txn.Mutate(context.Background(), &api.Mutation{SetJson: out, CommitNow: true})
	if err != nil {
		return err
	}

	return nil
}
