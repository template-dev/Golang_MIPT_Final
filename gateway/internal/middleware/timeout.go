package middleware

import (
	"context"
	"net/http"
	"os"
	"strconv"
	"time"

	"final/gateway/internal/httpx"
)

func Timeout(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		timeout := timeoutFromEnv()
		ctx, cancel := context.WithTimeout(r.Context(), timeout)
		defer cancel()

		rr := &respRecorder{ResponseWriter: w}
		next.ServeHTTP(rr, r.WithContext(ctx))

		if ctx.Err() == context.DeadlineExceeded && !rr.wrote {
			httpx.WriteError(w, http.StatusGatewayTimeout, "timeout")
		}
	})
}

type respRecorder struct {
	http.ResponseWriter
	wrote bool
}

func (r *respRecorder) WriteHeader(statusCode int) {
	r.wrote = true
	r.ResponseWriter.WriteHeader(statusCode)
}

func (r *respRecorder) Write(b []byte) (int, error) {
	r.wrote = true
	return r.ResponseWriter.Write(b)
}

func timeoutFromEnv() time.Duration {
	if v := os.Getenv("REQUEST_TIMEOUT_MS"); v != "" {
		if ms, err := strconv.Atoi(v); err == nil && ms > 0 {
			return time.Duration(ms) * time.Millisecond
		}
	}
	return 2 * time.Second
}
