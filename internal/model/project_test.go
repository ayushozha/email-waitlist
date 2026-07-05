package model

import (
	"strings"
	"testing"
)

func TestGenerateKeyPrefixesAndUniqueness(t *testing.T) {
	secret, err := generateKey(secretKeyPrefix, 32)
	if err != nil {
		t.Fatal(err)
	}
	public, err := generateKey(PublicKeyPrefix, 16)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(secret, "wl_sec_") || len(secret) != len("wl_sec_")+64 {
		t.Errorf("secret key malformed: %q", secret)
	}
	if !strings.HasPrefix(public, "wl_pub_") || len(public) != len("wl_pub_")+32 {
		t.Errorf("public key malformed: %q", public)
	}

	other, _ := generateKey(secretKeyPrefix, 32)
	if secret == other {
		t.Error("two generated keys are identical")
	}
}

func TestHashAPIKeyIsStableHex(t *testing.T) {
	h1 := HashAPIKey("wl_sec_abc")
	h2 := HashAPIKey("wl_sec_abc")
	if h1 != h2 {
		t.Error("hash not deterministic")
	}
	if len(h1) != 64 {
		t.Errorf("hash length = %d, want 64 hex chars", len(h1))
	}
	if h1 == HashAPIKey("wl_sec_abd") {
		t.Error("different keys produced the same hash")
	}
}

func TestCreateProjectRequestValidate(t *testing.T) {
	valid := CreateProjectRequest{Name: "My App", Slug: "my-app"}
	if err := valid.Validate(); err != nil {
		t.Errorf("valid request rejected: %v", err)
	}

	for _, bad := range []CreateProjectRequest{
		{Name: "", Slug: "my-app"},
		{Name: "  ", Slug: "my-app"},
		{Name: "X", Slug: "My-App"},
		{Name: "X", Slug: "my_app"},
		{Name: "X", Slug: "-leading"},
		{Name: "X", Slug: ""},
	} {
		if err := bad.Validate(); err == nil {
			t.Errorf("Validate(%+v) = nil, want error", bad)
		}
	}
}
