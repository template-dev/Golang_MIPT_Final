package ledger

import (
	"context"
	"errors"
	"time"

	ledgerv1 "final/ledger/ledger/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type GRPCServer struct {
	ledgerv1.UnimplementedLedgerServiceServer
	svc Service
}

func NewGRPCServer(svc Service) *GRPCServer {
	return &GRPCServer{svc: svc}
}

func (s *GRPCServer) AddTransaction(ctx context.Context, req *ledgerv1.CreateTransactionRequest) (*ledgerv1.Transaction, error) {
	tx, err := txFromReq(req)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	created, err := s.svc.AddTransaction(ctx, tx)
	if err != nil {
		return nil, mapServiceErr(err)
	}
	return txToPB(created), nil
}

func (s *GRPCServer) ListTransactions(ctx context.Context, _ *emptypb.Empty) (*ledgerv1.ListTransactionsResponse, error) {
	items, err := s.svc.ListTransactions(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, "internal error")
	}
	out := make([]*ledgerv1.Transaction, 0, len(items))
	for _, t := range items {
		tt := t
		out = append(out, txToPB(tt))
	}
	return &ledgerv1.ListTransactionsResponse{Items: out}, nil
}

func (s *GRPCServer) SetBudget(ctx context.Context, req *ledgerv1.CreateBudgetRequest) (*ledgerv1.Budget, error) {
	b := Budget{
		Category: req.GetCategory(),
		Limit:    req.GetLimit(),
		Period:   req.GetPeriod(),
	}
	created, err := s.svc.SetBudget(ctx, b)
	if err != nil {
		return nil, mapServiceErr(err)
	}
	return budgetToPB(created), nil
}

func (s *GRPCServer) ListBudgets(ctx context.Context, _ *emptypb.Empty) (*ledgerv1.ListBudgetsResponse, error) {
	items, err := s.svc.ListBudgets(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, "internal error")
	}
	out := make([]*ledgerv1.Budget, 0, len(items))
	for _, b := range items {
		bb := b
		out = append(out, budgetToPB(bb))
	}
	return &ledgerv1.ListBudgetsResponse{Items: out}, nil
}

func (s *GRPCServer) GetReportSummary(ctx context.Context, req *ledgerv1.ReportSummaryRequest) (*ledgerv1.ReportSummaryResponse, error) {
	from, err := time.Parse("2006-01-02", req.GetFrom())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid from")
	}
	to, err := time.Parse("2006-01-02", req.GetTo())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid to")
	}

	totals, err := s.svc.ReportSummary(ctx, from, to)
	if err != nil {
		return nil, mapServiceErr(err)
	}

	return &ledgerv1.ReportSummaryResponse{Totals: totals}, nil
}

func (s *GRPCServer) BulkImportTransactions(ctx context.Context, req *ledgerv1.BulkImportTransactionsRequest) (*ledgerv1.BulkImportTransactionsResponse, error) {
	workers := int(req.GetWorkers())
	items := req.GetItems()

	in := make([]ImportItem, 0, len(items))
	for i, it := range items {
		tx, err := txFromReq(it)
		if err != nil {
			in = append(in, ImportItem{Index: i, Tx: Transaction{}})
			continue
		}
		in = append(in, ImportItem{Index: i, Tx: tx})
	}

	summary, err := s.svc.BulkImportTransactions(ctx, in, workers)
	if err != nil {
		return nil, mapServiceErr(err)
	}

	errs := make([]*ledgerv1.ImportError, 0, len(summary.Errors))
	for _, e := range summary.Errors {
		ee := e
		errs = append(errs, &ledgerv1.ImportError{Index: int32(ee.Index), Error: ee.Error})
	}

	return &ledgerv1.BulkImportTransactionsResponse{
		Accepted: summary.Accepted,
		Rejected: summary.Rejected,
		Errors:   errs,
	}, nil
}

func txFromReq(req *ledgerv1.CreateTransactionRequest) (Transaction, error) {
	dt, err := time.Parse(time.RFC3339, req.GetDate())
	if err != nil {
		return Transaction{}, errors.New("invalid date")
	}
	return Transaction{
		Amount:      req.GetAmount(),
		Category:    req.GetCategory(),
		Description: req.GetDescription(),
		Date:        dt,
	}, nil
}

func txToPB(t Transaction) *ledgerv1.Transaction {
	return &ledgerv1.Transaction{
		Id:          int64(t.ID),
		Amount:      t.Amount,
		Category:    t.Category,
		Description: t.Description,
		Date:        t.Date.Format(time.RFC3339),
	}
}

func budgetToPB(b Budget) *ledgerv1.Budget {
	return &ledgerv1.Budget{
		Category: b.Category,
		Limit:    b.Limit,
		Period:   b.Period,
	}
}

func mapServiceErr(err error) error {
	if errors.Is(err, ErrBudgetExceeded) || err.Error() == "budget exceeded" {
		return status.Error(codes.FailedPrecondition, "budget exceeded")
	}
	if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
		return status.Error(codes.DeadlineExceeded, "timeout")
	}
	if err != nil {
		if isValidationErr(err) {
			return status.Error(codes.InvalidArgument, err.Error())
		}
		return status.Error(codes.Internal, "internal error")
	}
	return nil
}

func isValidationErr(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	switch msg {
	case "amount must be > 0",
		"transaction category is empty",
		"date is required",
		"budget category is empty",
		"limit must be > 0",
		"from must be <= to",
		"invalid date",
		"invalid from",
		"invalid to":
		return true
	default:
		return false
	}
}
