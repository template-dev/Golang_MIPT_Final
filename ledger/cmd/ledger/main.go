package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"time"

	"final/ledger"
)

func main() {
	fmt.Println("Ledger service started")

	if err := ledger.Init(context.Background()); err != nil {
		fmt.Println("ledger init error:", err)
		os.Exit(1)
	}

	if _, err := ledger.SetBudget(ledger.Budget{Category: "еда", Limit: 5000, Period: "month"}); err != nil {
		fmt.Println("SetBudget error:", err)
	}
	if _, err := ledger.SetBudget(ledger.Budget{Category: "transport", Limit: 3000, Period: "month"}); err != nil {
		fmt.Println("SetBudget error:", err)
	}

	f, err := os.Open("budgets.json")
	if err == nil {
		defer f.Close()
		if err := ledger.LoadBudgets(bufio.NewReader(f)); err != nil {
			fmt.Println("LoadBudgets error:", err)
		} else {
			fmt.Println("Budgets loaded from budgets.json")
		}
	} else {
		fmt.Println("budgets.json not loaded:", err)
	}

	if _, err := ledger.AddTransaction(ledger.Transaction{
		Amount:      1200,
		Category:    "еда",
		Description: "groceries",
		Date:        time.Now(),
	}); err != nil {
		fmt.Println("AddTransaction error:", err)
	}

	if _, err := ledger.AddTransaction(ledger.Transaction{
		Amount:      3500,
		Category:    "еда",
		Description: "big purchase",
		Date:        time.Now(),
	}); err != nil {
		fmt.Println("AddTransaction error:", err)
	}

	txs, err := ledger.ListTransactions()
	if err != nil {
		fmt.Println("ListTransactions error:", err)
		return
	}
	fmt.Println("Transactions:", txs)
}
