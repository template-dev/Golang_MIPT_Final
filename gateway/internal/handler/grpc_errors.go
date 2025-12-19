package handler

import (
	"net/http"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func grpcToHTTP(err error) (int, string) {
	if err == nil {
		return http.StatusOK, ""
	}
	st, ok := status.FromError(err)
	if !ok {
		return http.StatusInternalServerError, "internal error"
	}

	switch st.Code() {
	case codes.InvalidArgument:
		return http.StatusBadRequest, st.Message()
	case codes.FailedPrecondition, codes.Aborted:
		return http.StatusConflict, st.Message()
	case codes.DeadlineExceeded:
		return http.StatusGatewayTimeout, "timeout"
	default:
		return http.StatusInternalServerError, "internal error"
	}
}
