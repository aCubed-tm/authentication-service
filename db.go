package main

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/dgraph-io/dgo/v2"
	"github.com/dgraph-io/dgo/v2/protos/api"
	"google.golang.org/grpc"
	"log"
	"time"
)

const databaseUrl = "dgraph-public:9080"

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

func VerifyEmailByToken(ctx context.Context, token string, verificationTime time.Time) error {
	c := newClient()

	variables := map[string]string{"$token": token}
	q := `
		query x($token: string){
			email(func: eq(verificationToken, $token)) {
				uid
				emailAddress
				verificationToken
				verifiedAt
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
			Uid        string    `json:"uid"`
			Address    string    `json:"emailAddress"`
			Token      string    `json:"verificationToken"`
			VerifiedAt time.Time `json:"verifiedAt"`
		} `json:"email"`
	}
	log.Println("JSON: " + string(resp.GetJson()))
	if err := json.Unmarshal(resp.GetJson(), &decode); err != nil {
		return err
	}

	if len(decode.All) == 0 {
		return errors.New("couldn't find token")
	}

	decode.All[0].VerifiedAt = verificationTime
	out, err := json.Marshal(decode.All[0])
	if err != nil {
		return err
	}

	_, err = txn.Mutate(context.Background(), &api.Mutation{SetJson: out, CommitNow: true})
	if err != nil {
		return err
	}

	return nil
}

func GetAllEmailsByUuid(ctx context.Context, uuid string) ([]string, error) {
	c := newClient()

	variables := map[string]string{"$uuid": uuid}
	q := `
		query x($uuid: string){
			account(func: eq(uuid, $uuid)) {
				uuid
				email {
					emailAddress
				}
			}
		}
	`

	resp, err := c.NewTxn().QueryWithVars(ctx, q, variables)
	if err != nil {
		return nil, err
	}

	var decode struct {
		All []struct {
			Uuid  string `json:"uuid"`
			Email []struct {
				Address string `json:"emailAddress"`
			} `json:"email"`
		} `json:"account"`
	}
	log.Println("JSON: " + string(resp.GetJson()))
	if err := json.Unmarshal(resp.GetJson(), &decode); err != nil {
		return nil, err
	}

	if len(decode.All) == 0 {
		return nil, errors.New("couldn't find uuid")
	}

	ret := make([]string, len(decode.All[0].Email))
	for i, v := range decode.All[0].Email {
		ret[i] = v.Address
	}
	return ret, nil
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

func GetInviteOrganizationsByEmail(ctx context.Context, email string) ([]string, error) {
	c := newClient()

	variables := map[string]string{"$email": email}
	q := `
		query x($email: string){
			email(func: eq(emailAddress, $email)) {
				emailAddress
				~email { // invite
					organisation {
						uuid
					}
				}
			}
		}
	`

	resp, err := c.NewTxn().QueryWithVars(ctx, q, variables)
	if err != nil {
		return nil, err
	}

	var decode struct {
		All []struct {
			Address string `json:"emailAddress"`
			Invite  []struct {
				Organisation struct {
					Uuid string `json:"uuid"`
				}
			} `json:"~email"`
		} `json:"email"`
	}
	log.Println("JSON: " + string(resp.GetJson()))
	if err := json.Unmarshal(resp.GetJson(), &decode); err != nil {
		return nil, err
	}

	if len(decode.All) == 0 {
		return nil, errors.New("couldn't find uuid")
	}

	var invites []string
	for _, inv := range decode.All[0].Invite {
		invites = append(invites, inv.Organisation.Uuid)
	}
	return invites, nil
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

// Remove tokens from database. Removes all tokens if token parameter is "-"
func RemoveJwtTokenByToken(ctx context.Context, token string, dropAll bool) error {
	c := newClient()

	variables := map[string]string{"$token": token}
	q := `
		query x($token: string) {
			user {
				uid
				JWTToken(func: eq(token, $token)) {
					uid
					token
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
		Users []struct {
			Uid       string `json:"uid"`
			JWTTokens []struct {
				Uid   string `json:"uid"`
				Token string `json:"token"`
			} `json:"JWTToken"`
		} `json:"user"`
	}
	log.Println("JSON: " + string(resp.GetJson()))
	if err := json.Unmarshal(resp.GetJson(), &decode); err != nil {
		return err
	}

	if len(decode.Users) == 0 {
		return errors.New("couldn't find users")
	}

	// note: should remove token nodes
	tokens := decode.Users[0].JWTTokens

	if !dropAll {
		var idx int
		for i, t := range tokens {
			if t.Token == token {
				idx = i
				break
			}
		}
		tokens = append(tokens[:idx], tokens[idx+1:]...)
	} else {
		tokens = nil
	}
	decode.Users[0].JWTTokens = tokens

	out, err := json.Marshal(decode.Users[0])
	if err != nil {
		return err
	}

	_, err = txn.Mutate(context.Background(), &api.Mutation{SetJson: out, CommitNow: true})
	if err != nil {
		return err
	}

	return nil
}

// Remove tokens from database. Removes all tokens if token parameter is "-"
func AddJwtTokenToUser(ctx context.Context, uuid string, token string) error {
	c := newClient()

	variables := map[string]string{"$uuid": uuid}
	q := `
		query x($uuid: string) {
			user(func: eq(uuid, $uuid)) {
				uid
				JWTToken {
					uid
					token
				}
			}
		}
	`

	txn := c.NewTxn()
	resp, err := txn.QueryWithVars(ctx, q, variables)
	if err != nil {
		return err
	}

	type TokenData struct {
		Uid   string `json:"uid"`
		Token string `json:"token"`
	}
	var decode struct {
		Users []struct {
			Uid       string      `json:"uid"`
			JWTTokens []TokenData `json:"JWTToken"`
		} `json:"user"`
	}
	log.Println("JSON: " + string(resp.GetJson()))
	if err := json.Unmarshal(resp.GetJson(), &decode); err != nil {
		return err
	}

	if len(decode.Users) == 0 {
		return errors.New("couldn't find users")
	}

	decode.Users[0].JWTTokens = append(decode.Users[0].JWTTokens, TokenData{
		Token: token,
	})

	out, err := json.Marshal(decode.Users[0])
	if err != nil {
		return err
	}

	_, err = txn.Mutate(context.Background(), &api.Mutation{SetJson: out, CommitNow: true})
	if err != nil {
		return err
	}

	return nil
}
