package payment

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"donation-site/services/api/internal/model"
)

type MockQRProvider struct{}

func (MockQRProvider) CreatePayment(_ context.Context, in CreatePaymentInput) (PaymentAction, ProviderPersistedFields, error) {
	config := in.Method.ConfigJSON
	instructions, _ := config["instructions"].(string)
	if instructions == "" {
		instructions = "开发测试支付；点击模拟回调后订单会变为已支付。"
	}
	expiresAt := time.Now().UTC().Add(15 * time.Minute)
	qrContent := fmt.Sprintf("mockpay://donate?order_no=%s&amount=%d&currency=%s", in.OrderNo, in.AmountCents, in.Currency)
	qrURL := fmt.Sprintf("%s/public/donations/%s/qr.png", in.APIBasePath, in.OrderNo)
	return PaymentAction{Mode: ActionQRContent, QRImageURL: qrURL, QRContent: qrContent, Instructions: renderTemplate(instructions, in), ExpiresAt: &expiresAt}, ProviderPersistedFields{ProviderQRContent: qrContent, ProviderPayload: model.JSONMap{"type": model.PaymentTypeMockQR}, ExpiredAt: &expiresAt}, nil
}

func (MockQRProvider) VerifyWebhook(context.Context, *http.Request) (WebhookEvent, error) {
	return WebhookEvent{}, ErrWebhookUnsupported
}
