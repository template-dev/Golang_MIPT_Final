package service

import (
	"context"
	"errors"
	"log"
	"sync"
	"time"
)

func (a *App) ReportSummary(ctx context.Context, from, to time.Time) (map[string]float64, error) {
	if from.After(to) {
		return nil, errors.New("from must be <= to")
	}

	cats, err := a.expenses.ListCategoriesInRange(ctx, from, to)
	if err != nil {
		return nil, err
	}

	out := make(map[string]float64, len(cats))
	if len(cats) == 0 {
		return out, nil
	}

	done := make(chan struct{})
	ticker := time.NewTicker(400 * time.Millisecond)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-done:
				return
			case <-ticker.C:
				log.Printf("report summary calculating")
			}
		}
	}()

	type res struct {
		cat string
		sum float64
		err error
	}

	ch := make(chan res, len(cats))
	var wg sync.WaitGroup
	wg.Add(len(cats))

	for _, c := range cats {
		c := c
		go func() {
			defer wg.Done()
			if ctx.Err() != nil {
				ch <- res{cat: c, err: ctx.Err()}
				return
			}
			s, err := a.expenses.SumByCategoryInRange(ctx, c, from, to)
			ch <- res{cat: c, sum: s, err: err}
		}()
	}

	wg.Wait()
	close(done)
	close(ch)

	for r := range ch {
		if r.err != nil {
			if ctx.Err() != nil {
				return nil, ctx.Err()
			}
			return nil, r.err
		}
		out[r.cat] = r.sum
	}

	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	return out, nil
}
