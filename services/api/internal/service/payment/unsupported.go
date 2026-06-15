package payment

import (
	"context"
	"errors"
	"net/http"
)

var (
	ErrProviderMisconfigured = errors.New("payment provider is misconfigured")
	ErrWebhookUnsupported    = errors.New("webhook is unsupported for this provider")
)

type UnsupportedProvider struct{}

func (UnsupportedProvider) CreatePayment(context.Context, CreatePaymentInput) (PaymentAction, ProviderPersistedFields, error) {
	return PaymentAction{}, ProviderPersistedFields{}, errors.New("payment provider requires merchant credentials before it can create payments")
}

func (UnsupportedProvider) VerifyWebhook(context.Context, *http.Request) (WebhookEvent, error) {
	return WebhookEvent{}, ErrWebhookUnsupported
}
