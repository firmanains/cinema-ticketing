package handler_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/firmanains/cinema-ticketing/internal/domain"
	"github.com/firmanains/cinema-ticketing/internal/handler"
	"github.com/firmanains/cinema-ticketing/internal/mock"
)

func TestShowtimeHandler_Create(t *testing.T) {
	baseTime := time.Now().UTC().Truncate(time.Second)
	tests := []struct {
		name        string
		body        map[string]interface{}
		mockSetup   func(svc *mock.MockShowtimeService)
		wantStatus  int
		wantSuccess bool
	}{
		{
			name: "success",
			body: map[string]interface{}{
				"movie_title": "Inception",
				"start_time":  baseTime,
				"end_time":    baseTime.Add(2 * time.Hour),
				"price":       50000.0,
				"total_seats": 100,
			},
			mockSetup: func(svc *mock.MockShowtimeService) {
				svc.EXPECT().
					Create(gomock.Any(), gomock.Any()).
					Return(&domain.Showtime{MovieTitle: "Inception"}, nil)
			},
			wantStatus:  fiber.StatusCreated,
			wantSuccess: true,
		},
		{
			name: "invalid time range",
			body: map[string]interface{}{
				"movie_title": "Inception",
				"start_time":  baseTime.Add(2 * time.Hour),
				"end_time":    baseTime,
				"price":       50000.0,
				"total_seats": 100,
			},
			mockSetup: func(svc *mock.MockShowtimeService) {
				svc.EXPECT().
					Create(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("start time must be before end time"))
			},
			wantStatus:  fiber.StatusUnprocessableEntity,
			wantSuccess: false,
		},
		{
			name:        "missing required field",
			body:        map[string]interface{}{"movie_title": "Inception"},
			mockSetup:   func(svc *mock.MockShowtimeService) {},
			wantStatus:  fiber.StatusUnprocessableEntity,
			wantSuccess: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockSvc := mock.NewMockShowtimeService(ctrl)
			tt.mockSetup(mockSvc)

			app := fiber.New()
			h := handler.NewShowtimeHandler(mockSvc)
			app.Post("/api/v1/showtimes", h.Create)

			body, _ := json.Marshal(tt.body)
			req := httptest.NewRequest(http.MethodPost, "/api/v1/showtimes", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")

			resp, err := app.Test(req)
			assert.NoError(t, err)
			assert.Equal(t, tt.wantStatus, resp.StatusCode)
		})
	}
}
