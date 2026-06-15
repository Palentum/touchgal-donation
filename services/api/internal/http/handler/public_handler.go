package handler

import (
	"strconv"

	"donation-site/services/api/internal/service"

	"github.com/gofiber/fiber/v3"
)

type PublicHandler struct {
	donations *service.DonationService
	routes    *service.RouteService
}

func NewPublicHandler(donations *service.DonationService, routes *service.RouteService) *PublicHandler {
	return &PublicHandler{donations: donations, routes: routes}
}

func (h *PublicHandler) Config(c fiber.Ctx) error {
	cfg, err := h.donations.PublicConfig(c)
	if err != nil {
		return err
	}
	return respondOK(c, cfg)
}

func (h *PublicHandler) RecentDonations(c fiber.Ctx) error {
	days, _ := strconv.Atoi(c.Query("days"))
	items, err := h.donations.RecentPublic(c, days)
	if err != nil {
		return err
	}
	return respondOK(c, fiber.Map{"items": items})
}

func (h *PublicHandler) CreateDonation(c fiber.Ctx) error {
	var req service.CreateDonationInput
	if err := c.Bind().Body(&req); err != nil {
		return service.Validation("请求格式无效", nil)
	}
	out, err := h.donations.CreateDonation(c, req, service.RequestMeta{IP: c.IP(), UserAgent: c.Get(fiber.HeaderUserAgent)})
	if err != nil {
		return err
	}
	return respondOK(c, out)
}

func (h *PublicHandler) DonationStatus(c fiber.Ctx) error {
	out, err := h.donations.DonationStatus(c, c.Params("orderNo"))
	if err != nil {
		return err
	}
	return respondOK(c, out)
}

func (h *PublicHandler) DonationQR(c fiber.Ctx) error {
	png, err := h.donations.QRPNG(c, c.Params("orderNo"))
	if err != nil {
		return err
	}
	c.Set(fiber.HeaderContentType, "image/png")
	return c.Send(png)
}

func (h *PublicHandler) ResolveRoute(c fiber.Ctx) error {
	path := c.Query("path")
	if path == "" {
		return service.Validation("path 不能为空", nil)
	}
	out, err := h.routes.Resolve(c, path)
	if err != nil {
		return err
	}
	return respondOK(c, out)
}
