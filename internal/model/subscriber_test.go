package model

import "testing"

func TestSubscribeRequestValidate(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string // normalized email on success
		wantErr bool
	}{
		{"plain address", "user@example.com", "user@example.com", false},
		{"uppercase normalized", "  User@Example.COM ", "user@example.com", false},
		{"display name stripped", "Bob <bob@example.com>", "bob@example.com", false},
		{"angle brackets stripped", "<bob@example.com>", "bob@example.com", false},
		{"missing domain dot", "user@localhost", "", true},
		{"no at sign", "not-an-email", "", true},
		{"empty", "", "", true},
		{"too long", "a@" + string(make([]byte, 320)) + ".com", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := SubscribeRequest{Email: tt.input}
			err := req.Validate()
			if tt.wantErr {
				if err == nil {
					t.Fatalf("Validate(%q) = nil, want error", tt.input)
				}
				return
			}
			if err != nil {
				t.Fatalf("Validate(%q) = %v, want nil", tt.input, err)
			}
			if req.Email != tt.want {
				t.Errorf("normalized email = %q, want %q", req.Email, tt.want)
			}
		})
	}
}

func TestValidateBlocksUniquenessBypass(t *testing.T) {
	// All of these must normalize to the same stored value, otherwise the
	// per-project UNIQUE(project_id, email) constraint can be bypassed.
	variants := []string{
		"bob@example.com",
		"BOB@EXAMPLE.COM",
		"Bob <bob@example.com>",
		"<bob@example.com>",
		"  bob@example.com  ",
	}
	for _, v := range variants {
		req := SubscribeRequest{Email: v}
		if err := req.Validate(); err != nil {
			t.Fatalf("Validate(%q) unexpected error: %v", v, err)
		}
		if req.Email != "bob@example.com" {
			t.Errorf("Validate(%q) normalized to %q, want bob@example.com", v, req.Email)
		}
	}
}
