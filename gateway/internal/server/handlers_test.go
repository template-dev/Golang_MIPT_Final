package server_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"final/gateway/internal/server"
	ledgerv1 "final/gen/ledger/v1"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type fakeLedgerClient struct {
	budgets      map[string]float64
	transactions []*ledgerv1.Transaction
}

func newFakeClient() *fakeLedgerClient {
	return &fakeLedgerClient{
		budgets:      map[string]float64{},
		transactions: []*ledgerv1.Transaction{},
	}
}

func normalizeCat(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}

func errInvalid(msg string) error {
	return status.Error(codes.InvalidArgument, msg)
}

func errBudgetExceeded() error {
	return status.Error(codes.FailedPrecondition, "budget exceeded")
}

func grpcMsg(err error) string {
	st, ok := status.FromError(err)
	if !ok {
		return err.Error()
	}
	return st.Message()
}

// --- LedgerServiceClient methods ---

func (f *fakeLedgerClient) AddTransaction(ctx context.Context, in *ledgerv1.CreateTransactionRequest, opts ...grpc.CallOption) (*ledgerv1.Transaction, error) {
	if in.GetAmount() <= 0 {
		return nil, errInvalid("amount must be > 0")
	}
	if strings.TrimSpace(in.GetCategory()) == "" {
		return nil, errInvalid("category is required")
	}
	if strings.TrimSpace(in.GetDate()) == "" {
		return nil, errInvalid("date is required")
	}

	cat := normalizeCat(in.GetCategory())
	limit, ok := f.budgets[cat]
	if ok {
		var spent float64
		for _, t := range f.transactions {
			if normalizeCat(t.GetCategory()) == cat {
				spent += t.GetAmount()
			}
		}
		if spent+in.GetAmount() > limit {
			return nil, errBudgetExceeded()
		}
	}

	id := int32(len(f.transactions) + 1)
	tx := &ledgerv1.Transaction{
		Id:          int64(id),
		Amount:      in.GetAmount(),
		Category:    in.GetCategory(),
		Description: in.GetDescription(),
		Date:        in.GetDate(),
	}
	f.transactions = append(f.transactions, tx)
	return tx, nil
}

func (f *fakeLedgerClient) ListTransactions(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*ledgerv1.ListTransactionsResponse, error) {
	return &ledgerv1.ListTransactionsResponse{Items: f.transactions}, nil
}

func (f *fakeLedgerClient) SetBudget(ctx context.Context, in *ledgerv1.CreateBudgetRequest, opts ...grpc.CallOption) (*ledgerv1.Budget, error) {
	if strings.TrimSpace(in.GetCategory()) == "" {
		return nil, errInvalid("budget category is empty")
	}
	if in.GetLimit() <= 0 {
		return nil, errInvalid("budget limit must be > 0")
	}

	cat := normalizeCat(in.GetCategory())
	f.budgets[cat] = in.GetLimit()

	// Period в твоём DTO есть, но в CreateBudgetRequest может не быть.
	// Возвращаем фикс.
	return &ledgerv1.Budget{
		Category: in.GetCategory(),
		Limit:    in.GetLimit(),
		Period:   "fixed",
	}, nil
}

func (f *fakeLedgerClient) ListBudgets(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*ledgerv1.ListBudgetsResponse, error) {
	items := make([]*ledgerv1.Budget, 0, len(f.budgets))
	for k, v := range f.budgets {
		items = append(items, &ledgerv1.Budget{Category: k, Limit: v, Period: "fixed"})
	}
	return &ledgerv1.ListBudgetsResponse{Items: items}, nil
}

func (f *fakeLedgerClient) GetReportSummary(ctx context.Context, in *ledgerv1.ReportSummaryRequest, opts ...grpc.CallOption) (*ledgerv1.ReportSummaryResponse, error) {
	totals := map[string]float64{}
	for _, t := range f.transactions {
		totals[normalizeCat(t.GetCategory())] += t.GetAmount()
	}
	return &ledgerv1.ReportSummaryResponse{Totals: totals}, nil
}

func (f *fakeLedgerClient) BulkImportTransactions(ctx context.Context, in *ledgerv1.BulkImportTransactionsRequest, opts ...grpc.CallOption) (*ledgerv1.BulkImportTransactionsResponse, error) {
	var accepted int64
	var rejected int64
	var errs []*ledgerv1.BulkImportError

	for i, it := range in.GetItems() {
		_, err := f.AddTransaction(ctx, it)
		if err != nil {
			rejected++
			errs = append(errs, &ledgerv1.BulkImportError{
				Index: int32(i),
				Error: grpcMsg(err),
			})
			continue
		}
		accepted++
	}

	return &ledgerv1.BulkImportTransactionsResponse{
		Accepted: accepted,
		Rejected: rejected,
		Errors:   errs,
	}, nil
}

// --- helpers ---

func doReq(t *testing.T, h http.Handler, method, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	return rr
}

