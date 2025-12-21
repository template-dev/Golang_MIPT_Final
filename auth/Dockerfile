FROM golang:1.25-alpine AS builder

WORKDIR /src

RUN apk add --no-cache git ca-certificates && update-ca-certificates

COPY auth/go.mod auth/go.sum ./auth/
RUN cd auth && go mod download

COPY auth ./auth
RUN cd auth && CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /out/auth ./cmd/auth

FROM alpine:latest AS runner

WORKDIR /app
RUN apk add --no-cache ca-certificates wget && update-ca-certificates

COPY --from=builder /out/auth ./auth

EXPOSE 8081
CMD ["./auth"]
