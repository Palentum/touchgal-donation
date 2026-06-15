package service

import (
	"context"
	"sort"
	"time"

	"donation-site/services/api/internal/model"

	"gorm.io/gorm"
)

type OverviewService struct{ db *gorm.DB }

func NewOverviewService(db *gorm.DB) *OverviewService { return &OverviewService{db: db} }

type DateRange struct {
	Start time.Time
	End   time.Time
}

type Overview struct {
	Range                map[string]string `json:"range"`
	TotalPaidAmountCents int64             `json:"total_paid_amount_cents"`
	TotalOrderCount      int64             `json:"total_order_count"`
	PaidOrderCount       int64             `json:"paid_order_count"`
	PendingOrderCount    int64             `json:"pending_order_count"`
	FailedOrderCount     int64             `json:"failed_order_count"`
	Daily                []DailyOverview   `json:"daily"`
}

type DailyOverview struct {
	Date            string `json:"date"`
	PaidAmountCents int64  `json:"paid_amount_cents"`
	PaidCount       int64  `json:"paid_count"`
	OrderCount      int64  `json:"order_count"`
}

func ParseDateRange(startValue, endValue string, fallbackDays int) (DateRange, error) {
	now := time.Now().UTC()
	if fallbackDays <= 0 {
		fallbackDays = 30
	}
	start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC).AddDate(0, 0, -fallbackDays+1)
	end := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC).AddDate(0, 0, 1)
	var err error
	if startValue != "" {
		start, err = time.ParseInLocation("2006-01-02", startValue, time.UTC)
		if err != nil {
			return DateRange{}, Validation("开始日期无效", nil)
		}
	}
	if endValue != "" {
		parsedEnd, err := time.ParseInLocation("2006-01-02", endValue, time.UTC)
		if err != nil {
			return DateRange{}, Validation("结束日期无效", nil)
		}
		end = parsedEnd.AddDate(0, 0, 1)
	}
	if !start.Before(end) {
		return DateRange{}, Validation("开始日期必须早于结束日期", nil)
	}
	return DateRange{Start: start, End: end}, nil
}

func (s *OverviewService) Overview(ctx context.Context, dr DateRange) (Overview, error) {
	var donations []model.Donation
	if err := s.db.WithContext(ctx).Where("(created_at >= ? AND created_at < ?) OR (paid_at >= ? AND paid_at < ?)", dr.Start, dr.End, dr.Start, dr.End).Find(&donations).Error; err != nil {
		return Overview{}, err
	}
	return BuildOverview(donations, dr), nil
}

func BuildOverview(donations []model.Donation, dr DateRange) Overview {
	daily := map[string]*DailyOverview{}
	ensureDay := func(t time.Time) *DailyOverview {
		day := t.UTC().Format("2006-01-02")
		if daily[day] == nil {
			daily[day] = &DailyOverview{Date: day}
		}
		return daily[day]
	}
	out := Overview{Range: map[string]string{"start": dr.Start.Format("2006-01-02"), "end": dr.End.AddDate(0, 0, -1).Format("2006-01-02")}}
	for _, donation := range donations {
		createdInRange := !donation.CreatedAt.Before(dr.Start) && donation.CreatedAt.Before(dr.End)
		paidInRange := donation.PaidAt != nil && !donation.PaidAt.Before(dr.Start) && donation.PaidAt.Before(dr.End)
		if createdInRange {
			out.TotalOrderCount++
			day := ensureDay(donation.CreatedAt)
			day.OrderCount++
			switch donation.Status {
			case model.DonationStatusPending:
				out.PendingOrderCount++
			case model.DonationStatusFailed:
				out.FailedOrderCount++
			}
		}
		if paidInRange && IsPaidRevenueStatus(donation.Status) {
			out.PaidOrderCount++
			out.TotalPaidAmountCents += donation.AmountCents
			day := ensureDay(*donation.PaidAt)
			day.PaidCount++
			day.PaidAmountCents += donation.AmountCents
		}
	}
	keys := make([]string, 0, len(daily))
	for key := range daily {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		out.Daily = append(out.Daily, *daily[key])
	}
	return out
}
