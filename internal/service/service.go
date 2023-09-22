package service

import (
	"sync"

	"github.com/hako/branca"
	"github.com/jackc/pgx/v4"
)

// Service contains the main logic.
type Service struct {
	Db *pgx.Conn
	Codec *branca.Branca
	timelineItemClients sync.Map
}
