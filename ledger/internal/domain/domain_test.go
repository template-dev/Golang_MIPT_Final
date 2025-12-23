package domain

import (
	"testing"
	"time"
)

func TestTransactionValidate(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		tx      Transaction
		wantErr bool
	}{
		{
			name: "ok",
			tx: Transaction{
				Amount:      10,
				Category:    "еда",
				Description: "x",
				Date:        time.Now(),
			},
			wantErr: false,
		},
		{
			name: "zero_amount",
			tx: Transaction{
				Amount:      0,
				Category:    "еда",
				Description: "x",
				Date:        time.Now(),
			},
			wantErr: true,
		},
		{
			name: "negative_amount",
			tx: Transaction{
				Amount:      -1,
				Category:    "еда",
				Description: "x",
				Date:        time.Now(),
			},
			wantErr: true,
		},
		{
			name: "empty_category",
			tx: Transaction{
				Amount:      10,
				Category:    "   ",
				Description: "x",
				Date:        time.Now(),
			},
			wantErr: true,
		},
		{
			name: "zero_date",
			tx: Transaction{
				Amount:      10,
				Category:    "еда",
				Description: "x",
				Date:        time.Time{},
			},
			wantErr: true,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			err := tc.tx.Validate()
			if tc.wantErr && err == nil {
				t.Fatalf("expected error, got nil")
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("expected nil, got %v", err)
			}
		})
	}
}

func TestBudgetValidate(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		b       Budget
		wantErr bool
	}{
		{name: "ok", b: Budget{Category: "еда", Limit: 1, Period: "fixed"}, wantErr: false},
		{name: "zero_limit", b: Budget{Category: "еда", Limit: 0, Period: "fixed"}, wantErr: true},
		{name: "negative_limit", b: Budget{Category: "еда", Limit: -1, Period: "fixed"}, wantErr: true},
		{name: "empty_category", b: Budget{Category: "   ", Limit: 10, Period: "fixed"}, wantErr: true},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			err := tc.b.Validate()
			if tc.wantErr && err == nil {
				t.Fatalf("expected error, got nil")
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("expected nil, got %v", err)
			}
		})
	}
}
