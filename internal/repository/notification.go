package repository

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/uma-arai/sbcntr-batch/internal/model"
)

// NotificationRepository は通知の永続化を担当するインターフェースです
type NotificationRepository interface {
	CreateNotifications(ctx context.Context, records []model.NotificationRecord) error
	Create(ctx context.Context, tx *sqlx.Tx, record *model.NotificationRecord) error
	GetByUserID(ctx context.Context, userID string) ([]model.NotificationRecord, error)
	UpdateIsRead(ctx context.Context, tx *sqlx.Tx, id int, isRead bool) error
}

// NotificationRepositoryImpl は通知の永続化を担当します
type NotificationRepositoryImpl struct {
	db *DB
}

// NewNotificationRepository は新しいNotificationRepositoryを作成します
func NewNotificationRepository(db *DB) *NotificationRepositoryImpl {
	return &NotificationRepositoryImpl{
		db: db,
	}
}

// CreateNotifications は複数の通知レコードを作成します
func (r *NotificationRepositoryImpl) CreateNotifications(ctx context.Context, records []model.NotificationRecord) error {
	tx, err := r.db.BeginTx()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	// トランザクションのロールバックを遅延実行
	// エラーが発生した場合のみロールバックを実行
	var rollbackErr error
	defer func() {
		if err != nil {
			if rbErr := tx.Rollback(); rbErr != nil {
				rollbackErr = fmt.Errorf("rollback failed: %v, original error: %w", rbErr, err)
			}
		}
	}()

	for _, record := range records {
		if err := r.Create(ctx, tx, &record); err != nil {
			return fmt.Errorf("failed to create notification: %w", err)
		}
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	if rollbackErr != nil {
		return rollbackErr
	}

	return nil
}

// Create は単一の通知レコードを作成します
func (r *NotificationRepositoryImpl) Create(ctx context.Context, tx *sqlx.Tx, record *model.NotificationRecord) error {
	query := `
		INSERT INTO notifications (
			user_id, title, message, is_read, type, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7
		)
		RETURNING id`

	err := tx.QueryRowContext(ctx,
		query,
		record.UserID,
		record.Title,
		record.Message,
		record.IsRead,
		record.Type,
		record.CreatedAt,
		record.UpdatedAt,
	).Scan(&record.ID)

	if err != nil {
		return err
	}

	return nil
}

// BeginTx は新しいトランザクションを開始します
func (r *NotificationRepositoryImpl) BeginTx() (*sqlx.Tx, error) {
	return r.db.BeginTxx(context.Background(), nil)
}

// GetByUserID は指定されたユーザーIDの通知を取得します
func (r *NotificationRepositoryImpl) GetByUserID(ctx context.Context, userID string) ([]model.NotificationRecord, error) {
	query := `
		SELECT id, user_id, title, message, is_read, type, created_at, updated_at
		FROM notifications
		WHERE user_id = $1
		ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query notifications: %w", err)
	}
	defer rows.Close()

	var records []model.NotificationRecord
	for rows.Next() {
		var record model.NotificationRecord
		err := rows.Scan(
			&record.ID,
			&record.UserID,
			&record.Title,
			&record.Message,
			&record.IsRead,
			&record.Type,
			&record.CreatedAt,
			&record.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan notification: %w", err)
		}
		records = append(records, record)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating notifications: %w", err)
	}

	return records, nil
}

// UpdateIsRead は通知の既読状態を更新します
func (r *NotificationRepositoryImpl) UpdateIsRead(ctx context.Context, tx *sqlx.Tx, id int, isRead bool) error {
	query := `
		UPDATE notifications
		SET is_read = $1, updated_at = CURRENT_TIMESTAMP
		WHERE id = $2`

	result, err := tx.ExecContext(ctx, query, isRead, id)
	if err != nil {
		return fmt.Errorf("failed to update notification is_read: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("notification with id %d not found", id)
	}

	return nil
}
