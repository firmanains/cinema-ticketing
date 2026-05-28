package domain

import "time"

type CreateShowtimeRequest struct {
	MovieTitle string    `json:"movie_title" validate:"required"`
	StartTime  time.Time `json:"start_time" validate:"required"`
	EndTime    time.Time `json:"end_time" validate:"required"`
	Price      float64   `json:"price" validate:"required,gt=0"`
	TotalSeats int       `json:"total_seats" validate:"required,gt=0"`
}

type UpdateShowtimeRequest struct {
	MovieTitle *string    `json:"movie_title"`
	StartTime  *time.Time `json:"start_time"`
	EndTime    *time.Time `json:"end_time"`
	Price      *float64   `json:"price" validate:"omitempty,gt=0"`
	TotalSeats *int       `json:"total_seats" validate:"omitempty,gt=0"`
}
