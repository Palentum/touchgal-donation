package service

import (
	"strings"
	"testing"
	"time"
)

func TestBuildDonationCSVProtectsFormulaCells(t *testing.T) {
	csvBytes, err := BuildDonationCSV([]DonationExportRow{{OrderNo: "D20260615021000ABCDEF", Nickname: "=cmd", AmountCents: 2990, Currency: "CNY", Status: "paid", Time: time.Date(2026, 6, 15, 2, 10, 0, 0, time.UTC)}})
	if err != nil {
		t.Fatal(err)
	}
	content := string(csvBytes)
	if !strings.Contains(content, "'=") {
		t.Fatalf("formula nickname was not protected: %q", content)
	}
	if !strings.Contains(content, "订单号,昵称,金额,币种,状态,时间") {
		t.Fatalf("header missing or reordered: %q", content)
	}
}
