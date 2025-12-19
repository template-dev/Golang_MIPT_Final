package api

import (
	"time"

	"final/ledger"
)

func ToLedgerTransaction(r CreateTransactionRequest) (ledger.Transaction, error) {
	var t time.Time
	if r.Date == "" {
		t = time.Now()
	} else {
		parsed, err := time.Parse(time.RFC3339, r.Date)
		if err != nil {
			return ledger.Transaction{}, err
		}
		t = parsed
	}

	return ledger.Transaction{
		Amount:      r.Amount,
		Category:    r.Category,
		Description: r.Description,
		Date:        t,
	}, nil
}

func ToTransactionResponse(tx ledger.Transaction) TransactionResponse {
	return TransactionResponse{
		ID:          tx.ID,
		Amount:      tx.Amount,
		Category:    tx.Category,
		Description: tx.Description,
		Date:        tx.Date.Format(time.RFC3339),
	}
}

func ToLedgerBudget(r CreateBudgetRequest) ledger.Budget {
	return ledger.Budget{
		Category: r.Category,
		Limit:    r.Limit,
		Period:   "fixed",
	}
}

func ToBudgetResponse(b ledger.Budget) BudgetResponse {
	return BudgetResponse{
		Category: b.Category,
		Limit:    b.Limit,
		Period:   b.Period,
	}
}
