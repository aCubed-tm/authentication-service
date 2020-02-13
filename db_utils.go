package main

import (
	"github.com/neo4j/neo4j-go-driver/neo4j"
	"log"
)

func newSessionRead() neo4j.Session  { return newSession(neo4j.AccessModeRead) }
func newSessionWrite() neo4j.Session { return newSession(neo4j.AccessModeWrite) }

func FetchSingle(query string, variables map[string]interface{}) (interface{}, error) {
	return newSessionRead().ReadTransaction(func(transaction neo4j.Transaction) (interface{}, error) {
		result, err := transaction.Run(query, variables)
		if err != nil {
			return nil, err
		}

		if result.Next() {
			val := result.Record().GetByIndex(0)

			// make sure we only got 1 result
			if result.Next() {
				log.Printf("WARN: got multiple results running '%s' query", query)
			}

			return val, nil
		}

		return nil, result.Err()
	})
}

func Fetch(query string, variables map[string]interface{}, fetchFn func(tx neo4j.Result) (interface{}, error)) (interface{}, error) {
	return newSessionRead().ReadTransaction(func(transaction neo4j.Transaction) (interface{}, error) {
		result, err := transaction.Run(query, variables)
		if err != nil {
			return nil, err
		}

		return fetchFn(result)
	})
}

func FetchStringArray(query string, variables map[string]interface{}) ([]string, error) {
	ret, err := Fetch(query, variables, func(result neo4j.Result) (interface{}, error) {
		var ret []string
		for result.Next() {
			rec := result.Record()
			ret = append(ret, rec.GetByIndex(0).(string))
		}

		return ret, nil
	})
	if err != nil {
		return nil, err
	}

	return ret.([]string), nil
}

func Write(query string, variables map[string]interface{}) error {
	_, err := newSessionWrite().WriteTransaction(func(transaction neo4j.Transaction) (interface{}, error) {
		_, err := transaction.Run(query, variables)
		return nil, err
	})
	return err
}
