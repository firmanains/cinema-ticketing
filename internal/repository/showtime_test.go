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

func TestShowtimeRepository_Create(t *testing.T) {
	newShowtime := func() *domain.Showtime {
		return &domain.Showtime{
			ID:          uuid.New(),
			MovieTitle:  "Inception",
			StartTime:   time.Now(),
			EndTime:     time.Now().Add(2 * time.Hour),
			Price:       50000,
			TotalSeats:  100,
			BookedSeats: 0,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}
	}
	tests := []struct {
		name       string
		showtime   *domain.Showtime
		mockSetup  func(m sqlmock.Sqlmock)
		wantErr    bool
		wantErrMsg string
	}{
		{
			name:     "success",
			showtime: newShowtime(),
			mockSetup: func(m sqlmock.Sqlmock) {
				m.ExpectExec("INSERT INTO showtimes").
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			wantErr: false,
		},
		{
			name:     "invalid time range",
			showtime: newShowtime(),
			mockSetup: func(m sqlmock.Sqlmock) {
				m.ExpectExec("INSERT INTO showtimes").
					WillReturnError(fmt.Errorf("invalid time range constraint violated"))
			},
			wantErr:    true,
			wantErrMsg: "start time must be before end time",
		},
		{
			name:     "db error",
			showtime: newShowtime(),
			mockSetup: func(m sqlmock.Sqlmock) {
				m.ExpectExec("INSERT INTO showtimes").
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

			repo := repository.NewShowtimeRepository(sqlxDB)
			err = repo.Create(context.Background(), tt.showtime)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.wantErrMsg != "" {
					assert.Contains(t, err.Error(), tt.wantErrMsg)
				}
			} else {
				assert.NoError(t, err)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestShowtimeRepository_FindAll(t *testing.T) {
	id := uuid.New()
	now := time.Now()
	tests := []struct {
		name       string
		page       int
		limit      int
		mockSetup  func(m sqlmock.Sqlmock)
		wantErr    bool
		wantTotal  int
		wantLen    int
	}{
		{
			name:  "success with pagination",
			page:  1,
			limit: 10,
			mockSetup: func(m sqlmock.Sqlmock) {
				countRows := sqlmock.NewRows([]string{"count"}).AddRow(1)
				m.ExpectQuery("SELECT COUNT").WillReturnRows(countRows)

				rows := sqlmock.NewRows([]string{"id", "movie_title", "start_time", "end_time", "price", "total_seats", "booked_seats", "created_at", "updated_at"}).
					AddRow(id, "Inception", now, now.Add(2*time.Hour), 50000.0, 100, 0, now, now)
				m.ExpectQuery("SELECT \\* FROM showtimes ORDER BY").
					WithArgs(10, 0).
					WillReturnRows(rows)
			},
			wantErr:   false,
			wantTotal: 1,
			wantLen:   1,
		},
		{
			name:  "count query error",
			page:  1,
			limit: 10,
			mockSetup: func(m sqlmock.Sqlmock) {
				m.ExpectQuery("SELECT COUNT").
					WillReturnError(fmt.Errorf("db error"))
			},
			wantErr: true,
		},
		{
			name:  "select query error",
			page:  1,
			limit: 10,
			mockSetup: func(m sqlmock.Sqlmock) {
				countRows := sqlmock.NewRows([]string{"count"}).AddRow(1)
				m.ExpectQuery("SELECT COUNT").WillReturnRows(countRows)

				m.ExpectQuery("SELECT \\* FROM showtimes ORDER BY").
					WithArgs(10, 0).
					WillReturnError(fmt.Errorf("db error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			assert.NoError(t, err)
			defer db.Close()

			sqlxDB := sqlx.NewDb(db, "postgres")
			tt.mockSetup(mock)

			repo := repository.NewShowtimeRepository(sqlxDB)
			showtimes, total, err := repo.FindAll(context.Background(), tt.page, tt.limit)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantTotal, total)
				assert.Len(t, showtimes, tt.wantLen)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestShowtimeRepository_FindByID(t *testing.T) {
	id := uuid.New()
	now := time.Now()
	tests := []struct {
		name       string
		mockSetup  func(m sqlmock.Sqlmock)
		wantErr    bool
		wantErrMsg string
	}{
		{
			name: "found",
			mockSetup: func(m sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "movie_title", "start_time", "end_time", "price", "total_seats", "booked_seats", "created_at", "updated_at"}).
					AddRow(id, "Inception", now, now.Add(2*time.Hour), 50000.0, 100, 0, now, now)
				m.ExpectQuery("SELECT \\* FROM showtimes").
					WithArgs(id).
					WillReturnRows(rows)
			},
			wantErr: false,
		},
		{
			name: "not found",
			mockSetup: func(m sqlmock.Sqlmock) {
				m.ExpectQuery("SELECT \\* FROM showtimes").
					WithArgs(id).
					WillReturnError(sql.ErrNoRows)
			},
			wantErr:    true,
			wantErrMsg: "showtime not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			assert.NoError(t, err)
			defer db.Close()

			sqlxDB := sqlx.NewDb(db, "postgres")
			tt.mockSetup(mock)

			repo := repository.NewShowtimeRepository(sqlxDB)
			result, err := repo.FindByID(context.Background(), id)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErrMsg)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, id, result.ID)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestShowtimeRepository_Update(t *testing.T) {
	id := uuid.New()
	now := time.Now()
	title := "Interstellar"
	tests := []struct {
		name       string
		req        domain.UpdateShowtimeRequest
		mockSetup  func(m sqlmock.Sqlmock)
		wantErr    bool
		wantErrMsg string
	}{
		{
			name: "success",
			req:  domain.UpdateShowtimeRequest{MovieTitle: &title},
			mockSetup: func(m sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "movie_title", "start_time", "end_time", "price", "total_seats", "booked_seats", "created_at", "updated_at"}).
					AddRow(id, "Inception", now, now.Add(2*time.Hour), 50000.0, 100, 0, now, now)
				m.ExpectQuery("SELECT \\* FROM showtimes").
					WithArgs(id).
					WillReturnRows(rows)
				m.ExpectExec("UPDATE showtimes").
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			wantErr: false,
		},
		{
			name: "not found",
			req:  domain.UpdateShowtimeRequest{MovieTitle: &title},
			mockSetup: func(m sqlmock.Sqlmock) {
				m.ExpectQuery("SELECT \\* FROM showtimes").
					WithArgs(id).
					WillReturnError(sql.ErrNoRows)
			},
			wantErr:    true,
			wantErrMsg: "showtime not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			assert.NoError(t, err)
			defer db.Close()

			sqlxDB := sqlx.NewDb(db, "postgres")
			tt.mockSetup(mock)

			repo := repository.NewShowtimeRepository(sqlxDB)
			result, err := repo.Update(context.Background(), id, tt.req)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErrMsg)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestShowtimeRepository_Delete(t *testing.T) {
	id := uuid.New()
	tests := []struct {
		name       string
		mockSetup  func(m sqlmock.Sqlmock)
		wantErr    bool
		wantErrMsg string
	}{
		{
			name: "success",
			mockSetup: func(m sqlmock.Sqlmock) {
				m.ExpectExec("DELETE FROM showtimes").
					WithArgs(id).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
			wantErr: false,
		},
		{
			name: "not found",
			mockSetup: func(m sqlmock.Sqlmock) {
				m.ExpectExec("DELETE FROM showtimes").
					WithArgs(id).
					WillReturnResult(sqlmock.NewResult(0, 0))
			},
			wantErr:    true,
			wantErrMsg: "showtime not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			assert.NoError(t, err)
			defer db.Close()

			sqlxDB := sqlx.NewDb(db, "postgres")
			tt.mockSetup(mock)

			repo := repository.NewShowtimeRepository(sqlxDB)
			err = repo.Delete(context.Background(), id)

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
