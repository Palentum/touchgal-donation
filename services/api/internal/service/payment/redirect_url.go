package payment

import (
	"context"
	"net/http"

	"donation-site/services/api/internal/model"
)

type RedirectURLProvider struct{}

func (RedirectURLProvider) CreatePayment(_ context.Context, in CreatePaymentInput) (PaymentAction, ProviderPersistedFields, error) {
	config := in.Method.ConfigJSON
	template, _ := config["url_template"].(string)
	if template == "" {
		return PaymentAction{}, ProviderPersistedFields{}, ErrProviderMisconfigured
	}
	instructions, _ := config["instructions"].(string)
	if instructions == "" {
		instructions = "将跳转至第三方支付页面。"
	}
	redirectURL := renderTemplate(template, in)
	return PaymentAction{Mode: ActionRedirect, RedirectURL: redirectURL, Instructions: renderTemplate(instructions, in)}, ProviderPersistedFields{ProviderPaymentURL: redirectURL, ProviderPayload: model.JSONMap{"type": model.PaymentTypeRedirectURL}}, nil
}

func (RedirectURLProvider) VerifyWebhook(context.Context, *http.Request) (WebhookEvent, error) {
	return WebhookEvent{}, ErrWebhookUnsupported
}
