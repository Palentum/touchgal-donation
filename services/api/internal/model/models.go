package model

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	AdminStatusActive   = "active"
	AdminStatusDisabled = "disabled"

	PaymentTypeStaticQR       = "static_qr"
	PaymentTypeRedirectURL    = "redirect_url"
	PaymentTypeMockQR         = "mock_qr"
	PaymentTypeWechatNative   = "wechat_native"
	PaymentTypeAlipayF2F      = "alipay_f2f"
	PaymentTypeStripeCheckout = "stripe_checkout"

	DonationStatusCreated   = "created"
	DonationStatusPending   = "pending"
	DonationStatusPaid      = "paid"
	DonationStatusFailed    = "failed"
	DonationStatusCancelled = "cancelled"
	DonationStatusExpired   = "expired"
	DonationStatusRefunded  = "refunded"
)

type JSONMap map[string]any

func (m JSONMap) Value() (driver.Value, error) {
	if m == nil {
		return "{}", nil
	}
	b, err := json.Marshal(map[string]any(m))
	if err != nil {
		return nil, err
	}
	return string(b), nil
}

func (m *JSONMap) Scan(value any) error {
	if value == nil {
		*m = JSONMap{}
		return nil
	}
	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return fmt.Errorf("scan JSONMap from %T", value)
	}
	if len(bytes) == 0 {
		*m = JSONMap{}
		return nil
	}
	var decoded map[string]any
	if err := json.Unmarshal(bytes, &decoded); err != nil {
		return err
	}
	*m = JSONMap(decoded)
	return nil
}

type Admin struct {
	ID                 uuid.UUID  `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	Username           string     `gorm:"size:64;not null;uniqueIndex" json:"username"`
	PasswordHash       string     `gorm:"not null" json:"-"`
	Role               string     `gorm:"size:32;not null;default:owner" json:"role"`
	Status             string     `gorm:"size:20;not null;default:active" json:"status"`
	MustChangePassword bool       `gorm:"not null;default:false" json:"must_change_password"`
	LastLoginAt        *time.Time `json:"last_login_at,omitempty"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
}

func (a *Admin) BeforeCreate(_ *gorm.DB) error {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	return nil
}

type AdminSession struct {
	ID            uuid.UUID  `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	AdminID       uuid.UUID  `gorm:"type:uuid;not null;index" json:"admin_id"`
	Admin         Admin      `json:"admin"`
	TokenHash     string     `gorm:"not null;uniqueIndex" json:"-"`
	CSRFTokenHash string     `gorm:"not null" json:"-"`
	IPHash        string     `json:"-"`
	UserAgentHash string     `json:"-"`
	ExpiresAt     time.Time  `gorm:"not null;index" json:"expires_at"`
	RevokedAt     *time.Time `json:"revoked_at,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
}

func (s *AdminSession) BeforeCreate(_ *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return nil
}

