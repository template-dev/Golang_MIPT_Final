package ledger

import (
	"context"
	"time"
)

func AddTransaction(tx Transaction) (Transaction, error) {
	if err := tx.Validate(); err != nil {
		return Transaction{}, err
	}

	tx.Category = normalizeCategory(tx.Category)

	db := DB()
	if db == nil {
		return Transaction{}, ErrNotInitialized
	}

	ctx := context.Background()

	lim, hasBudget, err := getBudgetLimit(ctx, db, tx.Category)
	if err != nil {
		return Transaction{}, err
	}

	if hasBudget {
		var spent float64
		if err := db.QueryRowContext(ctx,
			`SELECT COALESCE(SUM(amount),0) FROM expenses WHERE category=$1`,
			tx.Category,
		).Scan(&spent); err != nil {
			return Transaction{}, err
		}

		if spent+tx.Amount > lim {
			return Transaction{}, ErrBudgetExceeded
		}
	}

	var id int
	dateOnly := time.Date(tx.Date.Year(), tx.Date.Month(), tx.Date.Day(), 0, 0, 0, 0, time.UTC)

	err = db.QueryRowContext(ctx,
		`INSERT INTO expenses(amount, category, description, date)
		 VALUES($1,$2,$3,$4)
		 RETURNING id`,
		tx.Amount, tx.Category, tx.Description, dateOnly,
	).Scan(&id)
	if err != nil {
		return Transaction{}, err
	}

	tx.ID = id
	tx.Date = dateOnly
	return tx, nil
}

func ListTransactions() ([]Transaction, error) {
	db := DB()
	if db == nil {
		return nil, ErrNotInitialized
	}

	rows, err := db.QueryContext(context.Background(),
		`SELECT id, amount, category, description, date
		 FROM expenses
		 ORDER BY date DESC, id DESC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]Transaction, 0)
	for rows.Next() {
		var tx Transaction
		if err := rows.Scan(&tx.ID, &tx.Amount, &tx.Category, &tx.Description, &tx.Date); err != nil {
			return nil, err
		}
		out = append(out, tx)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return out, nil
}
