package repository

import (
	"context"

	"donation-site/services/api/internal/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PaymentMethodRepository struct{ db *gorm.DB }

func NewPaymentMethodRepository(db *gorm.DB) *PaymentMethodRepository {
	return &PaymentMethodRepository{db: db}
}

func (r *PaymentMethodRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.PaymentMethod{}).Count(&count).Error
	return count, err
}

func (r *PaymentMethodRepository) Enabled(ctx context.Context) ([]model.PaymentMethod, error) {
	var methods []model.PaymentMethod
	err := r.db.WithContext(ctx).Where("enabled = true").Order("sort_order ASC, created_at ASC").Find(&methods).Error
	return methods, err
}

func (r *PaymentMethodRepository) Find(ctx context.Context, id uuid.UUID) (*model.PaymentMethod, error) {
	var method model.PaymentMethod
	err := r.db.WithContext(ctx).First(&method, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &method, nil
}
