package model

import (
	"context"
	"encoding/json"
	"fmt"
	"net/mail"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Subscriber struct {
	ID           string          `json:"id"`
	ProjectID    string          `json:"project_id"`
	Email        string          `json:"email"`
	Metadata     json.RawMessage `json:"metadata"`
	SubscribedAt time.Time       `json:"subscribed_at"`
}

type SubscribeRequest struct {
	Email    string          `json:"email"`
	Metadata json.RawMessage `json:"metadata,omitempty"`
}

func (r *SubscribeRequest) Validate() error {
	if len(r.Email) > 320 {
		return fmt.Errorf("email address too long")
	}
	addr, err := mail.ParseAddress(r.Email)
	if err != nil {
		return fmt.Errorf("invalid email address")
	}
	// Ensure it has a domain with a dot (reject "user@localhost")
	parts := strings.SplitN(addr.Address, "@", 2)
	if len(parts) != 2 || !strings.Contains(parts[1], ".") {
		return fmt.Errorf("invalid email domain")
	}
	return nil
}

type SubscriberStats struct {
	Total   int            `json:"total"`
	Today   int            `json:"today"`
	Weekly  int            `json:"this_week"`
	Monthly int            `json:"this_month"`
	ByDay   []DayCount     `json:"by_day"`
}

type DayCount struct {
	Date  string `json:"date"`
	Count int    `json:"count"`
}

func AddSubscriber(ctx context.Context, pool *pgxpool.Pool, projectID string, req SubscribeRequest) (*Subscriber, error) {
	metadata := req.Metadata
	if metadata == nil {
		metadata = json.RawMessage(`{}`)
	}

	s := &Subscriber{}
	err := pool.QueryRow(ctx,
		`INSERT INTO subscribers (project_id, email, metadata)
		 VALUES ($1, $2, $3)
		 RETURNING id, project_id, email, metadata, subscribed_at`,
		projectID, req.Email, metadata,
	).Scan(&s.ID, &s.ProjectID, &s.Email, &s.Metadata, &s.SubscribedAt)
	if err != nil {
		return nil, err
	}
	return s, nil
}

func ListSubscribers(ctx context.Context, pool *pgxpool.Pool, projectID string, limit, offset int) ([]Subscriber, int, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 500 {
		limit = 500
	}

	var total int
	err := pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM subscribers WHERE project_id = $1`, projectID,
	).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	rows, err := pool.Query(ctx,
		`SELECT id, project_id, email, metadata, subscribed_at
		 FROM subscribers WHERE project_id = $1
		 ORDER BY subscribed_at DESC
		 LIMIT $2 OFFSET $3`,
		projectID, limit, offset,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var subs []Subscriber
	for rows.Next() {
		var s Subscriber
		if err := rows.Scan(&s.ID, &s.ProjectID, &s.Email, &s.Metadata, &s.SubscribedAt); err != nil {
			return nil, 0, err
		}
		subs = append(subs, s)
	}
	return subs, total, nil
}

func DeleteSubscriber(ctx context.Context, pool *pgxpool.Pool, projectID, email string) error {
	tag, err := pool.Exec(ctx,
		`DELETE FROM subscribers WHERE project_id = $1 AND email = $2`,
		projectID, email,
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("subscriber not found")
	}
	return nil
}

func GetStats(ctx context.Context, pool *pgxpool.Pool, projectID string) (*SubscriberStats, error) {
	stats := &SubscriberStats{}

	err := pool.QueryRow(ctx,
		`SELECT
			COUNT(*),
			COUNT(*) FILTER (WHERE subscribed_at >= CURRENT_DATE),
			COUNT(*) FILTER (WHERE subscribed_at >= DATE_TRUNC('week', CURRENT_DATE)),
			COUNT(*) FILTER (WHERE subscribed_at >= DATE_TRUNC('month', CURRENT_DATE))
		 FROM subscribers WHERE project_id = $1`,
		projectID,
	).Scan(&stats.Total, &stats.Today, &stats.Weekly, &stats.Monthly)
	if err != nil {
		return nil, err
	}

	rows, err := pool.Query(ctx,
		`SELECT DATE(subscribed_at) as day, COUNT(*)
		 FROM subscribers WHERE project_id = $1 AND subscribed_at >= CURRENT_DATE - INTERVAL '30 days'
		 GROUP BY day ORDER BY day`,
		projectID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var dc DayCount
		var t time.Time
		if err := rows.Scan(&t, &dc.Count); err != nil {
			return nil, err
		}
		dc.Date = t.Format("2006-01-02")
		stats.ByDay = append(stats.ByDay, dc)
	}

	return stats, nil
}

func ExportSubscribersCSV(ctx context.Context, pool *pgxpool.Pool, projectID string) ([]Subscriber, error) {
	rows, err := pool.Query(ctx,
		`SELECT id, project_id, email, metadata, subscribed_at
		 FROM subscribers WHERE project_id = $1
		 ORDER BY subscribed_at ASC`,
		projectID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var subs []Subscriber
	for rows.Next() {
		var s Subscriber
		if err := rows.Scan(&s.ID, &s.ProjectID, &s.Email, &s.Metadata, &s.SubscribedAt); err != nil {
			return nil, err
		}
		subs = append(subs, s)
	}
	return subs, nil
}
