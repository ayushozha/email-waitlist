package model

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Project struct {
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	Slug           string    `json:"slug"`
	APIKey         string    `json:"api_key"`
	AllowedOrigins []string  `json:"allowed_origins"`
	CreatedAt      time.Time `json:"created_at"`
}

type CreateProjectRequest struct {
	Name           string   `json:"name"`
	Slug           string   `json:"slug"`
	AllowedOrigins []string `json:"allowed_origins"`
}

var slugRegex = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

func (r *CreateProjectRequest) Validate() error {
	if strings.TrimSpace(r.Name) == "" {
		return fmt.Errorf("name is required")
	}
	if !slugRegex.MatchString(r.Slug) {
		return fmt.Errorf("slug must be lowercase alphanumeric with hyphens (e.g. 'my-app')")
	}
	return nil
}

func GenerateAPIKey() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return "wl_" + hex.EncodeToString(b), nil
}

func CreateProject(ctx context.Context, pool *pgxpool.Pool, req CreateProjectRequest) (*Project, error) {
	apiKey, err := GenerateAPIKey()
	if err != nil {
		return nil, fmt.Errorf("generate api key: %w", err)
	}

	p := &Project{}
	err = pool.QueryRow(ctx,
		`INSERT INTO projects (name, slug, api_key, allowed_origins)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id, name, slug, api_key, allowed_origins, created_at`,
		req.Name, req.Slug, apiKey, req.AllowedOrigins,
	).Scan(&p.ID, &p.Name, &p.Slug, &p.APIKey, &p.AllowedOrigins, &p.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("insert project: %w", err)
	}

	return p, nil
}

func GetProjectByAPIKey(ctx context.Context, pool *pgxpool.Pool, apiKey string) (*Project, error) {
	p := &Project{}
	err := pool.QueryRow(ctx,
		`SELECT id, name, slug, api_key, allowed_origins, created_at
		 FROM projects WHERE api_key = $1`,
		apiKey,
	).Scan(&p.ID, &p.Name, &p.Slug, &p.APIKey, &p.AllowedOrigins, &p.CreatedAt)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func ListProjects(ctx context.Context, pool *pgxpool.Pool) ([]Project, error) {
	rows, err := pool.Query(ctx,
		`SELECT id, name, slug, api_key, allowed_origins, created_at
		 FROM projects ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projects []Project
	for rows.Next() {
		var p Project
		if err := rows.Scan(&p.ID, &p.Name, &p.Slug, &p.APIKey, &p.AllowedOrigins, &p.CreatedAt); err != nil {
			return nil, err
		}
		projects = append(projects, p)
	}
	return projects, nil
}
