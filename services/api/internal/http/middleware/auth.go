package middleware

import (
	"donation-site/services/api/internal/service"

	"github.com/gofiber/fiber/v3"
)

const (
	LocalAdmin   = "admin"
	LocalSession = "admin_session_model"
)

func AuthRequired(auth *service.AuthService) fiber.Handler {
	return func(c fiber.Ctx) error {
		ctx, err := auth.Authenticate(c, c.Cookies(service.AdminSessionCookie))
		if err != nil {
			return err
		}
		c.Locals(LocalAdmin, &ctx.Admin)
		c.Locals(LocalSession, &ctx.Session)
		return c.Next()
	}
}
