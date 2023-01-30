package service

import (
	"github.com/jackc/pgx/v4"
	"github.com/hako/branca"
)

// Service contains the main logic.
type Service struct {
	Db *pgx.Conn
	Codec *branca.Branca

}

// New service implementation
// func New(db *sql.Db, codec)