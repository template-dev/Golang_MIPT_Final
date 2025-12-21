FROM golang:1.25-alpine AS builder

WORKDIR /src

RUN apk add --no-cache git ca-certificates && update-ca-certificates

COPY gateway/go.mod gateway/go.sum ./gateway/
RUN cd gateway && go mod download

COPY gateway ./gateway
RUN cd gateway && CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /out/gateway ./cmd/gateway

FROM alpine:latest AS runner

WORKDIR /app
RUN apk add --no-cache ca-certificates wget && update-ca-certificates

COPY --from=builder /out/gateway ./gateway

EXPOSE 8080
CMD ["./gateway"]
