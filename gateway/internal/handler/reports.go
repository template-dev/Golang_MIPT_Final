package handler

import (
	"net/http"

	"final/gateway/internal/httpx"
	ledgerv1 "final/gateway/ledger/v1"
)

func (h *Handler) ReportSummary(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	from := r.URL.Query().Get("from")
	to := r.URL.Query().Get("to")
	if from == "" || to == "" {
		httpx.WriteError(w, http.StatusBadRequest, "from and to are required")
		return
	}

	resp, err := h.client.GetReportSummary(r.Context(), &ledgerv1.ReportSummaryRequest{From: from, To: to})
	if err != nil {
		code, msg := grpcToHTTP(err)
		httpx.WriteError(w, code, msg)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, resp.GetTotals())
}
