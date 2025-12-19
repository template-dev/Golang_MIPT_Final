package ledger

import (
	"context"
	"database/sql"
	"log"
	"sync"

	"final/ledger/internal/cache"
	"final/ledger/internal/db"

	"github.com/redis/go-redis/v9"
)

var (
	stateMu sync.RWMutex
	pg      *sql.DB
	rdb     *redis.Client
)

func Init(ctx context.Context) error {
	dsn := db.BuildDSNFromEnv()

	conn, err := db.Open(dsn)
	if err != nil {
		return err
	}

	client, err := cache.NewClient()
	if err != nil {
		log.Printf("redis disabled: %v", err)
		client = nil
	} else {
		log.Printf("redis connected")
	}

	stateMu.Lock()
	pg = conn
	rdb = client
	stateMu.Unlock()

	log.Printf("postgres connected")
	return nil
}

func DB() *sql.DB {
	stateMu.RLock()
	defer stateMu.RUnlock()
	return pg
}

func Redis() *redis.Client {
	stateMu.RLock()
	defer stateMu.RUnlock()
	return rdb
}
