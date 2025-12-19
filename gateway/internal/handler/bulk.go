package handler

import (
	"net/http"
	"strconv"

	"final/gateway/internal/api"
	"final/gateway/internal/httpx"
	ledgerv1 "final/gateway/ledger/v1"
)

func (h *Handler) BulkImportTransactions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	workers := int32(4)
	if v := r.URL.Query().Get("workers"); v != "" {
		if x, err := strconv.Atoi(v); err == nil && x > 0 {
			workers = int32(x)
		}
	}

	var req []api.CreateTransactionRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid json: "+err.Error())
		return
	}

	items := make([]*ledgerv1.CreateTransactionRequest, 0, len(req))
	for _, it := range req {
		items = append(items, &ledgerv1.CreateTransactionRequest{
			Amount:      it.Amount,
			Category:    it.Category,
			Description: it.Description,
			Date:        it.Date,
		})
	}

	resp, err := h.client.BulkImportTransactions(r.Context(), &ledgerv1.BulkImportTransactionsRequest{
		Items:   items,
		Workers: workers,
	})
	if err != nil {
		code, msg := grpcToHTTP(err)
		httpx.WriteError(w, code, msg)
		return
	}

	errorsOut := make([]map[string]any, 0, len(resp.GetErrors()))
	for _, e := range resp.GetErrors() {
		errorsOut = append(errorsOut, map[string]any{
			"index": int(e.GetIndex()),
			"error": e.GetError(),
		})
	}

	httpx.WriteJSON(w, http.StatusOK, map[string]any{
		"accepted": resp.GetAccepted(),
		"rejected": resp.GetRejected(),
		"errors":   errorsOut,
	})
}
