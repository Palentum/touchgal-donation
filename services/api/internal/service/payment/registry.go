package payment

import (
	"donation-site/services/api/internal/model"
)

type Registry struct {
	providers map[string]Provider
}

func NewRegistry() *Registry {
	providers := map[string]Provider{
		model.PaymentTypeStaticQR:    StaticQRProvider{},
		model.PaymentTypeRedirectURL: RedirectURLProvider{},
		model.PaymentTypeMockQR:      MockQRProvider{},
	}
	unsupported := UnsupportedProvider{}
	providers[model.PaymentTypeWechatNative] = unsupported
	providers[model.PaymentTypeAlipayF2F] = unsupported
	providers[model.PaymentTypeStripeCheckout] = unsupported
	return &Registry{providers: providers}
}

func (r *Registry) Get(paymentType string) (Provider, bool) {
	provider, ok := r.providers[paymentType]
	return provider, ok
}

func SupportedPaymentTypes() map[string]bool {
	return map[string]bool{
		model.PaymentTypeStaticQR:       true,
		model.PaymentTypeRedirectURL:    true,
		model.PaymentTypeMockQR:         true,
		model.PaymentTypeWechatNative:   true,
		model.PaymentTypeAlipayF2F:      true,
		model.PaymentTypeStripeCheckout: true,
	}
}
