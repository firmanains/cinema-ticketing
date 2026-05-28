package domain

import (
	"time"

	"github.com/google/uuid"
)

type Showtime struct {
	ID          uuid.UUID `db:"id"           json:"id"`
	MovieTitle  string    `db:"movie_title"  json:"movie_title"`
	StartTime   time.Time `db:"start_time"   json:"start_time"`
	EndTime     time.Time `db:"end_time"     json:"end_time"`
	Price       float64   `db:"price"        json:"price"`
	TotalSeats  int       `db:"total_seats"  json:"total_seats"`
	BookedSeats int       `db:"booked_seats" json:"booked_seats"`
	CreatedAt   time.Time `db:"created_at"   json:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"   json:"updated_at"`
}
