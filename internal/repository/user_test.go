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

func TestUserRepository_Create(t *testing.T) {
	tests := []struct {
		name       string
		user       *domain.User
		mockSetup  func(m sqlmock.Sqlmock)
		wantErr    bool
		wantErrMsg string
	}{
		{
			name: "success",
			user: &domain.User{
				ID:           uuid.New(),
				Name:         "John Doe",
				Email:        "john@example.com",
				PasswordHash: "hash",
				CreatedAt:    time.Now(),
				UpdatedAt:    time.Now(),
			},
			mockSetup: func(m sqlmock.Sqlmock) {
				m.ExpectExec("INSERT INTO users").
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			wantErr: false,
		},
		{
			name: "duplicate email",
			user: &domain.User{
				ID:           uuid.New(),
				Name:         "John Doe",
				Email:        "john@example.com",
				PasswordHash: "hash",
				CreatedAt:    time.Now(),
				UpdatedAt:    time.Now(),
			},
			mockSetup: func(m sqlmock.Sqlmock) {
				m.ExpectExec("INSERT INTO users").
					WillReturnError(fmt.Errorf("duplicate key value violates unique constraint"))
			},
			wantErr:    true,
			wantErrMsg: "email already registered",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			assert.NoError(t, err)
			defer db.Close()

			sqlxDB := sqlx.NewDb(db, "postgres")
			tt.mockSetup(mock)

			repo := repository.NewUserRepository(sqlxDB)
			err = repo.Create(context.Background(), tt.user)

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

func TestUserRepository_FindByEmail(t *testing.T) {
	tests := []struct {
		name       string
		email      string
		mockSetup  func(m sqlmock.Sqlmock)
		wantErr    bool
		wantErrMsg string
	}{
		{
			name:  "found",
			email: "john@example.com",
			mockSetup: func(m sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "name", "email", "password_hash", "created_at", "updated_at"}).
					AddRow(uuid.New(), "John Doe", "john@example.com", "hash", time.Now(), time.Now())
				m.ExpectQuery("SELECT \\* FROM users").
					WithArgs("john@example.com").
					WillReturnRows(rows)
			},
			wantErr: false,
		},
		{
			name:  "not found",
			email: "ghost@example.com",
			mockSetup: func(m sqlmock.Sqlmock) {
				m.ExpectQuery("SELECT \\* FROM users").
					WithArgs("ghost@example.com").
					WillReturnError(sql.ErrNoRows)
			},
			wantErr:    true,
			wantErrMsg: "user not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			assert.NoError(t, err)
			defer db.Close()

			sqlxDB := sqlx.NewDb(db, "postgres")
			tt.mockSetup(mock)

			repo := repository.NewUserRepository(sqlxDB)
			user, err := repo.FindByEmail(context.Background(), tt.email)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErrMsg)
				assert.Nil(t, user)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, user)
				assert.Equal(t, tt.email, user.Email)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
