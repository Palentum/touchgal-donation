package util

import "testing"

func TestEscapeCSVFormula(t *testing.T) {
	cases := map[string]string{
		"=SUM(A1:A2)": "'=SUM(A1:A2)",
		"+cmd":        "'+cmd",
		"-cmd":        "'-cmd",
		"@cmd":        "'@cmd",
		"\tcmd":       "'\tcmd",
		"\ncmd":       "'\ncmd",
		"＝SUM()":      "'＝SUM()",
		"safe":        "safe",
	}
	for input, want := range cases {
		if got := EscapeCSVFormula(input); got != want {
			t.Fatalf("EscapeCSVFormula(%q)=%q want %q", input, got, want)
		}
	}
}

func TestPublicNicknameEscapesAndDefaults(t *testing.T) {
	if PublicNickname("") != "匿名捐赠者" {
		t.Fatalf("empty nickname should default")
	}
	if PublicNickname("<b>A</b>") != "&lt;b&gt;A&lt;/b&gt;" {
		t.Fatalf("nickname should be escaped")
	}
}
