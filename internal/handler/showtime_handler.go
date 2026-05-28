package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

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

func (h *ShowtimeHandler) GetAll(c *fiber.Ctx) error {
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 10)
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	result, err := h.svc.GetAll(c.UserContext(), page, limit)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "internal server error")
	}

	return response.Success(c, fiber.StatusOK, "ok", result)
}

func (h *ShowtimeHandler) GetByID(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, fiber.StatusUnprocessableEntity, "invalid id")
	}

	result, err := h.svc.GetByID(c.UserContext(), id)
	if err != nil {
		return response.Error(c, fiber.StatusNotFound, err.Error())
	}

	return response.Success(c, fiber.StatusOK, "ok", result)
}

func (h *ShowtimeHandler) Update(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, fiber.StatusUnprocessableEntity, "invalid id")
	}

	var req domain.UpdateShowtimeRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusUnprocessableEntity, "invalid request body")
	}
	if err := validator.Validate(req); err != nil {
		return response.Error(c, fiber.StatusUnprocessableEntity, err.Error())
	}

	result, err := h.svc.Update(c.UserContext(), id, req)
	if err != nil {
		switch err.Error() {
		case "showtime not found":
			return response.Error(c, fiber.StatusNotFound, err.Error())
		case "no fields to update", "start time must be before end time":
			return response.Error(c, fiber.StatusUnprocessableEntity, err.Error())
		default:
			return response.Error(c, fiber.StatusInternalServerError, "internal server error")
		}
	}

	return response.Success(c, fiber.StatusOK, "showtime updated", result)
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
