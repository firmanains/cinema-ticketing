package response

import (
	"errors"

	"github.com/gofiber/fiber/v2"

	"github.com/firmanains/cinema-ticketing/internal/constants"
)

type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

func Success(c *fiber.Ctx, status int, message string, data interface{}) error {
	return c.Status(status).JSON(Response{
		Success: true,
		Message: message,
		Data:    data,
	})
}

func Error(c *fiber.Ctx, status int, message string) error {
	return c.Status(status).JSON(Response{
		Success: false,
		Message: message,
		Data:    nil,
	})
}

func ErrorFromDomain(c *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, constants.ErrUserNotFound):
		return Error(c, fiber.StatusNotFound, err.Error())
	case errors.Is(err, constants.ErrShowtimeNotFound):
		return Error(c, fiber.StatusNotFound, err.Error())
	case errors.Is(err, constants.ErrInvalidCredentials):
		return Error(c, fiber.StatusUnauthorized, err.Error())
	case errors.Is(err, constants.ErrTokenInvalid):
		return Error(c, fiber.StatusUnauthorized, err.Error())
	case errors.Is(err, constants.ErrDuplicateEmail):
		return Error(c, fiber.StatusConflict, err.Error())
	case errors.Is(err, constants.ErrInvalidTimeRange):
		return Error(c, fiber.StatusUnprocessableEntity, err.Error())
	case errors.Is(err, constants.ErrNoFieldsToUpdate):
		return Error(c, fiber.StatusUnprocessableEntity, err.Error())
	default:
		return Error(c, fiber.StatusInternalServerError, "internal server error")
	}
}
