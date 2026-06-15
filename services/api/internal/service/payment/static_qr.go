package payment

import (
	"context"
	"net/http"
	"strings"

	"donation-site/services/api/internal/model"
	"donation-site/services/api/internal/util"
)

type StaticQRProvider struct{}

func (StaticQRProvider) CreatePayment(_ context.Context, in CreatePaymentInput) (PaymentAction, ProviderPersistedFields, error) {
	config := in.Method.ConfigJSON
	qrURL, _ := config["qr_image_url"].(string)
	if qrURL == "" {
		return PaymentAction{}, ProviderPersistedFields{}, ErrProviderMisconfigured
	}
	instructions, _ := config["instructions"].(string)
	if instructions == "" {
		instructions = "请扫码支付，备注订单号 {order_no}，支付后等待确认。"
	}
	instructions = renderTemplate(instructions, in)
	return PaymentAction{Mode: ActionQRImage, QRImageURL: qrURL, Instructions: instructions}, ProviderPersistedFields{ProviderPayload: model.JSONMap{"type": model.PaymentTypeStaticQR}}, nil
}

func (StaticQRProvider) VerifyWebhook(context.Context, *http.Request) (WebhookEvent, error) {
	return WebhookEvent{}, ErrWebhookUnsupported
}

func renderTemplate(template string, in CreatePaymentInput) string {
	replacer := strings.NewReplacer(
		"{order_no}", in.OrderNo,
		"{amount_decimal}", util.FormatCents(in.AmountCents),
		"{success_url}", in.SuccessURL,
		"{cancel_url}", in.CancelURL,
	)
	return replacer.Replace(template)
}
