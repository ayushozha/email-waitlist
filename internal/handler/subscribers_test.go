package handler

import "testing"

func TestCSVCellNeutralizesFormulas(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"user@example.com", "user@example.com"},
		{`=HYPERLINK("http://evil")@x.com`, `'=HYPERLINK("http://evil")@x.com`},
		{"+15551234", "'+15551234"},
		{"-2+3", "'-2+3"},
		{"@cmd", "'@cmd"},
		{"\tpayload", "'\tpayload"},
		{"", ""},
		{`{"source":"landing"}`, `{"source":"landing"}`},
	}
	for _, tt := range tests {
		if got := csvCell(tt.in); got != tt.want {
			t.Errorf("csvCell(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}
