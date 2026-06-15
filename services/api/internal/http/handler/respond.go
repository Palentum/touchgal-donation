package handler

import (
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/requestid"
)

type successResponse struct {
	Data      any    `json:"data"`
	RequestID string `json:"request_id"`
}

func respondOK(c fiber.Ctx, data any) error {
	return c.JSON(successResponse{Data: data, RequestID: requestid.FromContext(c)})
}
