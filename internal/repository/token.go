package repository

import (
	"context"
	"errors"

	"github.com/jmoiron/sqlx"

	"github.com/firmanains/cinema-ticketing/internal/domain"
)

type refreshTokenRepository struct {
	db *sqlx.DB
}

func NewRefreshTokenRepository(db *sqlx.DB) *refreshTokenRepository {
	return &refreshTokenRepository{db: db}
}

func (r *refreshTokenRepository) FindByTokenHash(ctx context.Context, tokenHash string) (*domain.RefreshToken, error) {
	var token domain.RefreshToken
	err := r.db.GetContext(ctx, &token,
		"SELECT * FROM refresh_tokens WHERE token_hash = $1 AND revoked_at IS NULL AND expires_at > NOW()",
		tokenHash,
	)
	if err != nil {
		return nil, errors.New("token is invalid or expired")
	}
	return &token, nil
}

func (r *refreshTokenRepository) RevokeByTokenHash(ctx context.Context, tokenHash string) error {
	_, err := r.db.ExecContext(ctx,
		"UPDATE refresh_tokens SET revoked_at = NOW() WHERE token_hash = $1",
		tokenHash,
	)
	return err
}

func (r *refreshTokenRepository) Store(ctx context.Context, token *domain.RefreshToken) error {
	query := `
		INSERT INTO refresh_tokens (id, user_id, token_hash, expires_at, created_at)
		VALUES (:id, :user_id, :token_hash, :expires_at, :created_at)
	`
	_, err := r.db.NamedExecContext(ctx, query, token)
	return err
}
