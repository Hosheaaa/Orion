package database

import (
	"context"
	"database/sql"
	"fmt"
)

// Migrate 执行基础表结构初始化（幂等）
func Migrate(ctx context.Context, db *sql.DB) error {
	statements := []string{
		`CREATE TABLE IF NOT EXISTS activities (
			id UUID PRIMARY KEY,
			title TEXT NOT NULL,
			description TEXT DEFAULT '',
			speaker TEXT NOT NULL,
			start_time TIMESTAMPTZ NOT NULL,
			end_time TIMESTAMPTZ,
			input_language TEXT NOT NULL,
			target_languages JSONB NOT NULL,
			cover_url TEXT,
			status TEXT NOT NULL,
			viewer_url TEXT,
			created_at TIMESTAMPTZ NOT NULL,
			updated_at TIMESTAMPTZ NOT NULL
		);`,
		`CREATE INDEX IF NOT EXISTS idx_activities_status ON activities (status);`,
		`CREATE TABLE IF NOT EXISTS activity_tokens (
			id UUID PRIMARY KEY,
			activity_id UUID NOT NULL REFERENCES activities(id) ON DELETE CASCADE,
			type TEXT NOT NULL,
			value TEXT NOT NULL,
			expires_at TIMESTAMPTZ NOT NULL,
			max_audience INT,
			created_at TIMESTAMPTZ NOT NULL,
			status TEXT NOT NULL,
			UNIQUE(activity_id, type, value)
		);`,
		`CREATE INDEX IF NOT EXISTS idx_activity_tokens_activity ON activity_tokens (activity_id);`,
		`CREATE TABLE IF NOT EXISTS viewer_entries (
			activity_id UUID PRIMARY KEY REFERENCES activities(id) ON DELETE CASCADE,
			share_url TEXT NOT NULL,
			qr_type TEXT NOT NULL,
			qr_content TEXT NOT NULL,
			status TEXT NOT NULL,
			updated_at TIMESTAMPTZ NOT NULL
		);`,
	}

	for _, stmt := range statements {
		if _, err := db.ExecContext(ctx, stmt); err != nil {
			return fmt.Errorf("failed to execute migration statement: %w", err)
		}
	}

	return nil
}
