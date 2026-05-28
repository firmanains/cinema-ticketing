package handler

import (
	"github.com/gofiber/fiber/v2"

	"github.com/firmanains/cinema-ticketing/internal/domain"
	"github.com/firmanains/cinema-ticketing/pkg/response"
	"github.com/firmanains/cinema-ticketing/pkg/validator"
)

type AuthHandler struct {
	userSvc domain.UserService
}

func NewAuthHandler(userSvc domain.UserService) *AuthHandler {
	return &AuthHandler{userSvc: userSvc}
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
		return response.Error(c, fiber.StatusUnauthorized, err.Error())
	}

	return response.Success(c, fiber.StatusOK, "token refreshed", result)
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
		return response.Error(c, fiber.StatusUnauthorized, err.Error())
	}

	return response.Success(c, fiber.StatusOK, "login successful", result)
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
		if err.Error() == "email already registered" {
			return response.Error(c, fiber.StatusConflict, err.Error())
		}
		return response.Error(c, fiber.StatusInternalServerError, "internal server error")
	}

	return response.Success(c, fiber.StatusCreated, "registered successfully", result)
}
