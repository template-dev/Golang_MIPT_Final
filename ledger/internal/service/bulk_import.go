package service

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"

	"final/ledger/internal/domain"
)

func (a *App) BulkImportTransactions(ctx context.Context, items []domain.ImportItem, workers int) (domain.ImportSummary, error) {
	if workers <= 0 {
		workers = 4
	}
	if workers > 64 {
		workers = 64
	}

	type job struct {
		item domain.ImportItem
	}

	type result struct {
		index int
		err   error
	}

	jobs := make(chan job)
	results := make(chan result, workers)

	var wg sync.WaitGroup
	wg.Add(workers)

	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case j, ok := <-jobs:
					if !ok {
						return
					}
					tx := j.item.Tx
					tx.Category = domain.NormalizeCategory(tx.Category)
					_, err := a.AddTransaction(ctx, tx)
					select {
					case <-ctx.Done():
						return
					case results <- result{index: j.item.Index, err: err}:
					}
				}
			}
		}()
	}

	go func() {
		defer close(jobs)
		for _, it := range items {
			select {
			case <-ctx.Done():
				return
			case jobs <- job{item: it}:
			}
		}
	}()

	go func() {
		wg.Wait()
		close(results)
	}()

	var accepted int64
	var rejected int64
	errs := make([]domain.ImportError, 0)

	done := make(chan struct{})
	go func() {
		defer close(done)
		for r := range results {
			if r.err == nil {
				atomic.AddInt64(&accepted, 1)
				continue
			}
			atomic.AddInt64(&rejected, 1)
			errs = append(errs, domain.ImportError{Index: r.index, Error: r.err.Error()})
		}
	}()

	select {
	case <-ctx.Done():
		<-done
		return domain.ImportSummary{
			Accepted: accepted,
			Rejected: rejected,
			Errors:   errs,
		}, ctx.Err()
	case <-done:
	}

	s := domain.ImportSummary{
		Accepted: accepted,
		Rejected: rejected,
		Errors:   errs,
	}

	if ctx.Err() != nil {
		return s, ctx.Err()
	}

	return s, nil
}

var ErrImportCanceled = errors.New("import canceled")
