package model

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type EmailTemplate struct {
	ID        string    `json:"id"`
	ProjectID string    `json:"project_id"`
	Subject   string    `json:"subject"`
	HTMLBody  *string   `json:"html_body"`
	FromName  *string   `json:"from_name"`
	ReplyTo   *string   `json:"reply_to"`
	Enabled   bool      `json:"enabled"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type UpsertEmailTemplateRequest struct {
	Subject  *string `json:"subject"`
	HTMLBody *string `json:"html_body"`
	FromName *string `json:"from_name"`
	ReplyTo  *string `json:"reply_to"`
	Enabled  *bool   `json:"enabled"`
}

func GetEmailTemplate(ctx context.Context, pool *pgxpool.Pool, projectID string) (*EmailTemplate, error) {
	t := &EmailTemplate{}
	err := pool.QueryRow(ctx,
		`SELECT id, project_id, subject, html_body, from_name, reply_to, enabled, created_at, updated_at
		 FROM email_templates WHERE project_id = $1`,
		projectID,
	).Scan(&t.ID, &t.ProjectID, &t.Subject, &t.HTMLBody, &t.FromName, &t.ReplyTo, &t.Enabled, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return t, nil
}

func UpsertEmailTemplate(ctx context.Context, pool *pgxpool.Pool, projectID string, req UpsertEmailTemplateRequest) (*EmailTemplate, error) {
	subject := "You're on the waitlist!"
	if req.Subject != nil {
		subject = *req.Subject
	}
	if subject == "" {
		return nil, fmt.Errorf("subject cannot be empty")
	}

	enabled := true
	if req.Enabled != nil {
		enabled = *req.Enabled
	}

	t := &EmailTemplate{}
	err := pool.QueryRow(ctx,
		`INSERT INTO email_templates (project_id, subject, html_body, from_name, reply_to, enabled)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 ON CONFLICT (project_id) DO UPDATE SET
			subject = EXCLUDED.subject,
			html_body = EXCLUDED.html_body,
			from_name = EXCLUDED.from_name,
			reply_to = EXCLUDED.reply_to,
			enabled = EXCLUDED.enabled,
			updated_at = NOW()
		 RETURNING id, project_id, subject, html_body, from_name, reply_to, enabled, created_at, updated_at`,
		projectID, subject, req.HTMLBody, req.FromName, req.ReplyTo, enabled,
	).Scan(&t.ID, &t.ProjectID, &t.Subject, &t.HTMLBody, &t.FromName, &t.ReplyTo, &t.Enabled, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("upsert email template: %w", err)
	}
	return t, nil
}

func DeleteEmailTemplate(ctx context.Context, pool *pgxpool.Pool, projectID string) error {
	tag, err := pool.Exec(ctx,
		`DELETE FROM email_templates WHERE project_id = $1`,
		projectID,
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("no email template found")
	}
	return nil
}
