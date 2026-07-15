FROM golang:1.24 AS builder
WORKDIR /src
COPY go.mod ./
COPY . .
RUN CGO_ENABLED=0 go build -o /out/trading-service ./cmd/server

FROM alpine:latest
COPY --from=builder /out/trading-service /trading-service
CMD ["/trading-service"]
