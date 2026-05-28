package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/firmanains/cinema-ticketing/config"
	"github.com/firmanains/cinema-ticketing/internal/domain"
	"github.com/firmanains/cinema-ticketing/internal/mock"
	"github.com/firmanains/cinema-ticketing/internal/service"
)

func newTestRedis(t *testing.T) *redis.Client {
	t.Helper()
	mr := miniredis.RunT(t)
	return redis.NewClient(&redis.Options{Addr: mr.Addr()})
}

func newTestConfig() *config.Config {
	return &config.Config{
		JWTSecret:               "testsecret",
		JWTAccessExpiresMinutes: 30,
		JWTRefreshExpiresDays:   7,
	}
}

func TestUserService_Register(t *testing.T) {
	tests := []struct {
		name       string
		input      domain.RegisterRequest
		mockSetup  func(userRepo *mock.MockUserRepository, tokenRepo *mock.MockRefreshTokenRepository)
		wantErr    bool
		wantErrMsg string
	}{
		{
			name: "success",
			input: domain.RegisterRequest{
				Name:     "John Doe",
				Email:    "john@example.com",
				Password: "secret123",
			},
			mockSetup: func(userRepo *mock.MockUserRepository, tokenRepo *mock.MockRefreshTokenRepository) {
				userRepo.EXPECT().
					Create(gomock.Any(), gomock.Any()).
					Return(nil)
				tokenRepo.EXPECT().
					Store(gomock.Any(), gomock.Any()).
					Return(nil)
			},
			wantErr: false,
		},
		{
			name: "duplicate email",
			input: domain.RegisterRequest{
				Name:     "John Doe",
				Email:    "john@example.com",
				Password: "secret123",
			},
			mockSetup: func(userRepo *mock.MockUserRepository, tokenRepo *mock.MockRefreshTokenRepository) {
				userRepo.EXPECT().
					Create(gomock.Any(), gomock.Any()).
					Return(errors.New("email already registered"))
			},
			wantErr:    true,
			wantErrMsg: "email already registered",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			userRepo := mock.NewMockUserRepository(ctrl)
			tokenRepo := mock.NewMockRefreshTokenRepository(ctrl)
			tt.mockSetup(userRepo, tokenRepo)

			svc := service.NewUserService(userRepo, tokenRepo, newTestConfig(), newTestRedis(t))
			result, err := svc.Register(context.Background(), tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErrMsg)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.NotEmpty(t, result.AccessToken)
				assert.NotEmpty(t, result.RefreshToken)
			}
		})
	}
}
