package pg

import (
	"context"
	"database/sql"
	"time"

	"final/ledger/internal/domain"
)

type ExpenseRepo struct {
	db *sql.DB
}

func NewExpenseRepo(db *sql.DB) *ExpenseRepo {
	return &ExpenseRepo{db: db}
}

func (r *ExpenseRepo) Insert(ctx context.Context, t domain.Transaction) (int, error) {
	dateOnly := time.Date(t.Date.Year(), t.Date.Month(), t.Date.Day(), 0, 0, 0, 0, time.UTC)

	var id int
	err := r.db.QueryRowContext(ctx,
		`INSERT INTO expenses(amount, category, description, date)
		 VALUES($1,$2,$3,$4)
		 RETURNING id`,
		t.Amount, t.Category, t.Description, dateOnly,
	).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (r *ExpenseRepo) List(ctx context.Context) ([]domain.Transaction, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, amount, category, description, date
		 FROM expenses
		 ORDER BY date DESC, id DESC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]domain.Transaction, 0)
	for rows.Next() {
		var t domain.Transaction
		if err := rows.Scan(&t.ID, &t.Amount, &t.Category, &t.Description, &t.Date); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func (r *ExpenseRepo) SumByCategory(ctx context.Context, category string) (float64, error) {
	var sum float64
	if err := r.db.QueryRowContext(ctx,
		`SELECT COALESCE(SUM(amount),0) FROM expenses WHERE category=$1`,
		category,
	).Scan(&sum); err != nil {
		return 0, err
	}
	return sum, nil
}

func (r *ExpenseRepo) ListCategoriesInRange(ctx context.Context, from, to time.Time) ([]string, error) {
	fromD := time.Date(from.Year(), from.Month(), from.Day(), 0, 0, 0, 0, time.UTC)
	toD := time.Date(to.Year(), to.Month(), to.Day(), 0, 0, 0, 0, time.UTC)

	rows, err := r.db.QueryContext(ctx,
		`SELECT DISTINCT category
		 FROM expenses
		 WHERE date >= $1 AND date <= $2
		 ORDER BY category`,
		fromD, toD,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]string, 0)
	for rows.Next() {
		var c string
		if err := rows.Scan(&c); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func (r *ExpenseRepo) SumByCategoryInRange(ctx context.Context, category string, from, to time.Time) (float64, error) {
	fromD := time.Date(from.Year(), from.Month(), from.Day(), 0, 0, 0, 0, time.UTC)
	toD := time.Date(to.Year(), to.Month(), to.Day(), 0, 0, 0, 0, time.UTC)

	var sum float64
	if err := r.db.QueryRowContext(ctx,
		`SELECT COALESCE(SUM(amount),0)
		 FROM expenses
		 WHERE category=$1 AND date >= $2 AND date <= $3`,
		category, fromD, toD,
	).Scan(&sum); err != nil {
		return 0, err
	}
	return sum, nil
}
