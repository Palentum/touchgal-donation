package handler

import (
	"donation-site/services/api/internal/model"
	"donation-site/services/api/internal/service"

	"github.com/gofiber/fiber/v3"
)

type WebhookHandler struct{ donations *service.DonationService }

func NewWebhookHandler(donations *service.DonationService) *WebhookHandler {
	return &WebhookHandler{donations: donations}
}

type mockWebhookRequest struct {
	OrderNo string        `json:"order_no"`
	EventID string        `json:"event_id"`
	Raw     model.JSONMap `json:"raw"`
}

func (h *WebhookHandler) PaymentWebhook(c fiber.Ctx) error {
	provider := c.Params("provider")
	if provider != model.PaymentTypeMockQR {
		return service.ProviderUnavailable("该支付渠道尚未配置 webhook 签名验证")
	}
	var req mockWebhookRequest
	if err := c.Bind().Body(&req); err != nil {
		return service.Validation("请求格式无效", nil)
	}
	if req.OrderNo == "" {
		return service.Validation("order_no 不能为空", nil)
	}
	if req.Raw == nil {
		req.Raw = model.JSONMap{"order_no": req.OrderNo}
	}
	donation, err := h.donations.MockWebhookPaid(c, req.OrderNo, req.EventID, req.Raw)
	if err != nil {
		return err
	}
	return respondOK(c, fiber.Map{"order_no": donation.OrderNo, "status": donation.Status})
}
