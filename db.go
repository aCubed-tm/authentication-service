package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/dgraph-io/dgo/v2"
	"github.com/dgraph-io/dgo/v2/protos/api"
	"google.golang.org/grpc"
	"log"
)

// TODO: should use some discovery mechanism
const databaseUrl = "51.105.216.241:9080"

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

func setup(c *dgo.Dgraph) {
	// Install a schema into dgraph. Accounts have a `name` and a `balance`.
	_ = c.Alter(context.Background(), &api.Operation{
		Schema: `
			name: string @index(term) .
			balance: int .
		`,
	})
}

func GetUserByEmail(ctx context.Context, email string) (string, error) {
	c := newClient()
	txn := c.NewTxn()
	defer txn.Discard(ctx)

	// TODO: injection vuln!!
	q := fmt.Sprintf(`
		{
			account {
				email(func: eq(address, "%v")) {
					address
				}
				password
			}
		}
	`, email)
	resp, err := txn.Query(context.Background(), q)
	if err != nil {
		log.Fatal(err)
	}

	// After we get the balances, we have to decode them into structs so that
	// we can manipulate the data.
	var decode struct {
		All []struct {
			Password string
			Email []struct {
				address string
			}
		}
	}
	if err := json.Unmarshal(resp.GetJson(), &decode); err != nil {
		return "", err
	}

	// TODO: unconditional index
	return decode.All[0].Password, nil
}
