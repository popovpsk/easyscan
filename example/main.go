package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v4/pgxpool"

	"github.com/popovpsk/easyscan"
)

type dbClient struct {
	*pgxpool.Pool
}

func (db *dbClient) Select(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	return easyscan.Select(ctx, db, dest, query, args...)
}

func (db *dbClient) Get(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	return easyscan.Get(ctx, db, dest, query, args...)
}

func newDbClient(ctx context.Context, connString string) (*dbClient, error) {
	pool, err := pgxpool.Connect(ctx, connString)
	if err != nil {
		return nil, err
	}
	err = pool.Ping(ctx)
	if err != nil {
		return nil, err
	}

	return &dbClient{Pool: pool}, nil
}

// you can find it in docker-compose.yml
const connString = "user=postgres password=postgres host=localhost dbname=easyscan port=5432"

type sqlImplementationInfo struct {
	ImplInfoID     string  `db:"implementation_info_id"`
	ImplInfoName   string  `db:"implementation_info_name"`
	IntegerValue   *int    `db:"integer_value"`
	CharacterValue *string `db:"character_value"`
	Comments       *string `db:"comments"`
}

func main() {
	ctx := context.Background()

	client, err := newDbClient(ctx, connString)
	if err != nil {
		panic(err)
	}
	defer client.Close()

	result := make([]sqlImplementationInfo, 0)
	const selectQuery = "SELECT * FROM information_schema.sql_implementation_info"
	err = client.Select(ctx, &result, selectQuery)
	if err != nil {
		panic(err)
	}
	fmt.Print("\n\nSELECT result:\n")
	for i, r := range result {
		js, _ := json.Marshal(r)
		fmt.Println(i, string(js))
	}

	var getResult sqlImplementationInfo
	const getQuery = "SELECT * FROM information_schema.sql_implementation_info WHERE implementation_info_id = $1"
	err = client.Get(ctx, &getResult, getQuery, "26")
	if err != nil {
		panic(err)
	}
	fmt.Print("\n\nGET result:\n")
	js, _ := json.Marshal(getResult)
	fmt.Println(string(js))
}
