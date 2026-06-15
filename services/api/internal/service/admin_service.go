package service

import (
	"context"
	"time"

	"donation-site/services/api/internal/config"
	"donation-site/services/api/internal/model"
	"donation-site/services/api/internal/repository"
	"donation-site/services/api/internal/util"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AdminService struct {
	db       *gorm.DB
	cfg      config.Config
	admins   *repository.AdminRepository
	settings *repository.SettingRepository
	tiers    *repository.TierRepository
	methods  *repository.PaymentMethodRepository
}

func NewAdminService(db *gorm.DB, cfg config.Config) *AdminService {
	return &AdminService{
		db:       db,
		cfg:      cfg,
		admins:   repository.NewAdminRepository(db),
		settings: repository.NewSettingRepository(db),
		tiers:    repository.NewTierRepository(db),
		methods:  repository.NewPaymentMethodRepository(db),
	}
}

func (s *AdminService) Seed(ctx context.Context) error {
	adminCount, err := s.admins.Count(ctx)
	if err != nil {
		return err
	}
	if adminCount == 0 {
		hash, err := util.HashPassword(s.cfg.InitialAdminPassword)
		if err != nil {
			return err
		}
		admin := &model.Admin{Username: s.cfg.InitialAdminUsername, PasswordHash: hash, Role: "owner", Status: model.AdminStatusActive, MustChangePassword: true}
		if err := s.admins.Create(ctx, admin); err != nil {
			return err
		}
	}
	if err := s.settings.CreateIfMissing(ctx, "site", model.JSONMap{
		"name":          "Support Us",
		"hero_title":    "支持我们的持续创作",
		"hero_subtitle": "每一笔捐赠都会帮助项目继续维护、迭代和服务社区。",
		"currency":      "CNY",
		"timezone":      s.cfg.SiteTimezone,
		"goal_cents":    float64(0),
		"show_goal":     false,
	}); err != nil {
		return err
	}
	if err := s.settings.CreateIfMissing(ctx, "admin", model.JSONMap{"base_path": s.cfg.InitialAdminBasePath}); err != nil {
		return err
	}
	tierCount, err := s.tiers.Count(ctx)
	if err != nil {
		return err
	}
	if tierCount == 0 {
		tiers := []model.DonationTier{
			{Name: "一杯咖啡", AmountCents: 990, Currency: "CNY", Description: "请我们喝一杯咖啡", SortOrder: 10, Enabled: true},
			{Name: "支持维护", AmountCents: 2990, Currency: "CNY", Description: "帮助我们覆盖服务器与维护成本", SortOrder: 20, Enabled: true, IsDefault: true},
			{Name: "赞助升级", AmountCents: 9900, Currency: "CNY", Description: "支持更长期的功能升级", SortOrder: 30, Enabled: true},
		}
		if err := s.db.WithContext(ctx).Create(&tiers).Error; err != nil {
			return err
		}
	}
	methodCount, err := s.methods.Count(ctx)
	if err != nil {
		return err
	}
	if methodCount == 0 {
		method := model.PaymentMethod{
			Code:       "mock_qr",
			Name:       "模拟二维码支付",
			Type:       model.PaymentTypeMockQR,
			Provider:   model.PaymentTypeMockQR,
			Enabled:    !s.cfg.IsProduction(),
			SortOrder:  10,
			ConfigJSON: model.JSONMap{"instructions": "开发测试支付；点击模拟回调后订单会变为已支付。"},
		}
		if err := s.db.WithContext(ctx).Create(&method).Error; err != nil {
			return err
		}
	}
	return nil
}

func (s *AdminService) Audit(ctx context.Context, adminID uuid.UUID, action, targetType, targetID string, metadata model.JSONMap, ip string) error {
	if metadata == nil {
		metadata = model.JSONMap{}
	}
	log := model.AuditLog{AdminID: &adminID, Action: action, TargetType: targetType, TargetID: targetID, Metadata: metadata, IPHash: util.HashFingerprint(s.cfg.SessionSecret, ip), CreatedAt: time.Now().UTC()}
	return s.db.WithContext(ctx).Create(&log).Error
}

func (s *AdminService) Settings() *repository.SettingRepository { return s.settings }
