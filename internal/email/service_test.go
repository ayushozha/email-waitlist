package email

import (
	"strings"
	"testing"
)

func TestExtractEmail(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"Waitlist <waitlist@example.com>", "waitlist@example.com"},
		{"waitlist@example.com", "waitlist@example.com"},
		{"Two <Words> <a@b.com>", "a@b.com"},
	}
	for _, tt := range tests {
		if got := extractEmail(tt.in); got != tt.want {
			t.Errorf("extractEmail(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestRenderBodyEscapesSubscriberData(t *testing.T) {
	// Quoted local parts let a subscriber smuggle HTML into their address;
	// the body renderer must escape data values.
	data := TemplateData{
		ProjectName: "My App",
		Email:       `"<script>alert(1)</script>"@example.com`,
	}
	out, err := renderBody(`<p>Hi {{.Email}}</p>`, data)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(out, "<script>") {
		t.Errorf("subscriber HTML not escaped: %s", out)
	}
}

func TestRenderBodyKeepsTemplateHTML(t *testing.T) {
	out, err := renderBody(`<h1>{{.ProjectName}}</h1>`, TemplateData{ProjectName: "My App"})
	if err != nil {
		t.Fatal(err)
	}
	if out != "<h1>My App</h1>" {
		t.Errorf("template HTML mangled: %s", out)
	}
}

func TestRenderSubject(t *testing.T) {
	out, err := renderSubject("Welcome to {{.ProjectName}}!", TemplateData{ProjectName: "My App"})
	if err != nil {
		t.Fatal(err)
	}
	if out != "Welcome to My App!" {
		t.Errorf("subject = %q", out)
	}
}

func TestDefaultTemplateRenders(t *testing.T) {
	out, err := renderBody(defaultTemplate, TemplateData{ProjectName: "My App", Email: "a@b.com"})
	if err != nil {
		t.Fatalf("default template must always render: %v", err)
	}
	if !strings.Contains(out, "My App") || !strings.Contains(out, "a@b.com") {
		t.Error("default template missing substituted values")
	}
}
