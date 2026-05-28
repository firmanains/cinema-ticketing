package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Showtime struct {
	ID          uuid.UUID `db:"id"`
	MovieTitle  string    `db:"movie_title"`
	StartTime   time.Time `db:"start_time"`
	EndTime     time.Time `db:"end_time"`
	Price       float64   `db:"price"`
	TotalSeats  int       `db:"total_seats"`
	BookedSeats int       `db:"booked_seats"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

type ShowtimeRepository interface {
	Create(ctx context.Context, showtime *Showtime) error
	FindAll(ctx context.Context, page, limit int) ([]Showtime, int, error)
	FindByID(ctx context.Context, id uuid.UUID) (*Showtime, error)
	Update(ctx context.Context, id uuid.UUID, req UpdateShowtimeRequest) (*Showtime, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type ShowtimeService interface {
	Create(ctx context.Context, req CreateShowtimeRequest) (*Showtime, error)
	GetAll(ctx context.Context, page, limit int) (*PaginatedResult[Showtime], error)
	GetByID(ctx context.Context, id uuid.UUID) (*Showtime, error)
	Update(ctx context.Context, id uuid.UUID, req UpdateShowtimeRequest) (*Showtime, error)
	Delete(ctx context.Context, id uuid.UUID) error
}