type AppSetting struct {
	Key       string    `gorm:"primaryKey" json:"key"`
	Value     JSONMap   `gorm:"type:jsonb;not null;default:'{}'" json:"value"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type DonationTier struct {
	ID          uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	Name        string    `gorm:"size:80;not null" json:"name"`
	AmountCents int64     `gorm:"not null" json:"amount_cents"`
	Currency    string    `gorm:"size:3;not null;default:CNY" json:"currency"`
	Description string    `gorm:"size:300;not null;default:''" json:"description"`
	SortOrder   int       `gorm:"not null;default:0" json:"sort_order"`
	Enabled     bool      `gorm:"not null;default:true" json:"enabled"`
	IsDefault   bool      `gorm:"not null;default:false" json:"is_default"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (t *DonationTier) BeforeCreate(_ *gorm.DB) error {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	return nil
}

type PaymentMethod struct {
	ID             uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	Code           string    `gorm:"size:50;not null;uniqueIndex" json:"code"`
	Name           string    `gorm:"size:80;not null" json:"name"`
	Type           string    `gorm:"size:30;not null" json:"type"`
	Provider       string    `gorm:"size:50;not null;default:manual" json:"provider"`
	IconURL        string    `json:"icon_url"`
	ConfigJSON     JSONMap   `gorm:"column:config_json;type:jsonb;not null;default:'{}'" json:"config_json"`
	Enabled        bool      `gorm:"not null;default:true" json:"enabled"`
	SortOrder      int       `gorm:"not null;default:0" json:"sort_order"`
	MinAmountCents *int64    `json:"min_amount_cents,omitempty"`
	MaxAmountCents *int64    `json:"max_amount_cents,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

func (p *PaymentMethod) BeforeCreate(_ *gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return nil
}

type Donation struct {
	ID                 uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	OrderNo            string         `gorm:"size:40;not null;uniqueIndex" json:"order_no"`
	ClientRequestID    string         `gorm:"size:80;uniqueIndex" json:"client_request_id,omitempty"`
	TierID             *uuid.UUID     `gorm:"type:uuid" json:"tier_id,omitempty"`
	Tier               *DonationTier  `json:"tier,omitempty"`
	PaymentMethodID    *uuid.UUID     `gorm:"type:uuid" json:"payment_method_id,omitempty"`
	PaymentMethod      *PaymentMethod `json:"payment_method,omitempty"`
	Nickname           string         `gorm:"size:60;not null;default:''" json:"nickname"`
	Email              string         `gorm:"size:255;not null;default:''" json:"email,omitempty"`
	Message            string         `gorm:"size:300;not null;default:''" json:"message"`
	AmountCents        int64          `gorm:"not null" json:"amount_cents"`
	Currency           string         `gorm:"size:3;not null;default:CNY" json:"currency"`
	Status             string         `gorm:"size:20;not null;default:created" json:"status"`
	PublicVisible      bool           `gorm:"not null;default:true" json:"public_visible"`
	Provider           string         `gorm:"size:50;not null;default:manual" json:"provider"`
	ProviderTradeNo    string         `json:"provider_trade_no,omitempty"`
	ProviderPaymentURL string         `json:"provider_payment_url,omitempty"`
	ProviderQRContent  string         `json:"provider_qr_content,omitempty"`
	ProviderPayload    JSONMap        `gorm:"type:jsonb;not null;default:'{}'" json:"provider_payload"`
	PaidAt             *time.Time     `json:"paid_at,omitempty"`
	ExpiredAt          *time.Time     `json:"expired_at,omitempty"`
	ClientIPHash       string         `json:"-"`
	UserAgentHash      string         `json:"-"`
	CreatedAt          time.Time      `json:"created_at"`
	UpdatedAt          time.Time      `json:"updated_at"`
}

func (d *Donation) BeforeCreate(_ *gorm.DB) error {
	if d.ID == uuid.Nil {
		d.ID = uuid.New()
	}
	return nil
}

type PaymentEvent struct {
	ID         uuid.UUID  `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	DonationID *uuid.UUID `gorm:"type:uuid;index" json:"donation_id,omitempty"`
	Donation   *Donation  `json:"donation,omitempty"`
	Provider   string     `gorm:"size:50;not null;uniqueIndex:idx_provider_event" json:"provider"`
	EventType  string     `gorm:"size:100;not null" json:"event_type"`
	EventID    string     `gorm:"size:200;uniqueIndex:idx_provider_event" json:"event_id"`
	RawPayload JSONMap    `gorm:"type:jsonb;not null;default:'{}'" json:"raw_payload"`
	Verified   bool       `gorm:"not null;default:false" json:"verified"`
	ReceivedAt time.Time  `gorm:"not null;default:now()" json:"received_at"`
}

func (e *PaymentEvent) BeforeCreate(_ *gorm.DB) error {
	if e.ID == uuid.Nil {
		e.ID = uuid.New()
	}
	return nil
}

type AuditLog struct {
	ID         uuid.UUID  `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	AdminID    *uuid.UUID `gorm:"type:uuid" json:"admin_id,omitempty"`
	Action     string     `gorm:"size:100;not null" json:"action"`
	TargetType string     `gorm:"size:80;not null" json:"target_type"`
	TargetID   string     `json:"target_id,omitempty"`
	Metadata   JSONMap    `gorm:"type:jsonb;not null;default:'{}'" json:"metadata"`
	IPHash     string     `json:"-"`
	CreatedAt  time.Time  `json:"created_at"`
}

func (l *AuditLog) BeforeCreate(_ *gorm.DB) error {
	if l.ID == uuid.Nil {
		l.ID = uuid.New()
	}
	return nil
}
