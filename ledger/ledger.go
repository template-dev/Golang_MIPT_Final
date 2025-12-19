package ledger

import (
	"context"

	"final/ledger/internal/app"
	"final/ledger/internal/domain"
	"final/ledger/internal/service"
)

type Service = service.Service

type Transaction = domain.Transaction
type Budget = domain.Budget

type ImportItem = domain.ImportItem
type ImportSummary = domain.ImportSummary
type ImportError = domain.ImportError

var ErrBudgetExceeded = service.ErrBudgetExceeded

func New(ctx context.Context) (Service, func() error, error) {
	return app.Build(ctx)
}
