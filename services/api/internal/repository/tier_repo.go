package repository

import (
	"context"

	"donation-site/services/api/internal/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TierRepository struct{ db *gorm.DB }

func NewTierRepository(db *gorm.DB) *TierRepository { return &TierRepository{db: db} }

func (r *TierRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.DonationTier{}).Count(&count).Error
	return count, err
}

func (r *TierRepository) Enabled(ctx context.Context) ([]model.DonationTier, error) {
	var tiers []model.DonationTier
	err := r.db.WithContext(ctx).Where("enabled = true").Order("sort_order ASC, amount_cents ASC").Find(&tiers).Error
	return tiers, err
}

func (r *TierRepository) Find(ctx context.Context, id uuid.UUID) (*model.DonationTier, error) {
	var tier model.DonationTier
	err := r.db.WithContext(ctx).First(&tier, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &tier, nil
}
