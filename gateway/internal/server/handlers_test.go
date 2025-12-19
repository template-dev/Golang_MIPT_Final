package server_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"final/gateway/internal/server"
	"final/ledger"
)

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
	ledger.Reset()
	t.Cleanup(ledger.Reset)

	h := server.NewRouter()

	t.Run("ok_create_budget", func(t *testing.T) {
		rr := doReq(t, h, http.MethodPost, "/api/budgets", `{"category":"еда","limit":5000}`)

		if rr.Code != http.StatusCreated {
			t.Fatalf("expected %d, got %d, body=%s", http.StatusCreated, rr.Code, rr.Body.String())
		}

		ct := rr.Header().Get("Content-Type")
		if ct != "application/json; charset=utf-8" {
			t.Fatalf("unexpected content-type: %s", ct)
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

		ct := rr.Header().Get("Content-Type")
		if ct != "application/json; charset=utf-8" {
			t.Fatalf("unexpected content-type: %s", ct)
		}

		var arr []map[string]any
		if err := json.NewDecoder(bytes.NewReader(rr.Body.Bytes())).Decode(&arr); err != nil {
			t.Fatalf("decode error: %v", err)
		}

		found := false
		for _, it := range arr {
			if it["category"] == "еда" {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("expected budget in list, got %v", arr)
		}
	})

	t.Run("bad_budget_limit", func(t *testing.T) {
		rr := doReq(t, h, http.MethodPost, "/api/budgets", `{"category":"еда","limit":0}`)

		if rr.Code != http.StatusBadRequest {
			t.Fatalf("expected %d, got %d, body=%s", http.StatusBadRequest, rr.Code, rr.Body.String())
		}

		ct := rr.Header().Get("Content-Type")
		if ct != "application/json; charset=utf-8" {
			t.Fatalf("unexpected content-type: %s", ct)
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

		var got map[string]string
		_ = json.NewDecoder(bytes.NewReader(rr.Body.Bytes())).Decode(&got)
		if got["error"] == "" {
			t.Fatalf("expected error json, got %s", rr.Body.String())
		}
	})
}

func TestTransactionsFlow(t *testing.T) {
	ledger.Reset()
	t.Cleanup(ledger.Reset)

	h := server.NewRouter()

	t.Run("ok", func(t *testing.T) {
		_ = doReq(t, h, http.MethodPost, "/api/budgets", `{"category":"food","limit":5000}`)

		rr := doReq(t, h, http.MethodPost, "/api/transactions",
			`{"amount":1200,"category":"food","description":"groceries","date":"2025-12-19T21:29:42+03:00"}`)

		if rr.Code != http.StatusCreated {
			t.Fatalf("expected %d, got %d, body=%s", http.StatusCreated, rr.Code, rr.Body.String())
		}

		ct := rr.Header().Get("Content-Type")
		if ct != "application/json; charset=utf-8" {
			t.Fatalf("unexpected content-type: %s", ct)
		}

		var got map[string]any
		if err := json.NewDecoder(bytes.NewReader(rr.Body.Bytes())).Decode(&got); err != nil {
			t.Fatalf("decode error: %v", err)
		}

		if got["id"] == nil || got["amount"] == nil || got["category"] != "food" {
			t.Fatalf("unexpected json: %v", got)
		}

		rr2 := doReq(t, h, http.MethodGet, "/api/transactions", "")
		if rr2.Code != http.StatusOK {
			t.Fatalf("expected %d, got %d, body=%s", http.StatusOK, rr2.Code, rr2.Body.String())
		}

		var arr []map[string]any
		if err := json.NewDecoder(bytes.NewReader(rr2.Body.Bytes())).Decode(&arr); err != nil {
			t.Fatalf("decode error: %v", err)
		}
		if len(arr) != 1 {
			t.Fatalf("expected 1 transaction, got %d, body=%s", len(arr), rr2.Body.String())
		}
	})

	t.Run("exceeded", func(t *testing.T) {
		ledger.Reset()

		_ = doReq(t, h, http.MethodPost, "/api/budgets", `{"category":"food","limit":1000}`)

		rr1 := doReq(t, h, http.MethodPost, "/api/transactions",
			`{"amount":900,"category":"food","description":"a","date":"2025-12-19T00:00:00+03:00"}`)
		if rr1.Code != http.StatusCreated {
			t.Fatalf("expected %d, got %d, body=%s", http.StatusCreated, rr1.Code, rr1.Body.String())
		}

		rr2 := doReq(t, h, http.MethodPost, "/api/transactions",
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
		ledger.Reset()

		_ = doReq(t, h, http.MethodPost, "/api/budgets", `{"category":"food","limit":1000}`)

		rr := doReq(t, h, http.MethodPost, "/api/transactions",
			`{"amount":1200,"category":"food","description":"x","date":}`)
		if rr.Code != http.StatusBadRequest {
			t.Fatalf("expected %d, got %d, body=%s", http.StatusBadRequest, rr.Code, rr.Body.String())
		}

		var got map[string]string
		_ = json.NewDecoder(bytes.NewReader(rr.Body.Bytes())).Decode(&got)
		if got["error"] == "" {
			t.Fatalf("expected error json, got %s", rr.Body.String())
		}
	})
}
