FROM golang:1.25-alpine AS builder

WORKDIR /src

RUN apk add --no-cache git ca-certificates && update-ca-certificates

COPY go.work ./
COPY ledger/go.mod ledger/go.sum ./ledger/
COPY gateway/go.mod gateway/go.sum ./gateway/

RUN cd ledger && go mod download

COPY ledger ./ledger
COPY gateway ./gateway

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /out/ledger ./ledger/cmd/ledger

FROM alpine:latest AS runner

WORKDIR /app

RUN apk add --no-cache ca-certificates && update-ca-certificates

COPY --from=builder /out/ledger ./ledger

CMD ["./ledger"]
