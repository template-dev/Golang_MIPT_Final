package handler

import (
	"net/http"

	"final/gateway/internal/api"
	"final/gateway/internal/httpx"
	ledgerv1 "final/gen/ledger/v1"

	"google.golang.org/protobuf/types/known/emptypb"
)

func (h *Handler) CreateTransaction(w http.ResponseWriter, r *http.Request) {
	var req api.CreateTransactionRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid json: "+err.Error())
		return
	}

	txReq := &ledgerv1.CreateTransactionRequest{
		Amount:      req.Amount,
		Category:    req.Category,
		Description: req.Description,
		Date:        req.Date,
	}

	created, err := h.client.AddTransaction(r.Context(), txReq)
	if err != nil {
		code, msg := grpcToHTTP(err)
		httpx.WriteError(w, code, msg)
		return
	}

	httpx.WriteJSON(w, http.StatusCreated, api.TransactionResponse{
		ID:          int(created.GetId()),
		Amount:      created.GetAmount(),
		Category:    created.GetCategory(),
		Description: created.GetDescription(),
		Date:        created.GetDate(),
	})
}

func (h *Handler) ListTransactions(w http.ResponseWriter, r *http.Request) {
	resp, err := h.client.ListTransactions(r.Context(), &emptypb.Empty{})
	if err != nil {
		code, msg := grpcToHTTP(err)
		httpx.WriteError(w, code, msg)
		return
	}

	out := make([]api.TransactionResponse, 0, len(resp.GetItems()))
	for _, t := range resp.GetItems() {
		out = append(out, api.TransactionResponse{
			ID:          int(t.GetId()),
			Amount:      t.GetAmount(),
			Category:    t.GetCategory(),
			Description: t.GetDescription(),
			Date:        t.GetDate(),
		})
	}
	httpx.WriteJSON(w, http.StatusOK, out)
}
