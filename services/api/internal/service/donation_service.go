package service

import (
	"context"
	"errors"
	"fmt"
	"net/mail"
	"strings"
	"time"

	"donation-site/services/api/internal/config"
	"donation-site/services/api/internal/model"
	"donation-site/services/api/internal/repository"
	"donation-site/services/api/internal/service/payment"
	"donation-site/services/api/internal/util"

	"github.com/google/uuid"
	qrcode "github.com/skip2/go-qrcode"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type DonationService struct {
	db        *gorm.DB
	cfg       config.Config
	tiers     *repository.TierRepository
	methods   *repository.PaymentMethodRepository
	donations *repository.DonationRepository
	settings  *repository.SettingRepository
	registry  *payment.Registry
}

type CreateDonationInput struct {
	TierID          string `json:"tier_id"`
	AmountCents     int64  `json:"amount_cents"`
	Currency        string `json:"currency"`
	PaymentMethodID string `json:"payment_method_id"`
	Nickname        string `json:"nickname"`
	Email           string `json:"email"`
	Message         string `json:"message"`
	PublicVisible   *bool  `json:"public_visible"`
	ClientRequestID string `json:"client_request_id"`
}

type RequestMeta struct {
	IP        string
	UserAgent string
}

type CreateDonationOutput struct {
	OrderNo       string                `json:"order_no"`
	Status        string                `json:"status"`
	AmountCents   int64                 `json:"amount_cents"`
	Currency      string                `json:"currency"`
	PaymentAction payment.PaymentAction `json:"payment_action"`
	ThanksURL     string                `json:"thanks_url"`
}

type PublicConfig struct {
	Site           model.JSONMap         `json:"site"`
	Tiers          []PublicTier          `json:"tiers"`
	PaymentMethods []PublicPaymentMethod `json:"payment_methods"`
}

type PublicTier struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	AmountCents int64     `json:"amount_cents"`
	Currency    string    `json:"currency"`
	Description string    `json:"description"`
	IsDefault   bool      `json:"is_default"`
}

type PublicPaymentMethod struct {
	ID      uuid.UUID `json:"id"`
	Code    string    `json:"code"`
	Name    string    `json:"name"`
	Type    string    `json:"type"`
	IconURL string    `json:"icon_url,omitempty"`
}

type RecentDonationItem struct {
	Nickname    string     `json:"nickname"`
	Message     string     `json:"message"`
	AmountCents int64      `json:"amount_cents"`
	Currency    string     `json:"currency"`
	PaidAt      *time.Time `json:"paid_at"`
}

type DonationStatusOutput struct {
	OrderNo     string     `json:"order_no"`
	Status      string     `json:"status"`
	AmountCents int64      `json:"amount_cents"`
	Currency    string     `json:"currency"`
	PaidAt      *time.Time `json:"paid_at,omitempty"`
}

func NewDonationService(db *gorm.DB, cfg config.Config) *DonationService {
	return &DonationService{
		db:        db,
		cfg:       cfg,
		tiers:     repository.NewTierRepository(db),
		methods:   repository.NewPaymentMethodRepository(db),
		donations: repository.NewDonationRepository(db),
		settings:  repository.NewSettingRepository(db),
		registry:  payment.NewRegistry(),
	}
}

func (s *DonationService) PublicConfig(ctx context.Context) (PublicConfig, error) {
	site, err := s.settings.Get(ctx, "site")
	if err != nil {
		return PublicConfig{}, err
	}
	tiers, err := s.tiers.Enabled(ctx)
	if err != nil {
		return PublicConfig{}, err
	}
	methods, err := s.methods.Enabled(ctx)
	if err != nil {
		return PublicConfig{}, err
	}
	out := PublicConfig{Site: site, Tiers: make([]PublicTier, 0, len(tiers)), PaymentMethods: make([]PublicPaymentMethod, 0, len(methods))}
	for _, tier := range tiers {
		out.Tiers = append(out.Tiers, PublicTier{ID: tier.ID, Name: tier.Name, AmountCents: tier.AmountCents, Currency: tier.Currency, Description: tier.Description, IsDefault: tier.IsDefault})
	}
	for _, method := range methods {
		out.PaymentMethods = append(out.PaymentMethods, PublicPaymentMethod{ID: method.ID, Code: method.Code, Name: method.Name, Type: method.Type, IconURL: method.IconURL})
	}
	return out, nil
}

func (s *DonationService) RecentPublic(ctx context.Context, days int) ([]RecentDonationItem, error) {
	if days <= 0 {
		days = s.cfg.MaxPublicDonationsDays
	}
	if days > s.cfg.MaxPublicDonationsDays {
		days = s.cfg.MaxPublicDonationsDays
	}
	donations, err := s.donations.RecentPublicPaid(ctx, time.Now().UTC().AddDate(0, 0, -days), 200)
	if err != nil {
		return nil, err
	}
	items := make([]RecentDonationItem, 0, len(donations))
	for _, donation := range donations {
		items = append(items, RecentDonationItem{Nickname: util.PublicNickname(donation.Nickname), Message: util.CleanPublicText(donation.Message, s.cfg.MaxDonationMessageLen), AmountCents: donation.AmountCents, Currency: donation.Currency, PaidAt: donation.PaidAt})
	}
	return items, nil
}

