package service_test

import (
	"context"
	"errors"
	"testing"

	"crypto/sha256"
	"encoding/hex"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"golang.org/x/crypto/bcrypt"

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

func newDeadRedis() *redis.Client {
	mr := miniredis.NewMiniRedis()
	mr.Start()
	addr := mr.Addr()
	mr.Close()
	return redis.NewClient(&redis.Options{Addr: addr})
}

func TestUserService_Login(t *testing.T) {
	tests := []struct {
		name       string
		input      domain.LoginRequest
		mockSetup  func(userRepo *mock.MockUserRepository, tokenRepo *mock.MockRefreshTokenRepository)
		rdb        *redis.Client
		wantErr    bool
		wantErrMsg string
	}{
		{
			name:  "success",
			input: domain.LoginRequest{Email: "john@example.com", Password: "secret123"},
			mockSetup: func(userRepo *mock.MockUserRepository, tokenRepo *mock.MockRefreshTokenRepository) {
				hash, _ := bcrypt.GenerateFromPassword([]byte("secret123"), bcrypt.DefaultCost)
				userRepo.EXPECT().
					FindByEmail(gomock.Any(), "john@example.com").
					Return(&domain.User{PasswordHash: string(hash)}, nil)
				tokenRepo.EXPECT().
					Store(gomock.Any(), gomock.Any()).
					Return(nil)
			},
			wantErr: false,
		},
		{
			name:  "user not found",
			input: domain.LoginRequest{Email: "ghost@example.com", Password: "secret123"},
			mockSetup: func(userRepo *mock.MockUserRepository, tokenRepo *mock.MockRefreshTokenRepository) {
				userRepo.EXPECT().
					FindByEmail(gomock.Any(), "ghost@example.com").
					Return(nil, errors.New("user not found"))
			},
			wantErr:    true,
			wantErrMsg: "invalid credentials",
		},
		{
			name:  "wrong password",
			input: domain.LoginRequest{Email: "john@example.com", Password: "wrongpass"},
			mockSetup: func(userRepo *mock.MockUserRepository, tokenRepo *mock.MockRefreshTokenRepository) {
				hash, _ := bcrypt.GenerateFromPassword([]byte("secret123"), bcrypt.DefaultCost)
				userRepo.EXPECT().
					FindByEmail(gomock.Any(), "john@example.com").
					Return(&domain.User{PasswordHash: string(hash)}, nil)
			},
			wantErr:    true,
			wantErrMsg: "invalid credentials",
		},
		{
			name:  "store refresh token fails",
			input: domain.LoginRequest{Email: "john@example.com", Password: "secret123"},
			mockSetup: func(userRepo *mock.MockUserRepository, tokenRepo *mock.MockRefreshTokenRepository) {
				hash, _ := bcrypt.GenerateFromPassword([]byte("secret123"), bcrypt.DefaultCost)
				userRepo.EXPECT().
					FindByEmail(gomock.Any(), "john@example.com").
					Return(&domain.User{PasswordHash: string(hash)}, nil)
				tokenRepo.EXPECT().
					Store(gomock.Any(), gomock.Any()).
					Return(errors.New("db error"))
			},
			wantErr:    true,
			wantErrMsg: "db error",
		},
		{
			name:  "redis set fails after store",
			input: domain.LoginRequest{Email: "john@example.com", Password: "secret123"},
			mockSetup: func(userRepo *mock.MockUserRepository, tokenRepo *mock.MockRefreshTokenRepository) {
				hash, _ := bcrypt.GenerateFromPassword([]byte("secret123"), bcrypt.DefaultCost)
				userRepo.EXPECT().
					FindByEmail(gomock.Any(), "john@example.com").
					Return(&domain.User{PasswordHash: string(hash)}, nil)
				tokenRepo.EXPECT().
					Store(gomock.Any(), gomock.Any()).
					Return(nil)
			},
			rdb:     newDeadRedis(),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			userRepo := mock.NewMockUserRepository(ctrl)
			tokenRepo := mock.NewMockRefreshTokenRepository(ctrl)
			tt.mockSetup(userRepo, tokenRepo)

			rdb := tt.rdb
			if rdb == nil {
				rdb = newTestRedis(t)
			}

			svc := service.NewUserService(userRepo, tokenRepo, newTestConfig(), rdb)
			result, err := svc.Login(context.Background(), tt.input)

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

func TestUserService_Refresh(t *testing.T) {
	tests := []struct {
		name         string
		refreshToken string
		mockSetup    func(userRepo *mock.MockUserRepository, tokenRepo *mock.MockRefreshTokenRepository, rdb *redis.Client)
		wantErr      bool
		wantErrMsg   string
	}{
		{
			name:         "success - cache hit",
			refreshToken: "plaintexttoken",
			mockSetup: func(_ *mock.MockUserRepository, _ *mock.MockRefreshTokenRepository, rdb *redis.Client) {
				userID := uuid.New()
				sum := sha256.Sum256([]byte("plaintexttoken"))
				hash := hex.EncodeToString(sum[:])
				rdb.Set(context.Background(), "refresh_token:"+hash, userID.String(), time.Minute)
			},
			wantErr: false,
		},
		{
			name:         "success - cache miss, db hit",
			refreshToken: "plaintexttoken2",
			mockSetup: func(_ *mock.MockUserRepository, tokenRepo *mock.MockRefreshTokenRepository, _ *redis.Client) {
				userID := uuid.New()
				tokenRepo.EXPECT().
					FindByTokenHash(gomock.Any(), gomock.Any()).
					Return(&domain.RefreshToken{
						UserID:    userID,
						ExpiresAt: time.Now().Add(24 * time.Hour),
					}, nil)
			},
			wantErr: false,
		},
		{
			name:         "token not found",
			refreshToken: "badtoken",
			mockSetup: func(_ *mock.MockUserRepository, tokenRepo *mock.MockRefreshTokenRepository, _ *redis.Client) {
				tokenRepo.EXPECT().
					FindByTokenHash(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("token is invalid or expired"))
			},
			wantErr:    true,
			wantErrMsg: "token is invalid or expired",
		},
		{
			name:         "invalid uuid in cache",
			refreshToken: "plaintexttoken3",
			mockSetup: func(_ *mock.MockUserRepository, _ *mock.MockRefreshTokenRepository, rdb *redis.Client) {
				sum := sha256.Sum256([]byte("plaintexttoken3"))
				hash := hex.EncodeToString(sum[:])
				rdb.Set(context.Background(), "refresh_token:"+hash, "not-a-valid-uuid", time.Minute)
			},
			wantErr:    true,
			wantErrMsg: "token is invalid or expired",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			userRepo := mock.NewMockUserRepository(ctrl)
			tokenRepo := mock.NewMockRefreshTokenRepository(ctrl)
			rdb := newTestRedis(t)
			tt.mockSetup(userRepo, tokenRepo, rdb)

			svc := service.NewUserService(userRepo, tokenRepo, newTestConfig(), rdb)
			result, err := svc.Refresh(context.Background(), tt.refreshToken)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErrMsg)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.NotEmpty(t, result.AccessToken)
				assert.Equal(t, tt.refreshToken, result.RefreshToken)
			}
		})
	}
}

func TestUserService_Logout(t *testing.T) {
	tests := []struct {
		name         string
		refreshToken string
		mockSetup    func(tokenRepo *mock.MockRefreshTokenRepository)
		wantErr      bool
	}{
		{
			name:         "success",
			refreshToken: "plaintexttoken",
			mockSetup: func(tokenRepo *mock.MockRefreshTokenRepository) {
				tokenRepo.EXPECT().
					RevokeByTokenHash(gomock.Any(), gomock.Any()).
					Return(nil)
			},
			wantErr: false,
		},
		{
			name:         "revoke fails",
			refreshToken: "plaintexttoken",
			mockSetup: func(tokenRepo *mock.MockRefreshTokenRepository) {
				tokenRepo.EXPECT().
					RevokeByTokenHash(gomock.Any(), gomock.Any()).
					Return(errors.New("db error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			userRepo := mock.NewMockUserRepository(ctrl)
			tokenRepo := mock.NewMockRefreshTokenRepository(ctrl)
			tt.mockSetup(tokenRepo)

			svc := service.NewUserService(userRepo, tokenRepo, newTestConfig(), newTestRedis(t))
			err := svc.Logout(context.Background(), tt.refreshToken)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
