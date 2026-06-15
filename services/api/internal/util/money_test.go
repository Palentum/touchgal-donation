package util

import "testing"

func TestParseAndFormatCents(t *testing.T) {
	cases := map[string]int64{"0": 0, "9.9": 990, "29.90": 2990, "+100": 10000, "-1.25": -125}
	for input, want := range cases {
		got, err := ParseDecimalCents(input)
		if err != nil {
			t.Fatalf("ParseDecimalCents(%q): %v", input, err)
		}
		if got != want {
			t.Fatalf("ParseDecimalCents(%q)=%d want %d", input, got, want)
		}
	}
	if FormatCents(2990) != "29.90" || FormatCents(-125) != "-1.25" {
		t.Fatalf("unexpected formatted cents")
	}
	if _, err := ParseDecimalCents("1.234"); err == nil {
		t.Fatalf("expected too many decimals to fail")
	}
}
