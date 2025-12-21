package handler

import (
	"net/http"

	"final/gateway/internal/api"
	"final/gateway/internal/grpcx"
	"final/gateway/internal/httpx"
	ledgerv1 "final/gen/ledger/v1"

	"google.golang.org/protobuf/types/known/emptypb"
)

func (h *Handler) CreateBudget(w http.ResponseWriter, r *http.Request) {
	var req api.CreateBudgetRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid json: "+err.Error())
		return
	}

	ctx, err := grpcx.OutgoingContext(r.Context())
	if err != nil {
		httpx.WriteError(w, http.StatusUnauthorized, "missing user")
		return
	}

	resp, err := h.client.SetBudget(ctx, &ledgerv1.CreateBudgetRequest{
		Category: req.Category,
		Limit:    req.Limit,
		Period:   "fixed",
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
	ctx, err := grpcx.OutgoingContext(r.Context())
	if err != nil {
		httpx.WriteError(w, http.StatusUnauthorized, "missing user")
		return
	}

	resp, err := h.client.ListBudgets(ctx, &emptypb.Empty{})
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
