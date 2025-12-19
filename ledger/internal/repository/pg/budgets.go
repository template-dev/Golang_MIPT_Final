package pg

import (
	"context"
	"database/sql"

	"final/ledger/internal/domain"
)

type BudgetRepo struct {
	db *sql.DB
}

func NewBudgetRepo(db *sql.DB) *BudgetRepo {
	return &BudgetRepo{db: db}
}

func (r *BudgetRepo) Upsert(ctx context.Context, b domain.Budget) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO budgets(category, limit_amount)
		 VALUES($1,$2)
		 ON CONFLICT(category) DO UPDATE SET limit_amount=EXCLUDED.limit_amount`,
		b.Category, b.Limit,
	)
	return err
}

func (r *BudgetRepo) GetLimit(ctx context.Context, category string) (float64, bool, error) {
	var lim float64
	err := r.db.QueryRowContext(ctx, `SELECT limit_amount FROM budgets WHERE category=$1`, category).Scan(&lim)
	if err == sql.ErrNoRows {
		return 0, false, nil
	}
	if err != nil {
		return 0, false, err
	}
	return lim, true, nil
}

func (r *BudgetRepo) List(ctx context.Context) ([]domain.Budget, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT category, limit_amount FROM budgets ORDER BY category`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]domain.Budget, 0)
	for rows.Next() {
		var cat string
		var lim float64
		if err := rows.Scan(&cat, &lim); err != nil {
			return nil, err
		}
		out = append(out, domain.Budget{Category: cat, Limit: lim, Period: "fixed"})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}
