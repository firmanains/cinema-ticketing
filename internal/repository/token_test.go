package repository_test

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"

	"github.com/firmanains/cinema-ticketing/internal/domain"
	"github.com/firmanains/cinema-ticketing/internal/repository"
)

func TestRefreshTokenRepository_FindByTokenHash(t *testing.T) {
	tests := []struct {
		name       string
		tokenHash  string
		mockSetup  func(mock sqlmock.Sqlmock, tokenHash string)
		wantErr    bool
		wantErrMsg string
	}{
		{
			name:      "found",
			tokenHash: "abc123hash",
			mockSetup: func(m sqlmock.Sqlmock, tokenHash string) {
				rows := sqlmock.NewRows([]string{"id", "user_id", "token_hash", "expires_at", "revoked_at", "created_at"}).
					AddRow(uuid.New(), uuid.New(), tokenHash, time.Now().Add(24*time.Hour), nil, time.Now())
				m.ExpectQuery("SELECT \\* FROM refresh_tokens").
					WithArgs(tokenHash).
					WillReturnRows(rows)
			},
			wantErr: false,
		},
		{
			name:      "not found",
			tokenHash: "invalidhash",
			mockSetup: func(m sqlmock.Sqlmock, tokenHash string) {
				m.ExpectQuery("SELECT \\* FROM refresh_tokens").
					WithArgs(tokenHash).
					WillReturnError(sql.ErrNoRows)
			},
			wantErr:    true,
			wantErrMsg: "token is invalid or expired",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			assert.NoError(t, err)
			defer db.Close()

			sqlxDB := sqlx.NewDb(db, "postgres")
			tt.mockSetup(mock, tt.tokenHash)

			repo := repository.NewRefreshTokenRepository(sqlxDB)
			token, err := repo.FindByTokenHash(context.Background(), tt.tokenHash)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErrMsg)
				assert.Nil(t, token)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, token)
				assert.Equal(t, tt.tokenHash, token.TokenHash)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestRefreshTokenRepository_Store(t *testing.T) {
	tests := []struct {
		name      string
		token     *domain.RefreshToken
		mockSetup func(mock sqlmock.Sqlmock)
		wantErr   bool
	}{
		{
			name: "success",
			token: &domain.RefreshToken{
				ID:        uuid.New(),
				UserID:    uuid.New(),
				TokenHash: "somehash",
				ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
				CreatedAt: time.Now(),
			},
			mockSetup: func(m sqlmock.Sqlmock) {
				m.ExpectExec("INSERT INTO refresh_tokens").
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			assert.NoError(t, err)
			defer db.Close()

			sqlxDB := sqlx.NewDb(db, "postgres")
			tt.mockSetup(mock)

			repo := repository.NewRefreshTokenRepository(sqlxDB)
			err = repo.Store(context.Background(), tt.token)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestRefreshTokenRepository_RevokeByTokenHash(t *testing.T) {
	tests := []struct {
		name       string
		tokenHash  string
		mockSetup  func(m sqlmock.Sqlmock)
		wantErr    bool
		wantErrMsg string
	}{
		{
			name:      "success",
			tokenHash: "somehash",
			mockSetup: func(m sqlmock.Sqlmock) {
				m.ExpectExec("UPDATE refresh_tokens").
					WithArgs("somehash").
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
			wantErr: false,
		},
		{
			name:      "db error",
			tokenHash: "somehash",
			mockSetup: func(m sqlmock.Sqlmock) {
				m.ExpectExec("UPDATE refresh_tokens").
					WithArgs("somehash").
					WillReturnError(fmt.Errorf("connection refused"))
			},
			wantErr:    true,
			wantErrMsg: "connection refused",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			assert.NoError(t, err)
			defer db.Close()

			sqlxDB := sqlx.NewDb(db, "postgres")
			tt.mockSetup(mock)

			repo := repository.NewRefreshTokenRepository(sqlxDB)
			err = repo.RevokeByTokenHash(context.Background(), tt.tokenHash)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErrMsg)
			} else {
				assert.NoError(t, err)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
