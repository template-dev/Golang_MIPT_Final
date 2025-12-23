package app

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"os"
	"strings"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	"final/ledger/internal/cache"
	"final/ledger/internal/repository/pg"
	"final/ledger/internal/service"
)

func Build(ctx context.Context) (*service.Service, func() error, error) {
	dsn := strings.TrimSpace(os.Getenv("DATABASE_URL"))
	if dsn == "" {
		return nil, nil, errors.New("DATABASE_URL is required")
	}

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, nil, err
	}

	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(30 * time.Minute)

	pingCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	if err := db.PingContext(pingCtx); err != nil {
		_ = db.Close()
		return nil, nil, err
	}
	log.Printf("[ledger] postgres connected")

	// Redis
	cacheClient, cacheClose, err := cache.NewFromEnv(ctx)
	if err != nil {
		_ = db.Close()
		return nil, nil, err
	}

	// Repos (подстрой под свои имена/пакеты)
	budgetsRepo := pg.NewBudgetRepo(db)
	txsRepo := pg.NewExpenseRepo(db)

	svc := service.New(service.Deps{
		Budgets:      budgetsRepo,
		Transactions: txsRepo,
		Cache:        cacheClient,
	})

	closeFn := func() error {
		_ = cacheClose()
		return db.Close()
	}

	return svc, closeFn, nil
}
