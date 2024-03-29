package main

import (
	"context"
	"log"
	"net/http"
	"strconv"

	"github.com/dhruvsingh510/bond_social_api/internal/handler"
	"github.com/dhruvsingh510/bond_social_api/internal/service"
	"github.com/hako/branca"
	"github.com/jackc/pgx/v4"
)

const (
	databaseURL = "postgres://postgres:admin@localhost:5432/postgres"
	port        = 8080
	// this key should be set as env variable
	tokenKey = "supersecretkeyyoushouldnotcommit"
)

func main() {
	ctx := context.Background()
	db, err := pgx.Connect(ctx, databaseURL)

	if err != nil {
		log.Fatalf("could not open db connection: %v\n", err)
		return
	}

	if err = db.Ping(ctx); err != nil {
		log.Fatalf("could not ping to db: %v\n", err)
		return
	}

	codec := branca.NewBranca(tokenKey)
	codec.SetTTL(uint32(service.TokenLifespan.Seconds()))

	s := &service.Service{
		Db:    db,
		Codec: codec,
	}

	h := handler.New(s)

	log.Printf("accepting connections on port %d\n", port)

	if err = http.ListenAndServe(":"+strconv.Itoa(port), h); err != nil {
		log.Fatalf("could not start server: %v\n", err)
	}
}
