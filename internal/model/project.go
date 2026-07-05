package model

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// PublicKeyPrefix marks publishable keys, which are safe to embed in browser
// code and only grant access to POST /subscribe. Secret keys ("wl_sec_...",
// or legacy "wl_...") grant management access and are stored hashed.
const (
	PublicKeyPrefix = "wl_pub_"
	secretKeyPrefix = "wl_sec_"
)

type Project struct {
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	Slug           string    `json:"slug"`
	APIKey         string    `json:"api_key,omitempty"` // secret key; only returned at creation time
	PublicKey      string    `json:"public_key"`
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

func generateKey(prefix string, bytes int) (string, error) {
	b := make([]byte, bytes)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return prefix + hex.EncodeToString(b), nil
}

// HashAPIKey returns the hex SHA-256 of a secret key. Keys are high-entropy
// random values, so an unsalted hash is sufficient for storage. Lookups
// compare hashes via an indexed SQL equality — not constant-time, which is
// accepted: the compared values are 256-bit digests of random keys, so any
// timing signal reveals nothing an attacker can extend byte-by-byte.
func HashAPIKey(key string) string {
	sum := sha256.Sum256([]byte(key))
	return hex.EncodeToString(sum[:])
}

func CreateProject(ctx context.Context, pool *pgxpool.Pool, req CreateProjectRequest) (*Project, error) {
	secretKey, err := generateKey(secretKeyPrefix, 32)
	if err != nil {
		return nil, fmt.Errorf("generate secret key: %w", err)
	}
	publicKey, err := generateKey(PublicKeyPrefix, 16)
	if err != nil {
		return nil, fmt.Errorf("generate public key: %w", err)
	}

	p := &Project{Name: req.Name, Slug: req.Slug, APIKey: secretKey, PublicKey: publicKey}
	err = pool.QueryRow(ctx,
		`INSERT INTO projects (name, slug, api_key_hash, public_key, allowed_origins)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id, allowed_origins, created_at`,
		req.Name, req.Slug, HashAPIKey(secretKey), publicKey, req.AllowedOrigins,
	).Scan(&p.ID, &p.AllowedOrigins, &p.CreatedAt)
	if err != nil {
		if isUniqueViolation(err, "projects_slug_key") {
			return nil, ErrSlugTaken
		}
		return nil, fmt.Errorf("insert project: %w", err)
	}

	return p, nil
}

// GetProjectForSubscribe resolves either key type — the publishable key or
// the secret key — for the public subscribe endpoint.
func GetProjectForSubscribe(ctx context.Context, pool *pgxpool.Pool, key string) (*Project, error) {
	return getProject(ctx, pool,
		`WHERE public_key = $1 OR api_key_hash = $2`, key, HashAPIKey(key))
}

// GetProjectBySecretKey resolves only the secret key, for management endpoints.
func GetProjectBySecretKey(ctx context.Context, pool *pgxpool.Pool, key string) (*Project, error) {
	return getProject(ctx, pool, `WHERE api_key_hash = $1`, HashAPIKey(key))
}

func getProject(ctx context.Context, pool *pgxpool.Pool, where string, args ...any) (*Project, error) {
	p := &Project{}
	err := pool.QueryRow(ctx,
		`SELECT id, name, slug, public_key, allowed_origins, created_at
		 FROM projects `+where, args...,
	).Scan(&p.ID, &p.Name, &p.Slug, &p.PublicKey, &p.AllowedOrigins, &p.CreatedAt)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func ListProjects(ctx context.Context, pool *pgxpool.Pool) ([]Project, error) {
	rows, err := pool.Query(ctx,
		`SELECT id, name, slug, public_key, allowed_origins, created_at
		 FROM projects ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projects []Project
	for rows.Next() {
		var p Project
		if err := rows.Scan(&p.ID, &p.Name, &p.Slug, &p.PublicKey, &p.AllowedOrigins, &p.CreatedAt); err != nil {
			return nil, err
		}
		projects = append(projects, p)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return projects, nil
}
