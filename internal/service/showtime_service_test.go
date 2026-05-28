package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/firmanains/cinema-ticketing/internal/domain"
	"github.com/firmanains/cinema-ticketing/internal/mock"
	"github.com/firmanains/cinema-ticketing/internal/service"
)

func TestShowtimeService_GetAll(t *testing.T) {
	tests := []struct {
		name      string
		page      int
		limit     int
		mockSetup func(repo *mock.MockShowtimeRepository)
		wantErr   bool
		wantTotal int
	}{
		{
			name:  "success",
			page:  1,
			limit: 10,
			mockSetup: func(repo *mock.MockShowtimeRepository) {
				repo.EXPECT().
					FindAll(gomock.Any(), 1, 10).
					Return([]domain.Showtime{{MovieTitle: "Inception"}}, 1, nil)
			},
			wantErr:   false,
			wantTotal: 1,
		},
		{
			name:  "repository error",
			page:  1,
			limit: 10,
			mockSetup: func(repo *mock.MockShowtimeRepository) {
				repo.EXPECT().
					FindAll(gomock.Any(), 1, 10).
					Return(nil, 0, errors.New("db error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo := mock.NewMockShowtimeRepository(ctrl)
			tt.mockSetup(repo)

			svc := service.NewShowtimeService(repo)
			result, err := svc.GetAll(context.Background(), tt.page, tt.limit)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.wantTotal, result.Total)
			}
		})
	}
}

func TestShowtimeService_GetByID(t *testing.T) {
	id := uuid.New()
	tests := []struct {
		name       string
		mockSetup  func(repo *mock.MockShowtimeRepository)
		wantErr    bool
		wantErrMsg string
	}{
		{
			name: "success",
			mockSetup: func(repo *mock.MockShowtimeRepository) {
				repo.EXPECT().
					FindByID(gomock.Any(), id).
					Return(&domain.Showtime{MovieTitle: "Inception"}, nil)
			},
			wantErr: false,
		},
		{
			name: "not found",
			mockSetup: func(repo *mock.MockShowtimeRepository) {
				repo.EXPECT().
					FindByID(gomock.Any(), id).
					Return(nil, errors.New("showtime not found"))
			},
			wantErr:    true,
			wantErrMsg: "showtime not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo := mock.NewMockShowtimeRepository(ctrl)
			tt.mockSetup(repo)

			svc := service.NewShowtimeService(repo)
			result, err := svc.GetByID(context.Background(), id)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErrMsg)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}
		})
	}
}

func TestShowtimeService_Delete(t *testing.T) {
	id := uuid.New()
	tests := []struct {
		name       string
		mockSetup  func(repo *mock.MockShowtimeRepository)
		wantErr    bool
		wantErrMsg string
	}{
		{
			name: "success",
			mockSetup: func(repo *mock.MockShowtimeRepository) {
				repo.EXPECT().
					Delete(gomock.Any(), id).
					Return(nil)
			},
			wantErr: false,
		},
		{
			name: "not found",
			mockSetup: func(repo *mock.MockShowtimeRepository) {
				repo.EXPECT().
					Delete(gomock.Any(), id).
					Return(errors.New("showtime not found"))
			},
			wantErr:    true,
			wantErrMsg: "showtime not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo := mock.NewMockShowtimeRepository(ctrl)
			tt.mockSetup(repo)

			svc := service.NewShowtimeService(repo)
			err := svc.Delete(context.Background(), id)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErrMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestShowtimeService_Update(t *testing.T) {
	id := uuid.New()
	title := "Interstellar"
	tests := []struct {
		name       string
		req        domain.UpdateShowtimeRequest
		mockSetup  func(repo *mock.MockShowtimeRepository)
		wantErr    bool
		wantErrMsg string
	}{
		{
			name: "success",
			req:  domain.UpdateShowtimeRequest{MovieTitle: &title},
			mockSetup: func(repo *mock.MockShowtimeRepository) {
				repo.EXPECT().
					Update(gomock.Any(), id, gomock.Any()).
					Return(&domain.Showtime{MovieTitle: title}, nil)
			},
			wantErr: false,
		},
		{
			name:       "no fields to update",
			req:        domain.UpdateShowtimeRequest{},
			mockSetup: func(repo *mock.MockShowtimeRepository) {
				repo.EXPECT().
					Update(gomock.Any(), id, gomock.Any()).
					Return(nil, errors.New("no fields to update"))
			},
			wantErr:    true,
			wantErrMsg: "no fields to update",
		},
		{
			name: "not found",
			req:  domain.UpdateShowtimeRequest{MovieTitle: &title},
			mockSetup: func(repo *mock.MockShowtimeRepository) {
				repo.EXPECT().
					Update(gomock.Any(), id, gomock.Any()).
					Return(nil, errors.New("showtime not found"))
			},
			wantErr:    true,
			wantErrMsg: "showtime not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo := mock.NewMockShowtimeRepository(ctrl)
			tt.mockSetup(repo)

			svc := service.NewShowtimeService(repo)
			result, err := svc.Update(context.Background(), id, tt.req)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErrMsg)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}
		})
	}
}

func TestShowtimeService_Create(t *testing.T) {
	baseTime := time.Now()
	tests := []struct {
		name       string
		input      domain.CreateShowtimeRequest
		mockSetup  func(repo *mock.MockShowtimeRepository)
		wantErr    bool
		wantErrMsg string
	}{
		{
			name: "success",
			input: domain.CreateShowtimeRequest{
				MovieTitle: "Inception",
				StartTime:  baseTime,
				EndTime:    baseTime.Add(2 * time.Hour),
				Price:      50000,
				TotalSeats: 100,
			},
			mockSetup: func(repo *mock.MockShowtimeRepository) {
				repo.EXPECT().
					Create(gomock.Any(), gomock.Any()).
					Return(nil)
			},
			wantErr: false,
		},
		{
			name: "invalid time range",
			input: domain.CreateShowtimeRequest{
				MovieTitle: "Inception",
				StartTime:  baseTime.Add(2 * time.Hour),
				EndTime:    baseTime,
				Price:      50000,
				TotalSeats: 100,
			},
			mockSetup:  func(repo *mock.MockShowtimeRepository) {},
			wantErr:    true,
			wantErrMsg: "start time must be before end time",
		},
		{
			name: "repository error",
			input: domain.CreateShowtimeRequest{
				MovieTitle: "Inception",
				StartTime:  baseTime,
				EndTime:    baseTime.Add(2 * time.Hour),
				Price:      50000,
				TotalSeats: 100,
			},
			mockSetup: func(repo *mock.MockShowtimeRepository) {
				repo.EXPECT().
					Create(gomock.Any(), gomock.Any()).
					Return(errors.New("db error"))
			},
			wantErr:    true,
			wantErrMsg: "db error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo := mock.NewMockShowtimeRepository(ctrl)
			tt.mockSetup(repo)

			svc := service.NewShowtimeService(repo)
			result, err := svc.Create(context.Background(), tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErrMsg)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.input.MovieTitle, result.MovieTitle)
			}
		})
	}
}
