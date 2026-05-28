package handler

import (
	"context"

	"github.com/gofiber/fiber/v2"

	"github.com/firmanains/cinema-ticketing/internal/domain"
	"github.com/firmanains/cinema-ticketing/pkg/response"
	"github.com/firmanains/cinema-ticketing/pkg/validator"
)

type UserService interface {
	Register(ctx context.Context, req domain.RegisterRequest) (*domain.AuthResponse, error)
	Login(ctx context.Context, req domain.LoginRequest) (*domain.AuthResponse, error)
	Refresh(ctx context.Context, refreshToken string) (*domain.AuthResponse, error)
	Logout(ctx context.Context, refreshToken string) error
}

type AuthHandler struct {
	userSvc UserService
}

func NewAuthHandler(userSvc UserService) *AuthHandler {
	return &AuthHandler{userSvc: userSvc}
}

func (h *AuthHandler) Register(c *fiber.Ctx) error {
	var req domain.RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusUnprocessableEntity, "invalid request body")
	}
	if err := validator.Validate(req); err != nil {
		return response.Error(c, fiber.StatusUnprocessableEntity, err.Error())
	}

	result, err := h.userSvc.Register(c.UserContext(), req)
	if err != nil {
		return response.ErrorFromDomain(c, err)
	}

	return response.Success(c, fiber.StatusCreated, "registered successfully", result)
}

func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req domain.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusUnprocessableEntity, "invalid request body")
	}
	if err := validator.Validate(req); err != nil {
		return response.Error(c, fiber.StatusUnprocessableEntity, err.Error())
	}

	result, err := h.userSvc.Login(c.UserContext(), req)
	if err != nil {
		return response.ErrorFromDomain(c, err)
	}

	return response.Success(c, fiber.StatusOK, "login successful", result)
}

func (h *AuthHandler) Refresh(c *fiber.Ctx) error {
	var body struct {
		RefreshToken string `json:"refresh_token" validate:"required"`
	}
	if err := c.BodyParser(&body); err != nil {
		return response.Error(c, fiber.StatusUnprocessableEntity, "invalid request body")
	}
	if err := validator.Validate(body); err != nil {
		return response.Error(c, fiber.StatusUnprocessableEntity, err.Error())
	}

	result, err := h.userSvc.Refresh(c.UserContext(), body.RefreshToken)
	if err != nil {
		return response.ErrorFromDomain(c, err)
	}

	return response.Success(c, fiber.StatusOK, "token refreshed", result)
}

func (h *AuthHandler) Logout(c *fiber.Ctx) error {
	var body struct {
		RefreshToken string `json:"refresh_token" validate:"required"`
	}
	if err := c.BodyParser(&body); err != nil {
		return response.Error(c, fiber.StatusUnprocessableEntity, "invalid request body")
	}
	if err := validator.Validate(body); err != nil {
		return response.Error(c, fiber.StatusUnprocessableEntity, err.Error())
	}

	if err := h.userSvc.Logout(c.UserContext(), body.RefreshToken); err != nil {
		return response.ErrorFromDomain(c, err)
	}

	return response.Success(c, fiber.StatusOK, "logged out successfully", nil)
}
