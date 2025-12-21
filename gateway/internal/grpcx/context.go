package grpcx

import (
	"context"
	"errors"

	"final/gateway/internal/middleware"

	"google.golang.org/grpc/metadata"
)

var ErrNoUser = errors.New("missing user")

func OutgoingContext(ctx context.Context) (context.Context, error) {
	uid, ok := middleware.UserIDFromContext(ctx)
	if !ok {
		return nil, ErrNoUser
	}
	return metadata.AppendToOutgoingContext(ctx, "x-user-id", uid), nil
}
