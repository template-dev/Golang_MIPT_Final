package server

import (
	"net/http"

	"final/gateway/internal/handler"
	"final/gateway/internal/httpx"
	ledgerv1 "final/gateway/ledger/v1"
)

func NewRouter(client ledgerv1.LedgerServiceClient) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`"pong"`))
	})

	h := handler.New(client)

	mux.HandleFunc("/api/budgets", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			h.CreateBudget(w, r)
			return
		}
		if r.Method == http.MethodGet {
			h.ListBudgets(w, r)
			return
		}
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
	})

	mux.HandleFunc("/api/transactions", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			h.CreateTransaction(w, r)
			return
		}
		if r.Method == http.MethodGet {
			h.ListTransactions(w, r)
			return
		}
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
	})

	mux.HandleFunc("/api/reports/summary", func(w http.ResponseWriter, r *http.Request) {
		h.ReportSummary(w, r)
	})

	mux.HandleFunc("/api/transactions/bulk", func(w http.ResponseWriter, r *http.Request) {
		h.BulkImportTransactions(w, r)
	})

	return mux
}
