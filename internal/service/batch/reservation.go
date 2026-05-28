package batch

import (
	"context"
	"fmt"
	"log"
	"slices"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"

	"github.com/uma-arai/sbgcrun-batch/internal/common/config"
	"github.com/uma-arai/sbgcrun-batch/internal/common/database"
	"github.com/uma-arai/sbgcrun-batch/internal/common/utils"
	"github.com/uma-arai/sbgcrun-batch/internal/model"
	"github.com/uma-arai/sbgcrun-batch/internal/repository"
)

var reservationTracer = otel.Tracer("sbgcrun-batch/reservation")

// ReservationBatchService は予約バッチ処理を担当します
type ReservationBatchService struct {
	db               *database.DB
	reservationRepo  repository.ReservationRepository
	notificationRepo repository.NotificationRepository
	petRepo          repository.PetRepository
	cfg              *config.Config
}

// NewReservationBatchService は新しいReservationBatchServiceを作成します
func NewReservationBatchService(cfg *config.Config) (*ReservationBatchService, error) {
	db, err := database.NewDB(cfg.DB)
	if err != nil {
		return nil, fmt.Errorf("failed to create database connection: %w", err)
	}

	repoDb := &repository.DB{DB: db.DB}

	return &ReservationBatchService{
		db:               db,
		reservationRepo:  repository.NewReservationRepository(repoDb),
		notificationRepo: repository.NewNotificationRepository(repoDb),
		petRepo:          repository.NewPetRepository(repoDb),
		cfg:              cfg,
	}, nil
}

// Close は終了処理を行います
func (s *ReservationBatchService) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}

// Run は予約バッチ処理を実行します
func (s *ReservationBatchService) Run(ctx context.Context) error {
	ctx, span := reservationTracer.Start(ctx, "ReservationBatchService.Run")
	defer span.End()

	startTime := time.Now()

	// pending状態の予約を処理
	events, err := s.processReservationsByStatus(ctx, "pending")
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return utils.GetStackWithError(fmt.Errorf("failed to process pending reservations: %w", err))
	}

	// 通知レコードを作成
	if err := s.createNotifications(ctx, events); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return utils.GetStackWithError(fmt.Errorf("failed to create notifications: %w", err))
	}

	duration := time.Since(startTime)
	span.SetAttributes(attribute.String("duration", duration.String()))
	log.Printf("Reservation batch process completed successfully. Duration: %v", duration)
	return nil
}

// processReservationsByStatus は、指定されたステータスの予約を処理します
func (s *ReservationBatchService) processReservationsByStatus(ctx context.Context, status string) ([]model.ReservationEvent, error) {
	ctx, span := reservationTracer.Start(ctx, "ReservationBatchService.processReservationsByStatus")
	defer span.End()

	// 指定されたステータスの予約を取得
	reservations, err := s.reservationRepo.GetReservationsByStatus(ctx, status)
	if err != nil {
		return nil, fmt.Errorf("failed to get reservations with status %s: %w", status, err)
	}

	log.Printf("Found %d reservations with status %s", len(reservations), status)
	span.SetAttributes(attribute.Int("reservation_count", len(reservations)))

	var events []model.ReservationEvent

	for _, reservation := range reservations {
		tx, err := s.reservationRepo.BeginTx()
		if err != nil {
			log.Printf("Failed to begin transaction for reservation %d: %v",
				reservation.ReservationID, err)
			continue
		}

		exists, err := s.reservationRepo.CheckExistingReservation(ctx, reservation.PetID)
		if err != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				log.Printf("Failed to rollback transaction for reservation %d: %v",
					reservation.ReservationID, rollbackErr)
			}
			log.Printf("Failed to check existing reservation for pet %s: %v",
				reservation.PetID, err)
			continue
		}

		if exists {
			if err := s.reservationRepo.UpdateStatus(ctx, tx, reservation.ReservationID, "cancelled"); err != nil {
				if rollbackErr := tx.Rollback(); rollbackErr != nil {
					log.Printf("Failed to rollback transaction for reservation %d: %v",
						reservation.ReservationID, rollbackErr)
				}
				log.Printf("Failed to update reservation status to cancelled: %v", err)
				continue
			}
		} else {
			if err := s.reservationRepo.UpdateStatus(ctx, tx, reservation.ReservationID, "confirmed"); err != nil {
				if rollbackErr := tx.Rollback(); rollbackErr != nil {
					log.Printf("Failed to rollback transaction for reservation %d: %v",
						reservation.ReservationID, rollbackErr)
				}
				log.Printf("Failed to update reservation status to confirmed: %v", err)
				continue
			}
		}

		if err := tx.Commit(); err != nil {
			log.Printf("Failed to commit transaction for reservation %d: %v",
				reservation.ReservationID, err)
			continue
		}

		events = append(events, model.ReservationEvent{
			UserID:    reservation.UserID,
			DateTime:  reservation.ReservationDateTime,
			PetID:     reservation.PetID,
			CreatedAt: reservation.CreatedAt,
		})
	}

	return events, nil
}

// createNotifications はイベントから通知レコードを作成します
func (s *ReservationBatchService) createNotifications(ctx context.Context, events []model.ReservationEvent) error {
	ctx, span := reservationTracer.Start(ctx, "ReservationBatchService.createNotifications")
	defer span.End()

	if len(events) == 0 {
		log.Println("No events to create notifications for")
		return nil
	}

	// イベントを通知形式に変換
	notifications := make([]model.Notification, len(events))
	for i, event := range events {
		notifications[i] = model.NewReservationNotification(event)
	}

	// ペット名を取得（N+1防止）
	petNameMap, err := s.getPetNameMap(ctx, notifications)
	if err != nil {
		return fmt.Errorf("failed to get pet name map: %w", err)
	}

	// 通知をレコードに変換
	records := make([]model.NotificationRecord, len(notifications))
	for i, notification := range notifications {
		record, err := notification.ToNotificationRecord(petNameMap)
		if err != nil {
			return fmt.Errorf("failed to convert notification to record: %w", err)
		}
		records[i] = *record
	}

	// 通知レコードを作成
	if err := s.notificationRepo.CreateNotifications(ctx, records); err != nil {
		return fmt.Errorf("failed to create notification records: %w", err)
	}

	span.SetAttributes(attribute.Int("notification_count", len(records)))
	log.Printf("Created %d notification records", len(records))
	return nil
}

// getPetNameMap は通知データからペット名のMapを取得します
func (s *ReservationBatchService) getPetNameMap(ctx context.Context, notifications []model.Notification) (map[string]string, error) {
	petIDs := make([]string, 0)
	petNameMap := make(map[string]string)
	for _, notification := range notifications {
		data, ok := notification.Data.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("invalid notification data format")
		}

		petID, ok := data["pet_id"].(string)
		if !ok {
			return nil, fmt.Errorf("pet_id is not a string")
		}

		if slices.Contains(petIDs, petID) {
			continue
		}
		petIDs = append(petIDs, petID)
	}

	for _, petID := range petIDs {
		petName, err := s.petRepo.GetNameByID(ctx, petID)
		if err != nil {
			return nil, err
		}
		petNameMap[petID] = petName
	}

	return petNameMap, nil
}
