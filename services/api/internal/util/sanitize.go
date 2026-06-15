package util

import (
	"html"
	"strings"
	"unicode/utf8"
)

func LimitRunes(value string, max int) string {
	if max <= 0 || utf8.RuneCountInString(value) <= max {
		return value
	}
	var b strings.Builder
	b.Grow(len(value))
	count := 0
	for _, r := range value {
		if count >= max {
			break
		}
		b.WriteRune(r)
		count++
	}
	return b.String()
}

func CleanPublicText(value string, max int) string {
	return html.EscapeString(LimitRunes(strings.TrimSpace(value), max))
}

func PublicNickname(value string) string {
	clean := CleanPublicText(value, 60)
	if clean == "" {
		return "匿名捐赠者"
	}
	return clean
}

func EscapeCSVFormula(value string) string {
	if value == "" {
		return value
	}
	first, _ := utf8.DecodeRuneInString(value)
	switch first {
	case '=', '+', '-', '@', '\t', '\r', '\n', '＝', '＋', '－', '＠':
		return "'" + value
	default:
		return value
	}
}
