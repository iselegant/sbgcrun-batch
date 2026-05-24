package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"
	"time"

	"github.com/uma-arai/sbcntr-batch/internal/common/config"
	"github.com/uma-arai/sbcntr-batch/internal/common/tracing"
	"github.com/uma-arai/sbcntr-batch/internal/common/utils"
	"github.com/uma-arai/sbcntr-batch/internal/service/batch"
)

const (
	serviceName    = "sbcntr-batch"
	defaultTimeout = 5 * time.Minute
)

func main() {
	// 設定の読み込み
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v\nStack trace:\n%s", err, debug.Stack())
	}

	// トレーサー初期化（ENABLE_TRACING=true でCloud Trace、それ以外はNoop）
	shutdown, err := tracing.InitTracer(serviceName, cfg.EnableTracing)
	if err != nil {
		log.Printf("Failed to initialize tracer: %v", err)
	} else {
		defer func() {
			if err := shutdown(context.Background()); err != nil {
				log.Printf("Failed to shutdown tracer: %v", err)
			}
		}()
	}

	// サービスの初期化
	service, err := batch.NewReservationBatchService(cfg)
	if err != nil {
		log.Fatalf("Failed to create service: %v\nStack trace:\n%s", err, debug.Stack())
	}
	defer service.Close()

	// コンテキストの作成
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	// シグナルハンドリング
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// バッチ処理の実行
	errChan := make(chan error, 1)
	go func() {
		errChan <- utils.RunWithTimeout(ctx, defaultTimeout, service.Run)
	}()

	log.Println("Starting batch process...")

	select {
	case sig := <-sigChan:
		log.Printf("Received signal: %v", sig)
		cancel()
	case err := <-errChan:
		if err != nil {
			log.Printf("Batch process failed: %v\nStack trace:\n%s", err, debug.Stack())
			os.Exit(1)
		}
		log.Println("Batch process completed successfully")
	}
}
