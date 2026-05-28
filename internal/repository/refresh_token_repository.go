package repository

import (
	"context"

	"github.com/jmoiron/sqlx"

	"github.com/firmanains/cinema-ticketing/internal/domain"
)

type refreshTokenRepository struct {
	db *sqlx.DB
}

func NewRefreshTokenRepository(db *sqlx.DB) domain.RefreshTokenRepository {
	return &refreshTokenRepository{db: db}
}

func (r *refreshTokenRepository) Store(ctx context.Context, token *domain.RefreshToken) error {
	query := `
		INSERT INTO refresh_tokens (id, user_id, token_hash, expires_at, created_at)
		VALUES (:id, :user_id, :token_hash, :expires_at, :created_at)
	`
	_, err := r.db.NamedExecContext(ctx, query, token)
	return err
}
