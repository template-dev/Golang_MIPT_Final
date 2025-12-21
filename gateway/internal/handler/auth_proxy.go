package handler

import (
	"bytes"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"final/gateway/internal/httpx"
)

func (h *Handler) AuthRegister(w http.ResponseWriter, r *http.Request) {
	h.proxyAuth(w, r, "/auth/register")
}

func (h *Handler) AuthLogin(w http.ResponseWriter, r *http.Request) {
	h.proxyAuth(w, r, "/auth/login")
}

func (h *Handler) proxyAuth(w http.ResponseWriter, r *http.Request, path string) {
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	base := strings.TrimRight(os.Getenv("AUTH_HTTP_ADDR"), "/")
	if base == "" {
		base = "http://localhost:8081"
	}

	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid body: "+err.Error())
		return
	}

	req, err := http.NewRequestWithContext(r.Context(), http.MethodPost, base+path, bytes.NewReader(body))
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "failed to build request")
		return
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		if r.Context().Err() != nil {
			httpx.WriteError(w, http.StatusGatewayTimeout, "request timeout")
			return
		}
		httpx.WriteError(w, http.StatusBadGateway, "auth service unavailable")
		return
	}
	defer resp.Body.Close()

	out, _ := io.ReadAll(io.LimitReader(resp.Body, 2<<20))

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(resp.StatusCode)

	if len(out) == 0 {
		return
	}
	_, _ = w.Write(out)
}
