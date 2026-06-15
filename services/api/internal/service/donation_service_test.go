package service

import (
	"testing"
	"time"

	"donation-site/services/api/internal/model"
)

func TestDonationStatusTransitions(t *testing.T) {
	now := time.Date(2026, 6, 15, 2, 10, 0, 0, time.UTC)
	d := model.Donation{Status: model.DonationStatusPending}
	if err := ApplyAdminStatusTransition(&d, model.DonationStatusPaid, now); err != nil {
		t.Fatalf("pending -> paid should be allowed: %v", err)
	}
	if d.PaidAt == nil || !d.PaidAt.Equal(now) {
		t.Fatalf("paid transition should set paid_at")
	}
	refunded := model.Donation{Status: model.DonationStatusRefunded}
	if err := ApplyAdminStatusTransition(&refunded, model.DonationStatusPaid, now); err == nil {
		t.Fatalf("refunded -> paid should be blocked")
	}
	if !CanWebhookMarkPaid(model.DonationStatusCreated) || !CanWebhookMarkPaid(model.DonationStatusPending) {
		t.Fatalf("webhook should mark created and pending paid")
	}
	if CanWebhookMarkPaid(model.DonationStatusRefunded) || CanWebhookMarkPaid(model.DonationStatusPaid) {
		t.Fatalf("webhook should not overwrite terminal statuses")
	}
}
