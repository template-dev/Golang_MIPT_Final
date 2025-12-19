package ledger

import (
	"context"
	"errors"
)

func Reset() {
	db := DB()
	if db == nil {
		return
	}

	_, _ = db.ExecContext(context.Background(), "DELETE FROM expenses")
	_, _ = db.ExecContext(context.Background(), "DELETE FROM budgets")

	r := Redis()
	if r != nil {
		_ = r.FlushDB(context.Background()).Err()
	}
}

var ErrNotInitialized = errors.New("ledger not initialized")
