package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"

	"github.com/firmanains/cinema-ticketing/config"
	"github.com/firmanains/cinema-ticketing/internal/domain"
)

type userService struct {
	userRepo     domain.UserRepository
	tokenRepo    domain.RefreshTokenRepository
	cfg          *config.Config
	rdb          *redis.Client
}

func NewUserService(
	userRepo domain.UserRepository,
	tokenRepo domain.RefreshTokenRepository,
	cfg *config.Config,
	rdb *redis.Client,
) domain.UserService {
	return &userService{
		userRepo:  userRepo,
		tokenRepo: tokenRepo,
		cfg:       cfg,
		rdb:       rdb,
	}
}

func (s *userService) Register(ctx context.Context, req domain.RegisterRequest) (*domain.AuthResponse, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &domain.User{
		ID:           uuid.New(),
		Name:         req.Name,
		Email:        req.Email,
		PasswordHash: string(hash),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	accessToken, err := s.generateAccessToken(user.ID)
	if err != nil {
		return nil, err
	}

	refreshToken, err := s.generateAndStoreRefreshToken(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	return &domain.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *userService) Refresh(ctx context.Context, refreshToken string) (*domain.AuthResponse, error) {
	sum := sha256.Sum256([]byte(refreshToken))
	tokenHash := hex.EncodeToString(sum[:])

	key := fmt.Sprintf("refresh_token:%s", tokenHash)
	userIDStr, err := s.rdb.Get(ctx, key).Result()
	if err != nil {
		token, dbErr := s.tokenRepo.FindByTokenHash(ctx, tokenHash)
		if dbErr != nil {
			return nil, errors.New("token is invalid or expired")
		}
		userIDStr = token.UserID.String()
		ttl := time.Until(token.ExpiresAt)
		_ = s.rdb.Set(ctx, key, userIDStr, ttl).Err()
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, errors.New("token is invalid or expired")
	}

	accessToken, err := s.generateAccessToken(userID)
	if err != nil {
		return nil, err
	}

	return &domain.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *userService) Login(ctx context.Context, req domain.LoginRequest) (*domain.AuthResponse, error) {
	user, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, errors.New("invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, errors.New("invalid credentials")
	}

	accessToken, err := s.generateAccessToken(user.ID)
	if err != nil {
		return nil, err
	}

	refreshToken, err := s.generateAndStoreRefreshToken(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	return &domain.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *userService) generateAccessToken(userID uuid.UUID) (string, error) {
	claims := jwt.MapClaims{
		"sub": userID.String(),
		"exp": time.Now().Add(time.Duration(s.cfg.JWTAccessExpiresMinutes) * time.Minute).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.cfg.JWTSecret))
}

func (s *userService) generateAndStoreRefreshToken(ctx context.Context, userID uuid.UUID) (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	plaintext := hex.EncodeToString(b)

	sum := sha256.Sum256([]byte(plaintext))
	tokenHash := hex.EncodeToString(sum[:])

	expiresAt := time.Now().Add(time.Duration(s.cfg.JWTRefreshExpiresDays) * 24 * time.Hour)

	token := &domain.RefreshToken{
		ID:        uuid.New(),
		UserID:    userID,
		TokenHash: tokenHash,
		ExpiresAt: expiresAt,
		CreatedAt: time.Now(),
	}
	if err := s.tokenRepo.Store(ctx, token); err != nil {
		return "", err
	}

	ttl := time.Until(expiresAt)
	key := fmt.Sprintf("refresh_token:%s", tokenHash)
	if err := s.rdb.Set(ctx, key, userID.String(), ttl).Err(); err != nil {
		return "", err
	}

	return plaintext, nil
}
