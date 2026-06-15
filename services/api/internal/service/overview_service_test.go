package service

import (
	"testing"
	"time"

	"donation-site/services/api/internal/model"
)

func TestBuildOverviewCountsPaidRevenueOnly(t *testing.T) {
	start := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	dr := DateRange{Start: start, End: start.AddDate(0, 0, 2)}
	paidAt := start.Add(2 * time.Hour)
	outsideCreated := start.AddDate(0, 0, -1)
	donations := []model.Donation{
		{Status: model.DonationStatusPaid, AmountCents: 1000, CreatedAt: start.Add(time.Hour), PaidAt: &paidAt},
		{Status: model.DonationStatusPending, AmountCents: 5000, CreatedAt: start.Add(time.Hour)},
		{Status: model.DonationStatusFailed, AmountCents: 7000, CreatedAt: start.Add(time.Hour)},
		{Status: model.DonationStatusPaid, AmountCents: 3000, CreatedAt: outsideCreated, PaidAt: &paidAt},
	}
	overview := BuildOverview(donations, dr)
	if overview.TotalPaidAmountCents != 4000 || overview.PaidOrderCount != 2 {
		t.Fatalf("paid totals = %d/%d, want 4000/2", overview.TotalPaidAmountCents, overview.PaidOrderCount)
	}
	if overview.TotalOrderCount != 3 || overview.PendingOrderCount != 1 || overview.FailedOrderCount != 1 {
		t.Fatalf("created counts wrong: %#v", overview)
	}
}
