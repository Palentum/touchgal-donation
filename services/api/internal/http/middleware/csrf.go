package middleware

import (
	"donation-site/services/api/internal/model"
	"donation-site/services/api/internal/service"

	"github.com/gofiber/fiber/v3"
)

func CSRFRequired(auth *service.AuthService) fiber.Handler {
	return func(c fiber.Ctx) error {
		switch c.Method() {
		case fiber.MethodGet, fiber.MethodHead, fiber.MethodOptions:
			return c.Next()
		}
		session, ok := c.Locals(LocalSession).(*model.AdminSession)
		if !ok || !auth.VerifyCSRF(*session, c.Get("X-CSRF-Token")) {
			return service.Forbidden("CSRF token 无效")
		}
		return c.Next()
	}
}
