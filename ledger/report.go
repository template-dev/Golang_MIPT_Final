package ledger

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

type ReportSummaryRow struct {
	Category string  `json:"category"`
	Total    float64 `json:"total"`
}

func GetReportSummary(ctx context.Context, from, to time.Time) ([]ReportSummaryRow, error) {
	db := DB()
	if db == nil {
		return nil, ErrNotInitialized
	}

	if from.After(to) {
		return nil, errors.New("from must be <= to")
	}

	fromDate := time.Date(from.Year(), from.Month(), from.Day(), 0, 0, 0, 0, time.UTC)
	toDate := time.Date(to.Year(), to.Month(), to.Day(), 0, 0, 0, 0, time.UTC)

	key := fmt.Sprintf("report:summary:%s:%s", fromDate.Format("2006-01-02"), toDate.Format("2006-01-02"))

	r := Redis()
	if r != nil {
		if s, err := r.Get(ctx, key).Result(); err == nil {
			var cached []ReportSummaryRow
			if json.Unmarshal([]byte(s), &cached) == nil {
				return cached, nil
			}
		}
	}

	rows, err := db.QueryContext(ctx,
		`SELECT category, COALESCE(SUM(amount),0) AS total
		 FROM expenses
		 WHERE date >= $1 AND date <= $2
		 GROUP BY category
		 ORDER BY category`,
		fromDate, toDate,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]ReportSummaryRow, 0)
	for rows.Next() {
		var row ReportSummaryRow
		if err := rows.Scan(&row.Category, &row.Total); err != nil {
			return nil, err
		}
		out = append(out, row)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	if r != nil {
		if b, err := json.Marshal(out); err == nil {
			_ = r.Set(ctx, key, string(b), 30*time.Second).Err()
		}
	}

	return out, nil
}

func ensureDB(ctx context.Context, db *sql.DB) error {
	return db.PingContext(ctx)
}
