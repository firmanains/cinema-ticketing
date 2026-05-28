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
