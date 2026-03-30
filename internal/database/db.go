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
	`

	_, err := pool.Exec(ctx, schema)
	if err != nil {
		return fmt.Errorf("run migrations: %w", err)
	}

	return nil
}
