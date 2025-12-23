package domain

import "time"

type ReportSummary struct {
	From   time.Time
	To     time.Time
	Totals map[string]float64
}

type BudgetProgressItem struct {
	Category string
	Limit    float64
	Spent    float64
	Percent  float64
}

type ReportWithBudgetProgress struct {
	From     time.Time
	To       time.Time
	Total    float64
	ByCat    map[string]float64
	Progress []BudgetProgressItem
}