func TestBudgetsHandlers(t *testing.T) {
	fc := newFakeClient()
	h := server.NewRouter(fc)

	t.Run("ok_create_budget", func(t *testing.T) {
		rr := doReq(t, h, http.MethodPost, "/api/budgets", `{"category":"еда","limit":5000}`)
		if rr.Code != http.StatusCreated {
			t.Fatalf("expected %d, got %d, body=%s", http.StatusCreated, rr.Code, rr.Body.String())
		}
		if rr.Header().Get("Content-Type") != "application/json; charset=utf-8" {
			t.Fatalf("unexpected content-type: %s", rr.Header().Get("Content-Type"))
		}

		var got map[string]any
		if err := json.NewDecoder(bytes.NewReader(rr.Body.Bytes())).Decode(&got); err != nil {
			t.Fatalf("decode error: %v", err)
		}
		if got["category"] == nil || got["limit"] == nil {
			t.Fatalf("unexpected json: %v", got)
		}
	})

	t.Run("get_budgets_contains_added", func(t *testing.T) {
		_ = doReq(t, h, http.MethodPost, "/api/budgets", `{"category":"еда","limit":5000}`)
		rr := doReq(t, h, http.MethodGet, "/api/budgets", "")
		if rr.Code != http.StatusOK {
			t.Fatalf("expected %d, got %d, body=%s", http.StatusOK, rr.Code, rr.Body.String())
		}
		var arr []map[string]any
		_ = json.NewDecoder(bytes.NewReader(rr.Body.Bytes())).Decode(&arr)
		if len(arr) == 0 {
			t.Fatalf("expected budgets, got %v", arr)
		}
	})

	t.Run("bad_budget_limit", func(t *testing.T) {
		rr := doReq(t, h, http.MethodPost, "/api/budgets", `{"category":"еда","limit":0}`)
		if rr.Code != http.StatusBadRequest {
			t.Fatalf("expected %d, got %d, body=%s", http.StatusBadRequest, rr.Code, rr.Body.String())
		}
		var got map[string]string
		_ = json.NewDecoder(bytes.NewReader(rr.Body.Bytes())).Decode(&got)
		if got["error"] == "" {
			t.Fatalf("expected error json, got %s", rr.Body.String())
		}
	})

	t.Run("bad_json", func(t *testing.T) {
		rr := doReq(t, h, http.MethodPost, "/api/budgets", `{"category":"еда","limit":}`)
		if rr.Code != http.StatusBadRequest {
			t.Fatalf("expected %d, got %d, body=%s", http.StatusBadRequest, rr.Code, rr.Body.String())
		}
	})
}

func TestTransactionsFlow(t *testing.T) {
	fc := newFakeClient()
	h := server.NewRouter(fc)

	t.Run("ok", func(t *testing.T) {
		_ = doReq(t, h, http.MethodPost, "/api/budgets", `{"category":"food","limit":5000}`)

		rr := doReq(t, h, http.MethodPost, "/api/transactions",
			`{"amount":1200,"category":"food","description":"groceries","date":"2025-12-19T21:29:42+03:00"}`)
		if rr.Code != http.StatusCreated {
			t.Fatalf("expected %d, got %d, body=%s", http.StatusCreated, rr.Code, rr.Body.String())
		}

		rr2 := doReq(t, h, http.MethodGet, "/api/transactions", "")
		if rr2.Code != http.StatusOK {
			t.Fatalf("expected %d, got %d, body=%s", http.StatusOK, rr2.Code, rr2.Body.String())
		}
		var arr []map[string]any
		_ = json.NewDecoder(bytes.NewReader(rr2.Body.Bytes())).Decode(&arr)
		if len(arr) != 1 {
			t.Fatalf("expected 1 transaction, got %d, body=%s", len(arr), rr2.Body.String())
		}
	})

	t.Run("exceeded", func(t *testing.T) {
		fc2 := newFakeClient()
		h2 := server.NewRouter(fc2)

		_ = doReq(t, h2, http.MethodPost, "/api/budgets", `{"category":"food","limit":1000}`)

		rr1 := doReq(t, h2, http.MethodPost, "/api/transactions",
			`{"amount":900,"category":"food","description":"a","date":"2025-12-19T00:00:00+03:00"}`)
		if rr1.Code != http.StatusCreated {
			t.Fatalf("expected %d, got %d, body=%s", http.StatusCreated, rr1.Code, rr1.Body.String())
		}

		rr2 := doReq(t, h2, http.MethodPost, "/api/transactions",
			`{"amount":200,"category":"food","description":"b","date":"2025-12-19T00:00:00+03:00"}`)
		if rr2.Code != http.StatusConflict {
			t.Fatalf("expected %d, got %d, body=%s", http.StatusConflict, rr2.Code, rr2.Body.String())
		}

		var got map[string]string
		_ = json.NewDecoder(bytes.NewReader(rr2.Body.Bytes())).Decode(&got)
		if got["error"] != "budget exceeded" {
			t.Fatalf("expected budget exceeded, got %v", got)
		}
	})

	t.Run("bad_json", func(t *testing.T) {
		rr := doReq(t, h, http.MethodPost, "/api/transactions",
			`{"amount":1200,"category":"food","description":"x","date":}`)
		if rr.Code != http.StatusBadRequest {
			t.Fatalf("expected %d, got %d, body=%s", http.StatusBadRequest, rr.Code, rr.Body.String())
		}
	})
}
