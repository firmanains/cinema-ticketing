package handler_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/firmanains/cinema-ticketing/internal/domain"
	"github.com/firmanains/cinema-ticketing/internal/handler"
	"github.com/firmanains/cinema-ticketing/internal/mock"
)

func TestAuthHandler_Register(t *testing.T) {
	tests := []struct {
		name        string
		body        map[string]string
		mockSetup   func(svc *mock.MockUserService)
		wantStatus  int
		wantSuccess bool
	}{
		{
			name: "success",
			body: map[string]string{
				"name":     "John Doe",
				"email":    "john@example.com",
				"password": "secret123",
			},
			mockSetup: func(svc *mock.MockUserService) {
				svc.EXPECT().
					Register(gomock.Any(), gomock.Any()).
					Return(&domain.AuthResponse{
						AccessToken:  "access",
						RefreshToken: "refresh",
					}, nil)
			},
			wantStatus:  fiber.StatusCreated,
			wantSuccess: true,
		},
		{
			name: "duplicate email",
			body: map[string]string{
				"name":     "John Doe",
				"email":    "john@example.com",
				"password": "secret123",
			},
			mockSetup: func(svc *mock.MockUserService) {
				svc.EXPECT().
					Register(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("email already registered"))
			},
			wantStatus:  fiber.StatusConflict,
			wantSuccess: false,
		},
		{
			name: "invalid body",
			body: map[string]string{
				"name": "John Doe",
			},
			mockSetup:   func(svc *mock.MockUserService) {},
			wantStatus:  fiber.StatusUnprocessableEntity,
			wantSuccess: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockSvc := mock.NewMockUserService(ctrl)
			tt.mockSetup(mockSvc)

			app := fiber.New()
			h := handler.NewAuthHandler(mockSvc)
			app.Post("/api/v1/auth/register", h.Register)

			body, _ := json.Marshal(tt.body)
			req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")

			resp, err := app.Test(req)
			assert.NoError(t, err)
			assert.Equal(t, tt.wantStatus, resp.StatusCode)
		})
	}
}

func TestAuthHandler_Login(t *testing.T) {
	tests := []struct {
		name        string
		body        map[string]string
		mockSetup   func(svc *mock.MockUserService)
		wantStatus  int
		wantSuccess bool
	}{
		{
			name: "success",
			body: map[string]string{"email": "john@example.com", "password": "secret123"},
			mockSetup: func(svc *mock.MockUserService) {
				svc.EXPECT().
					Login(gomock.Any(), gomock.Any()).
					Return(&domain.AuthResponse{AccessToken: "access", RefreshToken: "refresh"}, nil)
			},
			wantStatus:  fiber.StatusOK,
			wantSuccess: true,
		},
		{
			name: "invalid credentials",
			body: map[string]string{"email": "john@example.com", "password": "wrong"},
			mockSetup: func(svc *mock.MockUserService) {
				svc.EXPECT().
					Login(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("invalid credentials"))
			},
			wantStatus:  fiber.StatusUnauthorized,
			wantSuccess: false,
		},
		{
			name:        "invalid body",
			body:        map[string]string{"email": "not-an-email"},
			mockSetup:   func(svc *mock.MockUserService) {},
			wantStatus:  fiber.StatusUnprocessableEntity,
			wantSuccess: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockSvc := mock.NewMockUserService(ctrl)
			tt.mockSetup(mockSvc)

			app := fiber.New()
			h := handler.NewAuthHandler(mockSvc)
			app.Post("/api/v1/auth/login", h.Login)

			body, _ := json.Marshal(tt.body)
			req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")

			resp, err := app.Test(req)
			assert.NoError(t, err)
			assert.Equal(t, tt.wantStatus, resp.StatusCode)
		})
	}
}

func TestAuthHandler_Refresh(t *testing.T) {
	tests := []struct {
		name        string
		body        map[string]string
		mockSetup   func(svc *mock.MockUserService)
		wantStatus  int
		wantSuccess bool
	}{
		{
			name: "success",
			body: map[string]string{"refresh_token": "validtoken"},
			mockSetup: func(svc *mock.MockUserService) {
				svc.EXPECT().
					Refresh(gomock.Any(), "validtoken").
					Return(&domain.AuthResponse{AccessToken: "new_access", RefreshToken: "validtoken"}, nil)
			},
			wantStatus:  fiber.StatusOK,
			wantSuccess: true,
		},
		{
			name: "token not found",
			body: map[string]string{"refresh_token": "badtoken"},
			mockSetup: func(svc *mock.MockUserService) {
				svc.EXPECT().
					Refresh(gomock.Any(), "badtoken").
					Return(nil, errors.New("token is invalid or expired"))
			},
			wantStatus:  fiber.StatusUnauthorized,
			wantSuccess: false,
		},
		{
			name:        "invalid body",
			body:        map[string]string{},
			mockSetup:   func(svc *mock.MockUserService) {},
			wantStatus:  fiber.StatusUnprocessableEntity,
			wantSuccess: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockSvc := mock.NewMockUserService(ctrl)
			tt.mockSetup(mockSvc)

			app := fiber.New()
			h := handler.NewAuthHandler(mockSvc)
			app.Post("/api/v1/auth/refresh", h.Refresh)

			body, _ := json.Marshal(tt.body)
			req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")

			resp, err := app.Test(req)
			assert.NoError(t, err)
			assert.Equal(t, tt.wantStatus, resp.StatusCode)
		})
	}
}

func TestAuthHandler_Logout(t *testing.T) {
	tests := []struct {
		name        string
		body        map[string]string
		mockSetup   func(svc *mock.MockUserService)
		wantStatus  int
		wantSuccess bool
	}{
		{
			name: "success",
			body: map[string]string{"refresh_token": "validtoken"},
			mockSetup: func(svc *mock.MockUserService) {
				svc.EXPECT().
					Logout(gomock.Any(), "validtoken").
					Return(nil)
			},
			wantStatus:  fiber.StatusOK,
			wantSuccess: true,
		},
		{
			name: "service error",
			body: map[string]string{"refresh_token": "validtoken"},
			mockSetup: func(svc *mock.MockUserService) {
				svc.EXPECT().
					Logout(gomock.Any(), "validtoken").
					Return(errors.New("db error"))
			},
			wantStatus:  fiber.StatusInternalServerError,
			wantSuccess: false,
		},
		{
			name:        "invalid body",
			body:        map[string]string{},
			mockSetup:   func(svc *mock.MockUserService) {},
			wantStatus:  fiber.StatusUnprocessableEntity,
			wantSuccess: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockSvc := mock.NewMockUserService(ctrl)
			tt.mockSetup(mockSvc)

			app := fiber.New()
			h := handler.NewAuthHandler(mockSvc)
			app.Post("/api/v1/auth/logout", h.Logout)

			body, _ := json.Marshal(tt.body)
			req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")

			resp, err := app.Test(req)
			assert.NoError(t, err)
			assert.Equal(t, tt.wantStatus, resp.StatusCode)
		})
	}
}