func (s *DonationService) CreateDonation(ctx context.Context, in CreateDonationInput, meta RequestMeta) (*CreateDonationOutput, error) {
	clientRequestID := strings.TrimSpace(in.ClientRequestID)
	if clientRequestID != "" {
		if _, err := uuid.Parse(clientRequestID); err != nil {
			return nil, Validation("client_request_id 必须是 uuid", nil)
		}
		if existing, err := s.donations.FindByClientRequestID(ctx, clientRequestID); err == nil {
			out := s.outputFromDonation(existing)
			return &out, nil
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
	}
	paymentMethodID, err := uuid.Parse(strings.TrimSpace(in.PaymentMethodID))
	if err != nil {
		return nil, Validation("支付方式无效", nil)
	}
	method, err := s.methods.Find(ctx, paymentMethodID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, Validation("支付方式无效", nil)
		}
		return nil, err
	}
	if !method.Enabled {
		return nil, Validation("支付方式不可用", nil)
	}
	amount := in.AmountCents
	currency := strings.ToUpper(strings.TrimSpace(in.Currency))
	if currency == "" {
		currency = "CNY"
	}
	var tierID *uuid.UUID
	if strings.TrimSpace(in.TierID) != "" {
		parsedTierID, err := uuid.Parse(strings.TrimSpace(in.TierID))
		if err != nil {
			return nil, Validation("捐赠档位无效", nil)
		}
		tier, err := s.tiers.Find(ctx, parsedTierID)
		if err != nil || !tier.Enabled {
			return nil, Validation("捐赠档位无效", nil)
		}
		amount = tier.AmountCents
		currency = tier.Currency
		tierID = &tier.ID
	}
	if amount <= 0 {
		return nil, Validation("金额无效", nil)
	}
	if !util.ValidCurrency(currency) {
		return nil, Validation("币种无效", nil)
	}
	if method.MinAmountCents != nil && amount < *method.MinAmountCents {
		return nil, Validation("金额低于该支付方式下限", nil)
	}
	if method.MaxAmountCents != nil && amount > *method.MaxAmountCents {
		return nil, Validation("金额高于该支付方式上限", nil)
	}
	nickname := strings.TrimSpace(in.Nickname)
	message := strings.TrimSpace(in.Message)
	email := strings.TrimSpace(in.Email)
	if len([]rune(nickname)) > 60 {
		return nil, Validation("昵称过长", nil)
	}
	if len([]rune(message)) > s.cfg.MaxDonationMessageLen {
		return nil, Validation("留言过长", nil)
	}
	if email != "" {
		addr, err := mail.ParseAddress(email)
		if err != nil || addr.Address != email {
			return nil, Validation("邮箱格式无效", nil)
		}
	}
	provider, ok := s.registry.Get(method.Type)
	if !ok {
		return nil, Validation("支付类型无效", nil)
	}
	orderNo, err := util.GenerateOrderNo(time.Now().UTC())
	if err != nil {
		return nil, err
	}
	visible := true
	if in.PublicVisible != nil {
		visible = *in.PublicVisible
	}
	donation := model.Donation{
		ID:              uuid.New(),
		OrderNo:         orderNo,
		ClientRequestID: clientRequestID,
		TierID:          tierID,
		PaymentMethodID: &method.ID,
		Nickname:        nickname,
		Email:           email,
		Message:         message,
		AmountCents:     amount,
		Currency:        currency,
		Status:          model.DonationStatusPending,
		PublicVisible:   visible,
		Provider:        method.Provider,
		ProviderPayload: model.JSONMap{},
		ClientIPHash:    util.HashFingerprint(s.cfg.SessionSecret, meta.IP),
		UserAgentHash:   util.HashFingerprint(s.cfg.SessionSecret, meta.UserAgent),
	}
	var action payment.PaymentAction
	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&donation).Error; err != nil {
			return err
		}
		createdAction, persisted, err := provider.CreatePayment(ctx, payment.CreatePaymentInput{
			DonationID:  donation.ID,
			OrderNo:     donation.OrderNo,
			AmountCents: donation.AmountCents,
			Currency:    donation.Currency,
			Nickname:    donation.Nickname,
			Email:       donation.Email,
			Message:     donation.Message,
			SuccessURL:  s.thanksURL(donation.OrderNo),
			CancelURL:   s.cfg.AppPublicURL,
			Method:      *method,
			APIBasePath: "/api/v1",
		})
		if err != nil {
			return ProviderUnavailable(err.Error())
		}
		action = createdAction
		updates := map[string]any{
			"provider_payment_url": persisted.ProviderPaymentURL,
			"provider_qr_content":  persisted.ProviderQRContent,
			"provider_payload":     persisted.ProviderPayload,
			"expired_at":           persisted.ExpiredAt,
		}
		return tx.Model(&model.Donation{}).Where("id = ?", donation.ID).Updates(updates).Error
	})
	if err != nil {
		return nil, err
	}
	return &CreateDonationOutput{OrderNo: donation.OrderNo, Status: donation.Status, AmountCents: donation.AmountCents, Currency: donation.Currency, PaymentAction: action, ThanksURL: s.thanksURL(donation.OrderNo)}, nil
}

