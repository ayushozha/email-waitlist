package database

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/ayush10/email-waitlist/internal/model"
	"github.com/jackc/pgx/v5/pgxpool"
)

// These tests need a THROWAWAY PostgreSQL database — they drop and recreate
// the public schema. Set TEST_DATABASE_URL to run them; they are skipped
// otherwise (including in CI).

func testPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TEST_DATABASE_URL not set")
	}
	pool, err := Connect(context.Background(), url)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	t.Cleanup(pool.Close)

	_, err = pool.Exec(context.Background(),
		`DROP SCHEMA public CASCADE; CREATE SCHEMA public;`)
	if err != nil {
		t.Fatalf("reset schema: %v", err)
	}
	return pool
}

func TestMigrationsFreshInstall(t *testing.T) {
	pool := testPool(t)
	ctx := context.Background()

	if err := RunMigrations(ctx, pool); err != nil {
		t.Fatalf("fresh migration: %v", err)
	}
	// Idempotency: a second run must be a clean no-op.
	if err := RunMigrations(ctx, pool); err != nil {
		t.Fatalf("second migration run: %v", err)
	}

	// The plaintext api_key column must not exist on fresh installs.
	var hasPlaintext bool
	err := pool.QueryRow(ctx,
		`SELECT EXISTS (SELECT 1 FROM information_schema.columns
		 WHERE table_schema = current_schema() AND table_name = 'projects' AND column_name = 'api_key')`,
	).Scan(&hasPlaintext)
	if err != nil {
		t.Fatal(err)
	}
	if hasPlaintext {
		t.Error("fresh install must not have a plaintext api_key column")
	}

	// End-to-end: create a project and look it up by both key types.
	p, err := model.CreateProject(ctx, pool, model.CreateProjectRequest{Name: "T", Slug: "t"})
	if err != nil {
		t.Fatalf("create project: %v", err)
	}
	if !strings.HasPrefix(p.APIKey, "wl_sec_") || !strings.HasPrefix(p.PublicKey, "wl_pub_") {
		t.Fatalf("unexpected key formats: %q / %q", p.APIKey, p.PublicKey)
	}
	if _, err := model.GetProjectBySecretKey(ctx, pool, p.APIKey); err != nil {
		t.Errorf("secret key lookup failed: %v", err)
	}
	if _, err := model.GetProjectForSubscribe(ctx, pool, p.PublicKey); err != nil {
		t.Errorf("public key subscribe lookup failed: %v", err)
	}
	if _, err := model.GetProjectBySecretKey(ctx, pool, p.PublicKey); err == nil {
		t.Error("public key must NOT resolve on management lookups")
	}
}

func TestMigrationsLegacyUpgrade(t *testing.T) {
	pool := testPool(t)
	ctx := context.Background()

	// Recreate the pre-upgrade schema shape with live data.
	_, err := pool.Exec(ctx, `
		CREATE EXTENSION IF NOT EXISTS "pgcrypto";
		CREATE TABLE projects (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			name VARCHAR(255) NOT NULL,
			slug VARCHAR(100) NOT NULL UNIQUE,
			api_key VARCHAR(128) NOT NULL UNIQUE,
			allowed_origins TEXT[] DEFAULT '{}',
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);
		CREATE TABLE subscribers (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
			email VARCHAR(320) NOT NULL,
			metadata JSONB DEFAULT '{}',
			subscribed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			UNIQUE(project_id, email)
		);
		INSERT INTO projects (name, slug, api_key) VALUES ('Legacy', 'legacy', 'wl_legacykey123');
		INSERT INTO subscribers (project_id, email)
		SELECT id, 'a@example.com' FROM projects WHERE slug = 'legacy';
		INSERT INTO subscribers (project_id, email)
		SELECT id, 'b@example.com' FROM projects WHERE slug = 'legacy';
	`)
	if err != nil {
		t.Fatalf("seed legacy schema: %v", err)
	}

	if err := RunMigrations(ctx, pool); err != nil {
		t.Fatalf("legacy upgrade migration: %v", err)
	}
	if err := RunMigrations(ctx, pool); err != nil {
		t.Fatalf("second run after upgrade: %v", err)
	}

	// The legacy key must still authenticate via its hash.
	p, err := model.GetProjectBySecretKey(ctx, pool, "wl_legacykey123")
	if err != nil {
		t.Fatalf("legacy key no longer authenticates after upgrade: %v", err)
	}
	if !strings.HasPrefix(p.PublicKey, "wl_pub_") {
		t.Errorf("legacy project did not receive a public key: %q", p.PublicKey)
	}

	// Plaintext column must be gone.
	var hasPlaintext bool
	if err := pool.QueryRow(ctx,
		`SELECT EXISTS (SELECT 1 FROM information_schema.columns
		 WHERE table_schema = current_schema() AND table_name = 'projects' AND column_name = 'api_key')`,
	).Scan(&hasPlaintext); err != nil {
		t.Fatal(err)
	}
	if hasPlaintext {
		t.Error("plaintext api_key column must be dropped by the upgrade")
	}

	// Pre-existing subscribers must have backfilled 0-indexed positions.
	var positions []int64
	rows, err := pool.Query(ctx, `SELECT position FROM subscribers ORDER BY position`)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		var p int64
		if err := rows.Scan(&p); err != nil {
			t.Fatal(err)
		}
		positions = append(positions, p)
	}
	if len(positions) != 2 || positions[0] != 0 || positions[1] != 1 {
		t.Errorf("backfilled positions = %v, want [0 1]", positions)
	}
}
