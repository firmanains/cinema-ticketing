package repository

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/firmanains/cinema-ticketing/internal/domain"
)

type showtimeRepository struct {
	db *sqlx.DB
}

func NewShowtimeRepository(db *sqlx.DB) domain.ShowtimeRepository {
	return &showtimeRepository{db: db}
}

func (r *showtimeRepository) FindAll(ctx context.Context, page, limit int) ([]domain.Showtime, int, error) {
	offset := (page - 1) * limit

	var total int
	if err := r.db.GetContext(ctx, &total, "SELECT COUNT(*) FROM showtimes"); err != nil {
		return nil, 0, err
	}

	var showtimes []domain.Showtime
	err := r.db.SelectContext(ctx, &showtimes,
		"SELECT * FROM showtimes ORDER BY start_time ASC LIMIT $1 OFFSET $2",
		limit, offset,
	)
	if err != nil {
		return nil, 0, err
	}

	return showtimes, total, nil
}

func (r *showtimeRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Showtime, error) {
	var s domain.Showtime
	err := r.db.GetContext(ctx, &s, "SELECT * FROM showtimes WHERE id = $1", id)
	if err != nil {
		return nil, errors.New("showtime not found")
	}
	return &s, nil
}

func (r *showtimeRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result, err := r.db.ExecContext(ctx, "DELETE FROM showtimes WHERE id = $1", id)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return errors.New("showtime not found")
	}
	return nil
}

func (r *showtimeRepository) Update(ctx context.Context, id uuid.UUID, req domain.UpdateShowtimeRequest) (*domain.Showtime, error) {
	existing, err := r.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.MovieTitle != nil {
		existing.MovieTitle = *req.MovieTitle
	}
	if req.StartTime != nil {
		existing.StartTime = *req.StartTime
	}
	if req.EndTime != nil {
		existing.EndTime = *req.EndTime
	}
	if req.Price != nil {
		existing.Price = *req.Price
	}
	if req.TotalSeats != nil {
		existing.TotalSeats = *req.TotalSeats
	}

	if req.MovieTitle == nil && req.StartTime == nil && req.EndTime == nil && req.Price == nil && req.TotalSeats == nil {
		return nil, errors.New("no fields to update")
	}

	existing.UpdatedAt = time.Now()

	_, err = r.db.NamedExecContext(ctx, `
		UPDATE showtimes
		SET movie_title = :movie_title, start_time = :start_time, end_time = :end_time,
		    price = :price, total_seats = :total_seats, updated_at = :updated_at
		WHERE id = :id
	`, existing)
	if err != nil {
		return nil, err
	}

	return existing, nil
}

func (r *showtimeRepository) Create(ctx context.Context, s *domain.Showtime) error {
	query := `
		INSERT INTO showtimes (id, movie_title, start_time, end_time, price, total_seats, booked_seats, created_at, updated_at)
		VALUES (:id, :movie_title, :start_time, :end_time, :price, :total_seats, :booked_seats, :created_at, :updated_at)
	`
	_, err := r.db.NamedExecContext(ctx, query, s)
	if err != nil {
		if strings.Contains(err.Error(), "invalid time range") {
			return errors.New("start time must be before end time")
		}
		return err
	}
	return nil
}
