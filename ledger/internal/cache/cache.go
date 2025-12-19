package cache

import (
	"context"
	"os"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

func NewClient() (*redis.Client, error) {
	addr := getenv("REDIS_ADDR", "localhost:6379")
	pass := os.Getenv("REDIS_PASSWORD")
	db := getenvInt("REDIS_DB", 0)

	rdb := redis.NewClient(&redis.Options{
		Addr:        addr,
		Password:    pass,
		DB:          db,
		DialTimeout: 2 * time.Second,
	})

	if err := rdb.Ping(context.Background()).Err(); err != nil {
		return nil, err
	}
	return rdb, nil
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func getenvInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if x, err := strconv.Atoi(v); err == nil {
			return x
		}
	}
	return def
}
