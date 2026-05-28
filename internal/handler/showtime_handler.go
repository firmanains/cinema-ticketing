package handler

import (
	"github.com/gofiber/fiber/v2"

	"github.com/firmanains/cinema-ticketing/internal/domain"
	"github.com/firmanains/cinema-ticketing/pkg/response"
	"github.com/firmanains/cinema-ticketing/pkg/validator"
)

type ShowtimeHandler struct {
	svc domain.ShowtimeService
}

func NewShowtimeHandler(svc domain.ShowtimeService) *ShowtimeHandler {
	return &ShowtimeHandler{svc: svc}
}

func (h *ShowtimeHandler) Create(c *fiber.Ctx) error {
	var req domain.CreateShowtimeRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusUnprocessableEntity, "invalid request body")
	}
	if err := validator.Validate(req); err != nil {
		return response.Error(c, fiber.StatusUnprocessableEntity, err.Error())
	}

	result, err := h.svc.Create(c.UserContext(), req)
	if err != nil {
		if err.Error() == "start time must be before end time" {
			return response.Error(c, fiber.StatusUnprocessableEntity, err.Error())
		}
		return response.Error(c, fiber.StatusInternalServerError, "internal server error")
	}

	return response.Success(c, fiber.StatusCreated, "showtime created", result)
}
