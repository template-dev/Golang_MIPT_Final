package handler

import (
	ledgerv1 "final/gen/ledger/v1"
)

type Handler struct {
	client ledgerv1.LedgerServiceClient
}

func New(client ledgerv1.LedgerServiceClient) *Handler {
	return &Handler{client: client}
}
