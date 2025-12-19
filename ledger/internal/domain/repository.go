package domain

import (
	"context"
	"time"
)

type BudgetRepo interface {
	Upsert(ctx context.Context, b Budget) error
	GetLimit(ctx context.Context, category string) (float64, bool, error)
	List(ctx context.Context) ([]Budget, error)
}

type ExpenseRepo interface {
	Insert(ctx context.Context, t Transaction) (int, error)
	List(ctx context.Context) ([]Transaction, error)
	SumByCategory(ctx context.Context, category string) (float64, error)

	ListCategoriesInRange(ctx context.Context, from, to time.Time) ([]string, error)
	SumByCategoryInRange(ctx context.Context, category string, from, to time.Time) (float64, error)
}
