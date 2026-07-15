FROM golang:1.25 AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /out/trading-service ./cmd/server

FROM alpine:latest
COPY --from=builder /out/trading-service /trading-service
CMD ["/trading-service"]
