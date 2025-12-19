package ledger

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"
)

func SetBudget(b Budget) (Budget, error) {
	if err := b.Validate(); err != nil {
		return Budget{}, err
	}

	b.Category = normalizeCategory(b.Category)
	if b.Period == "" {
		b.Period = "fixed"
	}

	db := DB()
	if db == nil {
		return Budget{}, ErrNotInitialized
	}

	_, err := db.ExecContext(context.Background(),
		`INSERT INTO budgets(category, limit_amount)
		 VALUES($1,$2)
		 ON CONFLICT(category) DO UPDATE SET limit_amount=EXCLUDED.limit_amount`,
		b.Category, b.Limit,
	)
	if err != nil {
		return Budget{}, err
	}

	r := Redis()
	if r != nil {
		_ = r.Del(context.Background(), "budgets:all").Err()
	}

	return b, nil
}

func ListBudgets() ([]Budget, error) {
	db := DB()
	if db == nil {
		return nil, ErrNotInitialized
	}

	r := Redis()
	if r != nil {
		if s, err := r.Get(context.Background(), "budgets:all").Result(); err == nil {
			var cached []Budget
			if json.Unmarshal([]byte(s), &cached) == nil {
				return cached, nil
			}
		}
	}

	rows, err := db.QueryContext(context.Background(),
		`SELECT category, limit_amount FROM budgets ORDER BY category`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]Budget, 0)
	for rows.Next() {
		var cat string
		var lim float64
		if err := rows.Scan(&cat, &lim); err != nil {
			return nil, err
		}
		out = append(out, Budget{Category: cat, Limit: lim, Period: "fixed"})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	if r != nil {
		if b, err := json.Marshal(out); err == nil {
			_ = r.Set(context.Background(), "budgets:all", string(b), 20*time.Second).Err()
		}
	}

	return out, nil
}

func getBudgetLimit(ctx context.Context, db *sql.DB, category string) (float64, bool, error) {
	var lim float64
	err := db.QueryRowContext(ctx, `SELECT limit_amount FROM budgets WHERE category=$1`, category).Scan(&lim)
	if err == sql.ErrNoRows {
		return 0, false, nil
	}
	if err != nil {
		return 0, false, err
	}
	return lim, true, nil
}
