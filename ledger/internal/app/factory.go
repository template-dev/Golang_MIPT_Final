package app

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strconv"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	"final/ledger/internal/repository/pg"
	"final/ledger/internal/service"
)

type Config struct {
	DSN string
}

func FromEnv() Config {
	if v := os.Getenv("DATABASE_URL"); v != "" {
		return Config{DSN: v}
	}

	host := getenv("DB_HOST", "localhost")
	port := getenv("DB_PORT", "5432")
	user := getenv("DB_USER", "postgres")
	pass := getenv("DB_PASS", "postgres")
	name := getenv("DB_NAME", "cashapp")
	ssl := getenv("DB_SSLMODE", "disable")

	return Config{
		DSN: fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s", user, pass, host, port, name, ssl),
	}
}

func Build(ctx context.Context) (service.Service, func() error, error) {
	cfg := FromEnv()

	db, err := sql.Open("pgx", cfg.DSN)
	if err != nil {
		return nil, nil, err
	}

	db.SetMaxOpenConns(getenvInt("DB_MAX_OPEN_CONNS", 10))
	db.SetMaxIdleConns(getenvInt("DB_MAX_IDLE_CONNS", 5))
	db.SetConnMaxLifetime(30 * time.Minute)

	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, nil, err
	}

	bRepo := pg.NewBudgetRepo(db)
	eRepo := pg.NewExpenseRepo(db)

	svc := service.New(bRepo, eRepo)
	closeFn := func() error { return db.Close() }

	return svc, closeFn, nil
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
