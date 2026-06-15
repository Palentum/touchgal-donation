package middleware

import (
	"time"

	"donation-site/services/api/internal/service"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/limiter"
)

func RateLimit(max int, expiration time.Duration) fiber.Handler {
	return limiter.New(limiter.Config{
		Max:        max,
		Expiration: expiration,
		LimitReached: func(c fiber.Ctx) error {
			return &service.Error{Status: fiber.StatusTooManyRequests, Code: "RATE_LIMITED", Message: "请求过于频繁，请稍后再试"}
		},
	})
}
