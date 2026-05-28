package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/firmanains/cinema-ticketing/internal/domain"
	"github.com/firmanains/cinema-ticketing/internal/mock"
	"github.com/firmanains/cinema-ticketing/internal/service"
)

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
