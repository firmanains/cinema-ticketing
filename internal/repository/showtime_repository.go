package repository

import (
	"context"
	"errors"
	"strings"

	"github.com/jmoiron/sqlx"

	"github.com/firmanains/cinema-ticketing/internal/domain"
)

type showtimeRepository struct {
	db *sqlx.DB
}

func NewShowtimeRepository(db *sqlx.DB) domain.ShowtimeRepository {
	return &showtimeRepository{db: db}
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
