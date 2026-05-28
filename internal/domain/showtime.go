package domain

import (
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
