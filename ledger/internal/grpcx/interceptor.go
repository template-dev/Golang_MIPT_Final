package grpcx

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func UserIDUnaryInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if ok {
			vals := md.Get("x-user-id")
			if len(vals) > 0 && vals[0] != "" {
				ctx = WithUserID(ctx, vals[0])
			}
		}
		return handler(ctx, req)
	}
}
