package handler

import (
	"time"

	"donation-site/services/api/internal/config"
	apimw "donation-site/services/api/internal/http/middleware"
	"donation-site/services/api/internal/model"
	"donation-site/services/api/internal/service"

	"github.com/gofiber/fiber/v3"
)

type AdminAuthHandler struct {
	auth  *service.AuthService
	cfg   config.Config
	audit *service.AdminService
}

func NewAdminAuthHandler(auth *service.AuthService, cfg config.Config, audit *service.AdminService) *AdminAuthHandler {
	return &AdminAuthHandler{auth: auth, cfg: cfg, audit: audit}
}

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (h *AdminAuthHandler) Login(c fiber.Ctx) error {
	var req loginRequest
	if err := c.Bind().Body(&req); err != nil {
		return service.Validation("请求格式无效", nil)
	}
	result, err := h.auth.Login(c, req.Username, req.Password, c.IP(), c.Get(fiber.HeaderUserAgent))
	if err != nil {
		return err
	}
	c.Cookie(&fiber.Cookie{Name: service.AdminSessionCookie, Value: result.Token, Path: "/", Expires: result.ExpiresAt, HTTPOnly: true, Secure: h.cfg.CookieSecure(), SameSite: "Lax"})
	return respondOK(c, fiber.Map{"admin": adminDTO(result.Admin), "csrf_token": result.CSRFToken})
}

func (h *AdminAuthHandler) Logout(c fiber.Ctx) error {
	if err := h.auth.Logout(c, c.Cookies(service.AdminSessionCookie)); err != nil {
		return err
	}
	if admin, ok := c.Locals(apimw.LocalAdmin).(*model.Admin); ok && h.audit != nil {
		_ = h.audit.Audit(c, admin.ID, "auth.logout", "admin", admin.ID.String(), model.JSONMap{}, c.IP())
	}
	c.Cookie(&fiber.Cookie{Name: service.AdminSessionCookie, Value: "", Path: "/", Expires: time.Unix(0, 0), MaxAge: -1, HTTPOnly: true, Secure: h.cfg.CookieSecure(), SameSite: "Lax"})
	return respondOK(c, fiber.Map{"ok": true})
}

func (h *AdminAuthHandler) Me(c fiber.Ctx) error {
	admin, ok := c.Locals(apimw.LocalAdmin).(*model.Admin)
	if !ok {
		return service.Unauthorized("请先登录")
	}
	return respondOK(c, fiber.Map{"admin": adminDTO(*admin)})
}

type changePasswordRequest struct {
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}

func (h *AdminAuthHandler) ChangePassword(c fiber.Ctx) error {
	admin, ok := c.Locals(apimw.LocalAdmin).(*model.Admin)
	if !ok {
		return service.Unauthorized("请先登录")
	}
	var req changePasswordRequest
	if err := c.Bind().Body(&req); err != nil {
		return service.Validation("请求格式无效", nil)
	}
	if err := h.auth.ChangePassword(c, admin.ID, req.OldPassword, req.NewPassword); err != nil {
		return err
	}
	if h.audit != nil {
		_ = h.audit.Audit(c, admin.ID, "auth.change_password", "admin", admin.ID.String(), model.JSONMap{}, c.IP())
	}
	return respondOK(c, fiber.Map{"ok": true})
}

func adminDTO(admin model.Admin) fiber.Map {
	return fiber.Map{"id": admin.ID, "username": admin.Username, "role": admin.Role, "must_change_password": admin.MustChangePassword}
}
