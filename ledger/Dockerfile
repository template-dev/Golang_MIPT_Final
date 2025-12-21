FROM golang:1.25-alpine AS builder

WORKDIR /src

RUN apk add --no-cache git ca-certificates && update-ca-certificates

COPY ledger/go.mod ledger/go.sum ./ledger/
RUN cd ledger && go mod download

COPY ledger ./ledger
RUN cd ledger && CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /out/ledger ./cmd/ledger-grpc

FROM alpine:latest AS runner

WORKDIR /app
RUN apk add --no-cache ca-certificates && update-ca-certificates

COPY --from=builder /out/ledger ./ledger

EXPOSE 50051
CMD ["./ledger"]
