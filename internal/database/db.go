package database

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

func Connect(ctx context.Context, databaseURL string) (*pgxpool.Pool, error) {
	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("parse database url: %w", err)
	}

	config.MaxConns = 20
	config.MinConns = 2

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("create connection pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}

	return pool, nil
}

func RunMigrations(ctx context.Context, pool *pgxpool.Pool) error {
	schema := `
	CREATE EXTENSION IF NOT EXISTS "pgcrypto";

	CREATE TABLE IF NOT EXISTS projects (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		name VARCHAR(255) NOT NULL,
		slug VARCHAR(100) NOT NULL UNIQUE,
		api_key VARCHAR(128) NOT NULL UNIQUE,
		allowed_origins TEXT[] DEFAULT '{}',
		created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
	);

	CREATE TABLE IF NOT EXISTS subscribers (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
		email VARCHAR(320) NOT NULL,
		metadata JSONB DEFAULT '{}',
		subscribed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		UNIQUE(project_id, email)
	);

	CREATE INDEX IF NOT EXISTS idx_subscribers_project_id ON subscribers(project_id);
	CREATE INDEX IF NOT EXISTS idx_subscribers_subscribed_at ON subscribers(subscribed_at);
	CREATE INDEX IF NOT EXISTS idx_projects_api_key ON projects(api_key);

	CREATE TABLE IF NOT EXISTS email_templates (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		project_id UUID NOT NULL UNIQUE REFERENCES projects(id) ON DELETE CASCADE,
		subject VARCHAR(500) NOT NULL DEFAULT 'You''re on the waitlist!',
		html_body TEXT,
		from_name VARCHAR(255),
		from_email VARCHAR(320),
		reply_to VARCHAR(320),
		enabled BOOLEAN NOT NULL DEFAULT true,
		created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
	);

	-- Backfill from_email column if table was created before it existed
	ALTER TABLE email_templates ADD COLUMN IF NOT EXISTS from_email VARCHAR(320);

	-- Generic per-project signup ranking + referral attribution.
	-- Position is 0-indexed and assigned at insert via a per-project counter.
	-- Referral code is project-scoped unique; clients may pass their own slug
	-- (e.g. paervo's "<prefix>-<MMDDYYYY>-#000") or accept the column being
	-- left null and compute their own display format from position + email.
	ALTER TABLE subscribers ADD COLUMN IF NOT EXISTS position BIGINT;
	ALTER TABLE subscribers ADD COLUMN IF NOT EXISTS referral_code TEXT;
	ALTER TABLE subscribers ADD COLUMN IF NOT EXISTS referred_by_id UUID
		REFERENCES subscribers(id) ON DELETE SET NULL;
	ALTER TABLE subscribers ADD COLUMN IF NOT EXISTS referral_count INT NOT NULL DEFAULT 0;

	-- Backfill positions for any subscribers added before this column existed
	-- so existing rows have stable ordinals. Uses subscribed_at as the tie-break.
	UPDATE subscribers s
	   SET position = sub.rn - 1
	  FROM (
	    SELECT id, ROW_NUMBER() OVER (
	      PARTITION BY project_id ORDER BY subscribed_at, id
	    ) AS rn
	    FROM subscribers
	  ) sub
	 WHERE s.id = sub.id AND s.position IS NULL;

	CREATE UNIQUE INDEX IF NOT EXISTS idx_subscribers_project_position
		ON subscribers(project_id, position);
	CREATE UNIQUE INDEX IF NOT EXISTS idx_subscribers_project_referral_code
		ON subscribers(project_id, referral_code) WHERE referral_code IS NOT NULL;
	CREATE INDEX IF NOT EXISTS idx_subscribers_referred_by
		ON subscribers(referred_by_id);
	`

	_, err := pool.Exec(ctx, schema)
	if err != nil {
		return fmt.Errorf("run migrations: %w", err)
	}

	return nil
}
