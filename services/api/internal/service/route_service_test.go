package service

import "testing"

func TestValidateAdminBasePath(t *testing.T) {
	valid := []string{"/support-console-9c2e", "/admin_tools/v2"}
	for _, path := range valid {
		if err := ValidateAdminBasePath(path); err != nil {
			t.Fatalf("%s should be valid: %v", path, err)
		}
	}
	invalid := []string{"admin", "/", "/api", "/api/private", "/thanks", "/_nuxt/x", "/bad path", "/短"}
	for _, path := range invalid {
		if err := ValidateAdminBasePath(path); err == nil {
			t.Fatalf("%s should be invalid", path)
		}
	}
}

func TestResolveAdminPath(t *testing.T) {
	res, ok := ResolveAdminPath("/secret/dashboard", "/secret")
	if !ok || res.SubPath != "/dashboard" || res.BasePath != "/secret" {
		t.Fatalf("unexpected resolution: %#v ok=%v", res, ok)
	}
	res, ok = ResolveAdminPath("/secret", "/secret")
	if !ok || res.SubPath != "/" {
		t.Fatalf("base path should resolve to root subpath")
	}
	if _, ok := ResolveAdminPath("/secretish/dashboard", "/secret"); ok {
		t.Fatalf("prefix without slash boundary must not match")
	}
}
