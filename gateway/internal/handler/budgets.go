package handler

import (
	"net/http"

	"final/gateway/internal/api"
	"final/gateway/internal/httpx"
	ledgerv1 "final/gateway/ledger/v1"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (h *Handler) CreateBudget(w http.ResponseWriter, r *http.Request) {
	var req api.CreateBudgetRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid json: "+err.Error())
		return
	}

	p := req.Period
	if p == "" {
		p = "fixed"
	}

	resp, err := h.client.SetBudget(r.Context(), &ledgerv1.CreateBudgetRequest{
		Category: req.Category,
		Limit:    req.Limit,
		Period:   p,
	})
	if err != nil {
		code, msg := grpcToHTTP(err)
		httpx.WriteError(w, code, msg)
		return
	}

	httpx.WriteJSON(w, http.StatusCreated, api.BudgetResponse{
		Category: resp.GetCategory(),
		Limit:    resp.GetLimit(),
		Period:   resp.GetPeriod(),
	})
}

func (h *Handler) ListBudgets(w http.ResponseWriter, r *http.Request) {
	resp, err := h.client.ListBudgets(r.Context(), &emptypb.Empty{})
	if err != nil {
		code, msg := grpcToHTTP(err)
		httpx.WriteError(w, code, msg)
		return
	}

	out := make([]api.BudgetResponse, 0, len(resp.GetItems()))
	for _, b := range resp.GetItems() {
		out = append(out, api.BudgetResponse{
			Category: b.GetCategory(),
			Limit:    b.GetLimit(),
			Period:   b.GetPeriod(),
		})
	}

	httpx.WriteJSON(w, http.StatusOK, out)
}
