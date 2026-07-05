package email

import (
	"bytes"
	"context"
	"fmt"
	htmltemplate "html/template"
	"log"
	texttemplate "text/template"
	"time"

	"github.com/ayush10/email-waitlist/internal/model"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/resend/resend-go/v3"
)

type Service struct {
	client   *resend.Client
	pool     *pgxpool.Pool
	fromAddr string
}

func NewService(apiKey string, pool *pgxpool.Pool, fromAddr string) *Service {
	return &Service{
		client:   resend.NewClient(apiKey),
		pool:     pool,
		fromAddr: fromAddr,
	}
}

type TemplateData struct {
	ProjectName string
	Email       string
}

func (s *Service) SendConfirmation(project *model.Project, sub *model.Subscriber) {
	// Bounded context: this runs in a fire-and-forget goroutine, so a hung DB
	// or Resend call must not leak the goroutine forever.
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// tmpl is nil when the project has no custom template. A lookup *error* is
	// different: falling back to defaults then could send an email the project
	// explicitly disabled, so skip sending instead.
	tmpl, err := model.GetEmailTemplate(ctx, s.pool, project.ID)
	if err != nil {
		log.Printf("email: template lookup failed, skipping send [project=%s, email=%s]: %v", project.Slug, sub.Email, err)
		return
	}

	// If a template exists but is disabled, skip
	if tmpl != nil && !tmpl.Enabled {
		return
	}

	data := TemplateData{
		ProjectName: project.Name,
		Email:       sub.Email,
	}

	subject := "You're on the waitlist!"
	htmlBody := defaultTemplate
	from := s.fromAddr

	if tmpl != nil {
		subject = tmpl.Subject
		if tmpl.HTMLBody != nil && *tmpl.HTMLBody != "" {
			htmlBody = *tmpl.HTMLBody
		}
		// Build the from address: prefer per-project from_email, fall back to default
		emailAddr := extractEmail(s.fromAddr)
		if tmpl.FromEmail != nil && *tmpl.FromEmail != "" {
			emailAddr = *tmpl.FromEmail
		}
		if tmpl.FromName != nil && *tmpl.FromName != "" {
			from = fmt.Sprintf("%s <%s>", *tmpl.FromName, emailAddr)
		} else {
			from = emailAddr
		}
	}

	renderedSubject, err := renderSubject(subject, data)
	if err != nil {
		log.Printf("email: failed to render subject [project=%s, email=%s]: %v", project.Slug, sub.Email, err)
		return
	}

	renderedBody, err := renderBody(htmlBody, data)
	if err != nil {
		log.Printf("email: failed to render body [project=%s, email=%s]: %v", project.Slug, sub.Email, err)
		return
	}

	params := &resend.SendEmailRequest{
		From:    from,
		To:      []string{sub.Email},
		Subject: renderedSubject,
		Html:    renderedBody,
	}

	if tmpl != nil && tmpl.ReplyTo != nil && *tmpl.ReplyTo != "" {
		params.ReplyTo = *tmpl.ReplyTo
	}

	_, err = s.client.Emails.SendWithContext(ctx, params)
	if err != nil {
		log.Printf("email: failed to send confirmation [project=%s, email=%s]: %v", project.Slug, sub.Email, err)
		return
	}

	log.Printf("email: confirmation sent [project=%s, email=%s]", project.Slug, sub.Email)
}

func renderSubject(text string, data TemplateData) (string, error) {
	t, err := texttemplate.New("subject").Parse(text)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// renderBody uses html/template so subscriber-controlled data ({{.Email}} —
// quoted local parts may contain HTML) is escaped; the template's own HTML
// passes through untouched.
func renderBody(text string, data TemplateData) (string, error) {
	t, err := htmltemplate.New("body").Parse(text)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// extractEmail pulls the email address from "Name <email>" format.
func extractEmail(from string) string {
	for i := len(from) - 1; i >= 0; i-- {
		if from[i] == '<' {
			return from[i+1 : len(from)-1]
		}
	}
	return from
}

var defaultTemplate = `<!DOCTYPE html>
<html>
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
</head>
<body style="margin:0;padding:0;background-color:#f4f4f5;font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,sans-serif;">
  <table role="presentation" width="100%" cellspacing="0" cellpadding="0" style="background-color:#f4f4f5;padding:40px 20px;">
    <tr>
      <td align="center">
        <table role="presentation" width="560" cellspacing="0" cellpadding="0" style="background-color:#ffffff;border-radius:12px;overflow:hidden;">
          <tr>
            <td style="background:linear-gradient(135deg,#18181b 0%,#27272a 100%);padding:32px 40px;text-align:center;">
              <h1 style="margin:0;color:#ffffff;font-size:24px;font-weight:700;">{{.ProjectName}}</h1>
            </td>
          </tr>
          <tr>
            <td style="padding:40px;">
              <h2 style="margin:0 0 16px;color:#18181b;font-size:22px;font-weight:600;">You're on the list!</h2>
              <p style="margin:0 0 16px;color:#52525b;font-size:16px;line-height:1.6;">
                Thanks for joining the <strong>{{.ProjectName}}</strong> waitlist. We've reserved your spot and will notify you as soon as we're ready to launch.
              </p>
              <p style="margin:0 0 24px;color:#52525b;font-size:16px;line-height:1.6;">
                Stay tuned — exciting things are coming.
              </p>
              <hr style="border:none;border-top:1px solid #e4e4e7;margin:24px 0;">
              <p style="margin:0;color:#a1a1aa;font-size:13px;">
                You're receiving this because <strong>{{.Email}}</strong> was added to the {{.ProjectName}} waitlist.
              </p>
            </td>
          </tr>
        </table>
      </td>
    </tr>
  </table>
</body>
</html>`