func (s *DonationService) DonationStatus(ctx context.Context, orderNo string) (DonationStatusOutput, error) {
	donation, err := s.donations.FindByOrderNo(ctx, orderNo)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return DonationStatusOutput{}, NotFound("订单不存在")
		}
		return DonationStatusOutput{}, err
	}
	return DonationStatusOutput{OrderNo: donation.OrderNo, Status: donation.Status, AmountCents: donation.AmountCents, Currency: donation.Currency, PaidAt: donation.PaidAt}, nil
}

func (s *DonationService) QRPNG(ctx context.Context, orderNo string) ([]byte, error) {
	donation, err := s.donations.FindByOrderNo(ctx, orderNo)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, NotFound("订单不存在")
		}
		return nil, err
	}
	if donation.ProviderQRContent == "" {
		return nil, NotFound("订单没有动态二维码")
	}
	return qrcode.Encode(donation.ProviderQRContent, qrcode.Medium, 256)
}

func (s *DonationService) MockWebhookPaid(ctx context.Context, orderNo, eventID string, raw model.JSONMap) (*model.Donation, error) {
	if s.cfg.IsProduction() {
		return nil, Forbidden("生产环境不允许 mock webhook")
	}
	if eventID == "" {
		eventID = "mock:" + orderNo + ":paid"
	}
	var donation model.Donation
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&donation, "order_no = ?", orderNo).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return NotFound("订单不存在")
			}
			return err
		}
		event := model.PaymentEvent{DonationID: &donation.ID, Provider: model.PaymentTypeMockQR, EventType: "payment.paid", EventID: eventID, RawPayload: raw, Verified: true, ReceivedAt: time.Now().UTC()}
		res := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&event)
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return nil
		}
		if CanWebhookMarkPaid(donation.Status) {
			now := time.Now().UTC()
			donation.Status = model.DonationStatusPaid
			donation.PaidAt = &now
			donation.ProviderTradeNo = eventID
			return tx.Model(&model.Donation{}).Where("id = ?", donation.ID).Updates(map[string]any{"status": donation.Status, "paid_at": donation.PaidAt, "provider_trade_no": eventID, "updated_at": now}).Error
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &donation, nil
}

func (s *DonationService) outputFromDonation(d *model.Donation) CreateDonationOutput {
	return CreateDonationOutput{OrderNo: d.OrderNo, Status: d.Status, AmountCents: d.AmountCents, Currency: d.Currency, PaymentAction: s.paymentActionFromDonation(d), ThanksURL: s.thanksURL(d.OrderNo)}
}

func (s *DonationService) paymentActionFromDonation(d *model.Donation) payment.PaymentAction {
	if d.ProviderPaymentURL != "" {
		return payment.PaymentAction{Mode: payment.ActionRedirect, RedirectURL: d.ProviderPaymentURL, ExpiresAt: d.ExpiredAt}
	}
	if d.ProviderQRContent != "" {
		return payment.PaymentAction{Mode: payment.ActionQRContent, QRImageURL: fmt.Sprintf("/api/v1/public/donations/%s/qr.png", d.OrderNo), QRContent: d.ProviderQRContent, ExpiresAt: d.ExpiredAt}
	}
	if d.PaymentMethod != nil && d.PaymentMethod.Type == model.PaymentTypeStaticQR {
		qrURL, _ := d.PaymentMethod.ConfigJSON["qr_image_url"].(string)
		instructions, _ := d.PaymentMethod.ConfigJSON["instructions"].(string)
		return payment.PaymentAction{Mode: payment.ActionQRImage, QRImageURL: qrURL, Instructions: instructions}
	}
	return payment.PaymentAction{}
}

func (s *DonationService) thanksURL(orderNo string) string {
	return s.cfg.AppPublicURL + "/thanks?order_no=" + orderNo
}

func CanWebhookMarkPaid(status string) bool {
	return status == model.DonationStatusCreated || status == model.DonationStatusPending
}

func CanAdminTransitionStatus(from, to string) bool {
	if from != model.DonationStatusPending {
		return false
	}
	switch to {
	case model.DonationStatusPaid, model.DonationStatusFailed, model.DonationStatusCancelled:
		return true
	default:
		return false
	}
}

func ApplyAdminStatusTransition(d *model.Donation, to string, now time.Time) error {
	if !CanAdminTransitionStatus(d.Status, to) {
		return Validation("不允许的状态变更", map[string]string{"from": d.Status, "to": to})
	}
	d.Status = to
	if to == model.DonationStatusPaid && d.PaidAt == nil {
		d.PaidAt = &now
	}
	return nil
}

func IsPaidRevenueStatus(status string) bool {
	return status == model.DonationStatusPaid
}
