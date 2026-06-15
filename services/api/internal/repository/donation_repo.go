package repository

import (
	"context"
	"time"

	"donation-site/services/api/internal/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type DonationRepository struct{ db *gorm.DB }

func NewDonationRepository(db *gorm.DB) *DonationRepository { return &DonationRepository{db: db} }

func (r *DonationRepository) Create(ctx context.Context, donation *model.Donation) error {
	return r.db.WithContext(ctx).Create(donation).Error
}

func (r *DonationRepository) FindByOrderNo(ctx context.Context, orderNo string) (*model.Donation, error) {
	var donation model.Donation
	err := r.db.WithContext(ctx).Preload("PaymentMethod").First(&donation, "order_no = ?", orderNo).Error
	if err != nil {
		return nil, err
	}
	return &donation, nil
}

func (r *DonationRepository) FindByClientRequestID(ctx context.Context, id string) (*model.Donation, error) {
	var donation model.Donation
	err := r.db.WithContext(ctx).Preload("PaymentMethod").First(&donation, "client_request_id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &donation, nil
}

func (r *DonationRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.Donation, error) {
	var donation model.Donation
	err := r.db.WithContext(ctx).Preload("Tier").Preload("PaymentMethod").First(&donation, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &donation, nil
}

func (r *DonationRepository) RecentPublicPaid(ctx context.Context, since time.Time, limit int) ([]model.Donation, error) {
	var donations []model.Donation
	err := r.db.WithContext(ctx).Where("status = ? AND public_visible = true AND paid_at >= ?", model.DonationStatusPaid, since).Order("paid_at DESC").Limit(limit).Find(&donations).Error
	return donations, err
}
