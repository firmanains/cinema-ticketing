package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type RefreshToken struct {
	ID        uuid.UUID  `db:"id"`
	UserID    uuid.UUID  `db:"user_id"`
	TokenHash string     `db:"token_hash"`
	ExpiresAt time.Time  `db:"expires_at"`
	RevokedAt *time.Time `db:"revoked_at"`
	CreatedAt time.Time  `db:"created_at"`
}

type RefreshTokenRepository interface {
	Store(ctx context.Context, token *RefreshToken) error
	FindByTokenHash(ctx context.Context, tokenHash string) (*RefreshToken, error)
}
