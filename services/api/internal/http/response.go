package httpapi

import (
	"errors"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/requestid"
)

type codedError interface {
	StatusCode() int
	ErrorCode() string
	PublicMessage() string
	ErrorDetails() any
}

type successResponse struct {
	Data      any    `json:"data"`
	RequestID string `json:"request_id"`
}

type errorEnvelope struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details any    `json:"details,omitempty"`
}

type errorResponse struct {
	Error     errorEnvelope `json:"error"`
	RequestID string        `json:"request_id"`
}

func OK(c fiber.Ctx, data any) error {
	return c.JSON(successResponse{Data: data, RequestID: requestid.FromContext(c)})
}

func Fail(c fiber.Ctx, status int, code, message string, details any) error {
	return c.Status(status).JSON(errorResponse{Error: errorEnvelope{Code: code, Message: message, Details: details}, RequestID: requestid.FromContext(c)})
}

func HandleError(c fiber.Ctx, err error) error {
	var coded codedError
	if errors.As(err, &coded) {
		return Fail(c, coded.StatusCode(), coded.ErrorCode(), coded.PublicMessage(), coded.ErrorDetails())
	}
	var fiberErr *fiber.Error
	if errors.As(err, &fiberErr) {
		return Fail(c, fiberErr.Code, "HTTP_ERROR", fiberErr.Message, nil)
	}
	return Fail(c, fiber.StatusInternalServerError, "INTERNAL_ERROR", "服务器内部错误", nil)
}
