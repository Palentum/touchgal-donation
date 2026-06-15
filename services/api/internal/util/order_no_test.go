package util

import (
	"testing"
	"time"
)

func TestGenerateOrderNoFormatAndUniqueness(t *testing.T) {
	now := time.Date(2026, 6, 15, 2, 10, 0, 0, time.FixedZone("CST", 8*3600))
	seen := map[string]bool{}
	for range 1000 {
		orderNo, err := GenerateOrderNo(now)
		if err != nil {
			t.Fatal(err)
		}
		if !IsOrderNo(orderNo) {
			t.Fatalf("invalid order number %q", orderNo)
		}
		if seen[orderNo] {
			t.Fatalf("duplicate order number %q", orderNo)
		}
		seen[orderNo] = true
	}
}
