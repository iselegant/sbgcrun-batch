.PHONY: all build test clean validate install-tools update-deps build-linux help

# ビルド後の出力先ディレクトリ
BUILD_DIR = bin
BINARY    = sbcntr-batch

# allターゲットでは「validate → build → run」を一括実行
all: validate build

# ビルド
build:
	@echo "==> Building batch binary"
	go build -ldflags "-s -w" -o $(BUILD_DIR)/$(BINARY) cmd/batch/main.go

# クリーンアップ
clean:
	@echo "==> Cleaning build outputs"
	rm -rf $(BUILD_DIR)/

##
# 検証系: fmt, vet, test
##
validate:
	@echo "==> Running go fmt"
	go fmt ./...

	@echo "==> Running go vet"
	go vet ./...

	@echo "==> Running golangci-lint"
	golangci-lint run

##
# 実行
##
run:
	@echo "==> Running batch"
	$(BUILD_DIR)/$(BINARY)

##
# テスト
##
test:
	@echo "==> Running tests"
	go test -v ./...

# 開発ツールのインストール
install-tools:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

##
# 依存関係の更新: go mod tidy
##
update-deps:
	@echo "==> Updating dependencies"
	go mod tidy

##
# Linux向けクロスコンパイル
##
build-linux:
	@echo "==> Cross compiling for Linux (amd64)"
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags "-s -w" -o $(BUILD_DIR)/$(BINARY) cmd/batch/main.go

##
# ヘルプ
##
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@echo "  all            validate + build を実行（デフォルト）"
	@echo "  build          バイナリをビルド ($(BUILD_DIR)/$(BINARY))"
	@echo "  build-linux    Linux arm64 向けクロスコンパイル"
	@echo "  run            ビルド済みバイナリを実行"
	@echo "  test           テストを実行"
	@echo "  validate       go fmt, go vet, golangci-lint を実行"
	@echo "  clean          ビルド成果物を削除"
	@echo "  update-deps    go mod tidy を実行"
	@echo "  install-tools  開発ツール（golangci-lint）をインストール"
	@echo "  help           このヘルプを表示"
	@echo ""
	@echo "Environment variables:"
	@echo "  DB_HOST          DBホスト (default: localhost)"
	@echo "  DB_PORT          DBポート (default: 5432)"
	@echo "  DB_USERNAME      DBユーザー名 (default: sbcntrapp)"
	@echo "  DB_PASSWORD      DBパスワード (default: password)"
	@echo "  DB_NAME          DB名 (default: sbcntrapp)"
	@echo "  ENABLE_TRACING   Cloud Trace有効化 (default: false)"
