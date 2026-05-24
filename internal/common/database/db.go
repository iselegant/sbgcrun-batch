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
	// localhostのDBの場合はSSLを無効化
	var sslModeValue string
	if cfg.Host == "localhost" || os.Getenv("DB_HOST") == "localhost" {
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
