package database

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type DB struct {
	*sqlx.DB
}

type Config struct {
	Host     string
	Port     int
	UserName string
	Password string
	DBName   string
}

func NewDB(cfg Config) (*DB, error) {
	// localhost / 127.0.0.1（Cloud SQL Auth Proxy サイドカー経由）の場合は SSL を無効化
	var sslModeValue string
	dbHostEnv := os.Getenv("DB_HOST")
	if cfg.Host == "localhost" || cfg.Host == "127.0.0.1" ||
		dbHostEnv == "localhost" || dbHostEnv == "127.0.0.1" {
		sslModeValue = "disable"
	} else {
		sslModeValue = "require" // 本番環境ではSSLを有効にする
	}

	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host,
		cfg.Port,
		cfg.UserName,
		cfg.Password,
		cfg.DBName,
		sslModeValue,
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// コネクションプールの設定
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	// 接続テスト
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &DB{sqlx.NewDb(db, "postgres")}, nil
}
