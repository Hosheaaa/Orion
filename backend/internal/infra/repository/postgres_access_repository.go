package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/hoshea/orion-backend/internal/domain"
)

// PostgresAccessRepository 负责令牌与观众入口的持久化
type PostgresAccessRepository struct {
	db *sql.DB
}

// NewPostgresAccessRepository 构造函数
func NewPostgresAccessRepository(db *sql.DB) *PostgresAccessRepository {
	return &PostgresAccessRepository{db: db}
}

// CreateToken 新增令牌
func (r *PostgresAccessRepository) CreateToken(ctx context.Context, token *domain.ActivityToken) error {
	if _, err := uuid.Parse(token.ID); err != nil {
		return fmt.Errorf("invalid token id: %w", err)
	}
	query := `INSERT INTO activity_tokens (
		id, activity_id, type, value, expires_at, max_audience, created_at, status
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8);`

	_, err := r.db.ExecContext(
		ctx,
		query,
		token.ID,
		token.ActivityID,
		string(token.Type),
		token.Value,
		token.ExpiresAt,
		token.MaxAudience,
		token.CreatedAt,
		string(token.Status),
	)
	if err != nil {
		return fmt.Errorf("failed to insert token: %w", err)
	}
	return nil
}

// ListTokens 列出活动所有令牌
func (r *PostgresAccessRepository) ListTokens(ctx context.Context, activityID string) ([]*domain.ActivityToken, error) {
	query := `SELECT id, activity_id, type, value, expires_at, max_audience, created_at, status
		FROM activity_tokens
		WHERE activity_id = $1
		ORDER BY created_at DESC;`

	rows, err := r.db.QueryContext(ctx, query, activityID)
	if err != nil {
		return nil, fmt.Errorf("failed to query tokens: %w", err)
	}
	defer rows.Close()

	var tokens []*domain.ActivityToken
	for rows.Next() {
		token, err := scanToken(rows)
		if err != nil {
			return nil, err
		}
		tokens = append(tokens, token)
	}
	return tokens, rows.Err()
}

// FindToken 根据值查找令牌
func (r *PostgresAccessRepository) FindTokenByID(ctx context.Context, id string) (*domain.ActivityToken, error) {
	row := r.db.QueryRowContext(ctx, `SELECT id, activity_id, type, value, expires_at, max_audience, created_at, status FROM activity_tokens WHERE id = $1;`, id)
	token, err := scanToken(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return token, nil
}

func (r *PostgresAccessRepository) FindToken(ctx context.Context, activityID string, tokenType domain.TokenType, value string) (*domain.ActivityToken, error) {
	query := `SELECT id, activity_id, type, value, expires_at, max_audience, created_at, status
		FROM activity_tokens
		WHERE activity_id = $1 AND type = $2 AND value = $3
		LIMIT 1;`

	row := r.db.QueryRowContext(ctx, query, activityID, string(tokenType), value)
	token, err := scanToken(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return token, nil
}

// UpdateTokenStatus 更新令牌状态
func (r *PostgresAccessRepository) UpdateTokenStatus(ctx context.Context, id string, status domain.TokenStatus) error {
	res, err := r.db.ExecContext(ctx, `UPDATE activity_tokens SET status = $2 WHERE id = $1;`, id, string(status))
	if err != nil {
		return fmt.Errorf("failed to update token status: %w", err)
	}
	if rows, _ := res.RowsAffected(); rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// RevokeTokens 根据类型撤销令牌
func (r *PostgresAccessRepository) RevokeTokens(ctx context.Context, activityID string, tokenType domain.TokenType) error {
	_, err := r.db.ExecContext(ctx, `UPDATE activity_tokens SET status = $3 WHERE activity_id = $1 AND type = $2 AND status = $4;`,
		activityID, string(tokenType), string(domain.TokenStatusRevoked), string(domain.TokenStatusActive))
	if err != nil {
		return fmt.Errorf("failed to revoke tokens: %w", err)
	}
	return nil
}

// UpsertViewerEntry 新增/更新观众入口
func (r *PostgresAccessRepository) UpsertViewerEntry(ctx context.Context, entry *domain.ViewerEntry) error {
	query := `INSERT INTO viewer_entries (activity_id, share_url, qr_type, qr_content, status, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (activity_id) DO UPDATE SET
			share_url = EXCLUDED.share_url,
			qr_type = EXCLUDED.qr_type,
			qr_content = EXCLUDED.qr_content,
			status = EXCLUDED.status,
			updated_at = EXCLUDED.updated_at;`

	_, err := r.db.ExecContext(
		ctx,
		query,
		entry.ActivityID,
		entry.ShareURL,
		entry.QRType,
		entry.QRContent,
		string(entry.Status),
		entry.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to upsert viewer entry: %w", err)
	}
	return nil
}

// GetViewerEntry 获取观众入口
func (r *PostgresAccessRepository) GetViewerEntry(ctx context.Context, activityID string) (*domain.ViewerEntry, error) {
	query := `SELECT activity_id, share_url, qr_type, qr_content, status, updated_at
		FROM viewer_entries
		WHERE activity_id = $1;`

	row := r.db.QueryRowContext(ctx, query, activityID)

	entry, err := scanViewerEntry(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return entry, nil
}

func scanToken(scanner interface {
	Scan(dest ...any) error
}) (*domain.ActivityToken, error) {
	var (
		id          string
		activityID  string
		tokenType   string
		value       string
		expiresAt   time.Time
		maxAudience sql.NullInt64
		createdAt   time.Time
		status      string
	)
	if err := scanner.Scan(&id, &activityID, &tokenType, &value, &expiresAt, &maxAudience, &createdAt, &status); err != nil {
		return nil, fmt.Errorf("failed to scan token: %w", err)
	}
	var maxAudiencePtr *int
	if maxAudience.Valid {
		v := int(maxAudience.Int64)
		maxAudiencePtr = &v
	}
	return &domain.ActivityToken{
		ID:          id,
		ActivityID:  activityID,
		Type:        domain.TokenType(tokenType),
		Value:       value,
		ExpiresAt:   expiresAt,
		MaxAudience: maxAudiencePtr,
		CreatedAt:   createdAt,
		Status:      domain.TokenStatus(status),
	}, nil
}

func scanViewerEntry(scanner interface {
	Scan(dest ...any) error
}) (*domain.ViewerEntry, error) {
	var (
		activityID string
		shareURL   string
		qrType     string
		qrContent  string
		status     string
		updatedAt  time.Time
	)
	if err := scanner.Scan(&activityID, &shareURL, &qrType, &qrContent, &status, &updatedAt); err != nil {
		return nil, fmt.Errorf("failed to scan viewer entry: %w", err)
	}
	return &domain.ViewerEntry{
		ActivityID: activityID,
		ShareURL:   shareURL,
		QRType:     qrType,
		QRContent:  qrContent,
		Status:     domain.ViewerEntryStatus(status),
		UpdatedAt:  updatedAt,
	}, nil
}

// DeleteViewerEntry 删除观众入口
func (r *PostgresAccessRepository) DeleteViewerEntry(ctx context.Context, activityID string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM viewer_entries WHERE activity_id = $1;`, activityID)
	return err
}
