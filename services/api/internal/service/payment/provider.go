package payment

import (
	"context"
	"net/http"
	"time"

	"donation-site/services/api/internal/model"

	"github.com/google/uuid"
)

const (
	ActionQRImage   = "qr_image"
	ActionQRContent = "qr_content"
	ActionRedirect  = "redirect"
)

type CreatePaymentInput struct {
	DonationID  uuid.UUID
	OrderNo     string
	AmountCents int64
	Currency    string
	Nickname    string
	Email       string
	Message     string
	SuccessURL  string
	CancelURL   string
	Method      model.PaymentMethod
	APIBasePath string
}

type PaymentAction struct {
	Mode         string     `json:"mode"`
	QRImageURL   string     `json:"qr_image_url,omitempty"`
	QRContent    string     `json:"qr_content,omitempty"`
	RedirectURL  string     `json:"redirect_url,omitempty"`
	Instructions string     `json:"instructions,omitempty"`
	ExpiresAt    *time.Time `json:"expires_at,omitempty"`
}

type ProviderPersistedFields struct {
	ProviderPaymentURL string
	ProviderQRContent  string
	ProviderPayload    model.JSONMap
	ExpiredAt          *time.Time
}

type WebhookEvent struct {
	Provider  string
	EventType string
	EventID   string
	OrderNo   string
	Paid      bool
	Raw       model.JSONMap
}

type Provider interface {
	CreatePayment(ctx context.Context, in CreatePaymentInput) (PaymentAction, ProviderPersistedFields, error)
	VerifyWebhook(ctx context.Context, r *http.Request) (WebhookEvent, error)
}
