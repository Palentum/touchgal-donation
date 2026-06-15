package util

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

func FormatCents(cents int64) string {
	sign := ""
	if cents < 0 {
		sign = "-"
		cents = -cents
	}
	return fmt.Sprintf("%s%d.%02d", sign, cents/100, cents%100)
}

func ParseDecimalCents(value string) (int64, error) {
	s := strings.TrimSpace(value)
	if s == "" {
		return 0, errors.New("empty amount")
	}
	if strings.HasPrefix(s, "+") {
		s = s[1:]
	}
	negative := false
	if strings.HasPrefix(s, "-") {
		negative = true
		s = s[1:]
	}
	parts := strings.Split(s, ".")
	if len(parts) > 2 || parts[0] == "" {
		return 0, errors.New("invalid amount")
	}
	for _, r := range parts[0] {
		if !unicode.IsDigit(r) {
			return 0, errors.New("invalid amount")
		}
	}
	whole, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return 0, err
	}
	fraction := int64(0)
	if len(parts) == 2 {
		if len(parts[1]) > 2 {
			return 0, errors.New("amount has more than two decimal places")
		}
		frac := parts[1]
		for _, r := range frac {
			if !unicode.IsDigit(r) {
				return 0, errors.New("invalid amount")
			}
		}
		if len(frac) == 1 {
			frac += "0"
		}
		if frac == "" {
			frac = "00"
		}
		fraction, err = strconv.ParseInt(frac, 10, 64)
		if err != nil {
			return 0, err
		}
	}
	cents := whole*100 + fraction
	if negative {
		cents = -cents
	}
	return cents, nil
}

func ValidCurrency(currency string) bool {
	if len(currency) != 3 {
		return false
	}
	for _, r := range currency {
		if r < 'A' || r > 'Z' {
			return false
		}
	}
	return true
}
