package repository

import (
	"context"
	"fmt"
)

// PetRepository はペット情報の永続化を担当するインターフェースです
type PetRepository interface {
	GetNameByID(ctx context.Context, petID string) (string, error)
}

// PetRepositoryImpl はPetRepositoryの実装です
type PetRepositoryImpl struct {
	db *DB
}

// NewPetRepository は新しいPetRepositoryを作成します
func NewPetRepository(db *DB) PetRepository {
	return &PetRepositoryImpl{
		db: db,
	}
}

// GetNameByID は指定されたペットIDからペット名を取得します
func (r *PetRepositoryImpl) GetNameByID(ctx context.Context, petID string) (string, error) {
	query := `
		SELECT name
		FROM pets
		WHERE id = $1`

	var name string
	err := r.db.QueryRowContext(ctx, query, petID).Scan(&name)
	if err != nil {
		return "", fmt.Errorf("failed to get pet name: %w", err)
	}

	return name, nil
}
