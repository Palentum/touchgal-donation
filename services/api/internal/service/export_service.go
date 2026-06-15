package service

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"time"

	"donation-site/services/api/internal/model"
	"donation-site/services/api/internal/util"

	"gorm.io/gorm"
)

type ExportService struct{ db *gorm.DB }

func NewExportService(db *gorm.DB) *ExportService { return &ExportService{db: db} }

type DonationExportFilter struct {
	Start     *time.Time
	End       *time.Time
	Status    string
	TimeField string
}

type DonationExportRow struct {
	OrderNo     string
	Nickname    string
	AmountCents int64
	Currency    string
	Status      string
	Time        time.Time
}

func BuildDonationCSV(rows []DonationExportRow) ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	buf.WriteString("\xEF\xBB\xBF")
	writer := csv.NewWriter(buf)
	if err := writer.Write([]string{"订单号", "昵称", "金额", "币种", "状态", "时间"}); err != nil {
		return nil, err
	}
	for _, row := range rows {
		record := []string{
			util.EscapeCSVFormula(row.OrderNo),
			util.EscapeCSVFormula(row.Nickname),
			util.EscapeCSVFormula(util.FormatCents(row.AmountCents)),
			util.EscapeCSVFormula(row.Currency),
			util.EscapeCSVFormula(row.Status),
			util.EscapeCSVFormula(row.Time.UTC().Format(time.RFC3339)),
		}
		if err := writer.Write(record); err != nil {
			return nil, err
		}
	}
	writer.Flush()
	return buf.Bytes(), writer.Error()
}

func (s *ExportService) CSV(ctx context.Context, filter DonationExportFilter) ([]byte, error) {
	var donations []model.Donation
	query := s.db.WithContext(ctx).Model(&model.Donation{})
	timeField := "created_at"
	if filter.TimeField == "paid_at" {
		timeField = "paid_at"
	}
	if filter.Start != nil {
		query = query.Where(timeField+" >= ?", *filter.Start)
	}
	if filter.End != nil {
		query = query.Where(timeField+" < ?", *filter.End)
	}
	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}
	if err := query.Order(timeField + " DESC").Find(&donations).Error; err != nil {
		return nil, err
	}
	rows := make([]DonationExportRow, 0, len(donations))
	for _, donation := range donations {
		t := donation.CreatedAt
		if donation.PaidAt != nil {
			t = *donation.PaidAt
		}
		rows = append(rows, DonationExportRow{OrderNo: donation.OrderNo, Nickname: donation.Nickname, AmountCents: donation.AmountCents, Currency: donation.Currency, Status: donation.Status, Time: t})
	}
	return BuildDonationCSV(rows)
}

func ExportFilename(start, end string) string {
	if start == "" {
		start = time.Now().UTC().Format("20060102")
	} else {
		start = compactDate(start)
	}
	if end == "" {
		end = time.Now().UTC().Format("20060102")
	} else {
		end = compactDate(end)
	}
	return fmt.Sprintf("donations_%s_%s.csv", start, end)
}

func compactDate(value string) string {
	if len(value) == len("2006-01-02") {
		return value[0:4] + value[5:7] + value[8:10]
	}
	return value
}
