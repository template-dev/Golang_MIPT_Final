package domain

import (
	"errors"
	"strings"
	"time"
)

type Validatable interface {
	Validate() error
}

type Transaction struct {
	ID          int
	Amount      float64
	Category    string
	Description string
	Date        time.Time
}

func (t Transaction) Validate() error {
	if t.Amount <= 0 {
		return errors.New("amount must be > 0")
	}
	if strings.TrimSpace(t.Category) == "" {
		return errors.New("transaction category is empty")
	}
	if t.Date.IsZero() {
		return errors.New("date is required")
	}
	return nil
}

type Budget struct {
	Category string
	Limit    float64
	Period   string
}

func (b Budget) Validate() error {
	if strings.TrimSpace(b.Category) == "" {
		return errors.New("budget category is empty")
	}
	if b.Limit <= 0 {
		return errors.New("limit must be > 0")
	}
	return nil
}

func NormalizeCategory(s string) string {
	return strings.TrimSpace(strings.ToLower(s))
}
