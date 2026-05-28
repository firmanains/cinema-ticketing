package service

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/firmanains/cinema-ticketing/internal/constant"
	"github.com/firmanains/cinema-ticketing/internal/domain"
)

type ShowtimeRepository interface {
	Create(ctx context.Context, showtime *domain.Showtime) error
	FindAll(ctx context.Context, page, limit int) ([]domain.Showtime, int, error)
	FindByID(ctx context.Context, id uuid.UUID) (*domain.Showtime, error)
	Update(ctx context.Context, id uuid.UUID, req domain.UpdateShowtimeRequest) (*domain.Showtime, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type showtimeService struct {
	repo ShowtimeRepository
}

func NewShowtimeService(repo ShowtimeRepository) *showtimeService {
	return &showtimeService{repo: repo}
}

func (s *showtimeService) Create(ctx context.Context, req domain.CreateShowtimeRequest) (*domain.Showtime, error) {
	if !req.EndTime.After(req.StartTime) {
		return nil, constant.ErrInvalidTimeRange
	}

	now := time.Now()
	showtime := &domain.Showtime{
		ID:          uuid.New(),
		MovieTitle:  req.MovieTitle,
		StartTime:   req.StartTime,
		EndTime:     req.EndTime,
		Price:       req.Price,
		TotalSeats:  req.TotalSeats,
		BookedSeats: 0,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := s.repo.Create(ctx, showtime); err != nil {
		return nil, err
	}

	return showtime, nil
}

func (s *showtimeService) GetAll(ctx context.Context, page, limit int) (*domain.PaginatedResult[domain.Showtime], error) {
	items, total, err := s.repo.FindAll(ctx, page, limit)
	if err != nil {
		return nil, err
	}
	return &domain.PaginatedResult[domain.Showtime]{
		Items: items,
		Total: total,
		Page:  page,
		Limit: limit,
	}, nil
}

func (s *showtimeService) GetByID(ctx context.Context, id uuid.UUID) (*domain.Showtime, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *showtimeService) Update(ctx context.Context, id uuid.UUID, req domain.UpdateShowtimeRequest) (*domain.Showtime, error) {
	if req.StartTime != nil && req.EndTime != nil && !req.EndTime.After(*req.StartTime) {
		return nil, constant.ErrInvalidTimeRange
	}

	return s.repo.Update(ctx, id, req)
}

func (s *showtimeService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}
