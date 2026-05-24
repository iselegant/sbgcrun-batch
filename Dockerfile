# === builder: 依存関係生成用 ===
FROM golang:1.26 AS builder
ENV GO111MODULE=on \
    GOPATH=/go \
    GOBIN=/go/bin \
    PATH=/go/bin:$PATH
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
RUN go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.63.4
COPY . /app

# ビルドを実行
RUN make validate && \
    make build-linux

# === runner: 本番イメージ ===
FROM debian:12-slim AS runner

COPY --from=builder /app/bin/sbcntr-batch /main
ENTRYPOINT ["/main"]
