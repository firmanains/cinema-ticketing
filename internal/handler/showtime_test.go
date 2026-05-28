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
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/firmanains/cinema-ticketing/internal/domain"
	"github.com/firmanains/cinema-ticketing/internal/handler"
	"github.com/firmanains/cinema-ticketing/internal/mock"
)

func TestShowtimeHandler_GetAll(t *testing.T) {
	tests := []struct {
		name        string
		query       string
		mockSetup   func(svc *mock.MockShowtimeService)
		wantStatus  int
		wantSuccess bool
	}{
		{
			name:  "success",
			query: "?page=1&limit=10",
			mockSetup: func(svc *mock.MockShowtimeService) {
				svc.EXPECT().
					GetAll(gomock.Any(), 1, 10).
					Return(&domain.PaginatedResult[domain.Showtime]{
						Items: []domain.Showtime{{MovieTitle: "Inception"}},
						Total: 1, Page: 1, Limit: 10,
					}, nil)
			},
			wantStatus:  fiber.StatusOK,
			wantSuccess: true,
		},
		{
			name:  "service error",
			query: "?page=1&limit=10",
			mockSetup: func(svc *mock.MockShowtimeService) {
				svc.EXPECT().
					GetAll(gomock.Any(), 1, 10).
					Return(nil, errors.New("db error"))
			},
			wantStatus:  fiber.StatusInternalServerError,
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
			app.Get("/api/v1/showtimes", h.GetAll)

			req := httptest.NewRequest(http.MethodGet, "/api/v1/showtimes"+tt.query, nil)
			resp, err := app.Test(req)
			assert.NoError(t, err)
			assert.Equal(t, tt.wantStatus, resp.StatusCode)
		})
	}
}

func TestShowtimeHandler_GetByID(t *testing.T) {
	id := uuid.New()
	tests := []struct {
		name        string
		id          string
		mockSetup   func(svc *mock.MockShowtimeService)
		wantStatus  int
		wantSuccess bool
	}{
		{
			name: "success",
			id:   id.String(),
			mockSetup: func(svc *mock.MockShowtimeService) {
				svc.EXPECT().
					GetByID(gomock.Any(), id).
					Return(&domain.Showtime{MovieTitle: "Inception"}, nil)
			},
			wantStatus:  fiber.StatusOK,
			wantSuccess: true,
		},
		{
			name: "not found",
			id:   id.String(),
			mockSetup: func(svc *mock.MockShowtimeService) {
				svc.EXPECT().
					GetByID(gomock.Any(), id).
					Return(nil, errors.New("showtime not found"))
			},
			wantStatus:  fiber.StatusNotFound,
			wantSuccess: false,
		},
		{
			name:        "invalid id",
			id:          "not-a-uuid",
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
			app.Get("/api/v1/showtimes/:id", h.GetByID)

			req := httptest.NewRequest(http.MethodGet, "/api/v1/showtimes/"+tt.id, nil)
			resp, err := app.Test(req)
			assert.NoError(t, err)
			assert.Equal(t, tt.wantStatus, resp.StatusCode)
		})
	}
}

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

func TestShowtimeHandler_Update(t *testing.T) {
	id := uuid.New()
	title := "Interstellar"
	baseTime := time.Now().UTC().Truncate(time.Second)
	tests := []struct {
		name        string
		id          string
		body        map[string]interface{}
		mockSetup   func(svc *mock.MockShowtimeService)
		wantStatus  int
		wantSuccess bool
	}{
		{
			name: "success",
			id:   id.String(),
			body: map[string]interface{}{"movie_title": title},
			mockSetup: func(svc *mock.MockShowtimeService) {
				svc.EXPECT().
					Update(gomock.Any(), id, gomock.Any()).
					Return(&domain.Showtime{MovieTitle: title}, nil)
			},
			wantStatus:  fiber.StatusOK,
			wantSuccess: true,
		},
		{
			name: "not found",
			id:   id.String(),
			body: map[string]interface{}{"movie_title": title},
			mockSetup: func(svc *mock.MockShowtimeService) {
				svc.EXPECT().
					Update(gomock.Any(), id, gomock.Any()).
					Return(nil, errors.New("showtime not found"))
			},
			wantStatus:  fiber.StatusNotFound,
			wantSuccess: false,
		},
		{
			name: "no fields to update",
			id:   id.String(),
			body: map[string]interface{}{"movie_title": title},
			mockSetup: func(svc *mock.MockShowtimeService) {
				svc.EXPECT().
					Update(gomock.Any(), id, gomock.Any()).
					Return(nil, errors.New("no fields to update"))
			},
			wantStatus:  fiber.StatusUnprocessableEntity,
			wantSuccess: false,
		},
		{
			name: "invalid time range",
			id:   id.String(),
			body: map[string]interface{}{
				"start_time": baseTime.Add(2 * time.Hour),
				"end_time":   baseTime,
			},
			mockSetup: func(svc *mock.MockShowtimeService) {
				svc.EXPECT().
					Update(gomock.Any(), id, gomock.Any()).
					Return(nil, errors.New("start time must be before end time"))
			},
			wantStatus:  fiber.StatusUnprocessableEntity,
			wantSuccess: false,
		},
		{
			name:        "invalid id",
			id:          "not-a-uuid",
			body:        map[string]interface{}{"movie_title": title},
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
			app.Put("/api/v1/showtimes/:id", h.Update)

			body, _ := json.Marshal(tt.body)
			req := httptest.NewRequest(http.MethodPut, "/api/v1/showtimes/"+tt.id, bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")

			resp, err := app.Test(req)
			assert.NoError(t, err)
			assert.Equal(t, tt.wantStatus, resp.StatusCode)
		})
	}
}

func TestShowtimeHandler_Delete(t *testing.T) {
	id := uuid.New()
	tests := []struct {
		name        string
		id          string
		mockSetup   func(svc *mock.MockShowtimeService)
		wantStatus  int
		wantSuccess bool
	}{
		{
			name: "success",
			id:   id.String(),
			mockSetup: func(svc *mock.MockShowtimeService) {
				svc.EXPECT().
					Delete(gomock.Any(), id).
					Return(nil)
			},
			wantStatus:  fiber.StatusOK,
			wantSuccess: true,
		},
		{
			name: "not found",
			id:   id.String(),
			mockSetup: func(svc *mock.MockShowtimeService) {
				svc.EXPECT().
					Delete(gomock.Any(), id).
					Return(errors.New("showtime not found"))
			},
			wantStatus:  fiber.StatusNotFound,
			wantSuccess: false,
		},
		{
			name:        "invalid id",
			id:          "not-a-uuid",
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
			app.Delete("/api/v1/showtimes/:id", h.Delete)

			req := httptest.NewRequest(http.MethodDelete, "/api/v1/showtimes/"+tt.id, nil)
			resp, err := app.Test(req)
			assert.NoError(t, err)
			assert.Equal(t, tt.wantStatus, resp.StatusCode)
		})
	}
}
