package service

import (
	"context"
	"errors"
	"time"

	"final/ledger/internal/domain"
)

type Deps struct {
	Budgets      domain.BudgetRepo
	Transactions domain.Transaction
	Cache        Cache
}

var ErrBudgetExceeded = errors.New("budget exceeded")

type Service interface {
	SetBudget(ctx context.Context, b domain.Budget) (domain.Budget, error)
	ListBudgets(ctx context.Context) ([]domain.Budget, error)

	AddTransaction(ctx context.Context, t domain.Transaction) (domain.Transaction, error)
	ListTransactions(ctx context.Context) ([]domain.Transaction, error)

	ReportSummary(ctx context.Context, from, to time.Time) (map[string]float64, error)
	BulkImportTransactions(ctx context.Context, items []domain.ImportItem, workers int) (domain.ImportSummary, error)
}

type App struct {
	budgets  domain.BudgetRepo
	expenses domain.ExpenseRepo
}

func New(b domain.BudgetRepo, e domain.ExpenseRepo) *App {
	return &App{budgets: b, expenses: e}
}

func (a *App) SetBudget(ctx context.Context, b domain.Budget) (domain.Budget, error) {
	if err := b.Validate(); err != nil {
		return domain.Budget{}, err
	}
	b.Category = domain.NormalizeCategory(b.Category)
	if b.Period == "" {
		b.Period = "fixed"
	}
	if err := a.budgets.Upsert(ctx, b); err != nil {
		return domain.Budget{}, err
	}
	return b, nil
}

func (a *App) ListBudgets(ctx context.Context) ([]domain.Budget, error) {
	return a.budgets.List(ctx)
}

func (a *App) AddTransaction(ctx context.Context, t domain.Transaction) (domain.Transaction, error) {
	if err := t.Validate(); err != nil {
		return domain.Transaction{}, err
	}

	t.Category = domain.NormalizeCategory(t.Category)

	limit, hasBudget, err := a.budgets.GetLimit(ctx, t.Category)
	if err != nil {
		return domain.Transaction{}, err
	}

	if hasBudget {
		spent, err := a.expenses.SumByCategory(ctx, t.Category)
		if err != nil {
			return domain.Transaction{}, err
		}
		if spent+t.Amount > limit {
			return domain.Transaction{}, ErrBudgetExceeded
		}
	}

	id, err := a.expenses.Insert(ctx, t)
	if err != nil {
		return domain.Transaction{}, err
	}

	t.ID = id
	return t, nil
}

func (a *App) ListTransactions(ctx context.Context) ([]domain.Transaction, error) {
	return a.expenses.List(ctx)
}
