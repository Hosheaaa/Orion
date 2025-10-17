package repository

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/hoshea/orion-backend/internal/domain"
)

// PostgresActivityRepository PostgreSQL 实现的活动仓储
type PostgresActivityRepository struct {
	db *sql.DB
}

// NewPostgresActivityRepository 构造函数
func NewPostgresActivityRepository(db *sql.DB) *PostgresActivityRepository {
	return &PostgresActivityRepository{db: db}
}

// Create 创建活动
func (r *PostgresActivityRepository) Create(activity *domain.Activity) error {
	if _, err := uuid.Parse(activity.ID); err != nil {
		return fmt.Errorf("invalid activity id: %w", err)
	}

	targetLanguages, err := json.Marshal(activity.TargetLanguages)
	if err != nil {
		return fmt.Errorf("failed to marshal target languages: %w", err)
	}

	query := `INSERT INTO activities (
		id, title, description, speaker, start_time, end_time, input_language,
		target_languages, cover_url, status, viewer_url, created_at, updated_at
	) VALUES (
		$1, $2, $3, $4, $5, $6, $7,
		$8, $9, $10, $11, $12, $13
	);`

	_, err = r.db.Exec(
		query,
		activity.ID,
		activity.Title,
		activity.Description,
		activity.Speaker,
		activity.StartTime,
		activity.EndTime,
		activity.InputLanguage,
		targetLanguages,
		activity.CoverURL,
		activity.Status,
		activity.ViewerURL,
		activity.CreatedAt,
		activity.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to insert activity: %w", err)
	}
	return nil
}

// Update 更新活动
func (r *PostgresActivityRepository) Update(activity *domain.Activity) error {
	targetLanguages, err := json.Marshal(activity.TargetLanguages)
	if err != nil {
		return fmt.Errorf("failed to marshal target languages: %w", err)
	}

	query := `UPDATE activities SET
		title = $2,
		description = $3,
		speaker = $4,
		start_time = $5,
		end_time = $6,
		input_language = $7,
		target_languages = $8,
		cover_url = $9,
		status = $10,
		viewer_url = $11,
		updated_at = $12
	WHERE id = $1;`

	res, err := r.db.Exec(
		query,
		activity.ID,
		activity.Title,
		activity.Description,
		activity.Speaker,
		activity.StartTime,
		activity.EndTime,
		activity.InputLanguage,
		targetLanguages,
		activity.CoverURL,
		activity.Status,
		activity.ViewerURL,
		activity.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to update activity: %w", err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return domain.ErrActivityNotFound
	}
	return nil
}

// Delete 删除活动
func (r *PostgresActivityRepository) Delete(id string) error {
	res, err := r.db.Exec(`DELETE FROM activities WHERE id = $1;`, id)
	if err != nil {
		return fmt.Errorf("failed to delete activity: %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return domain.ErrActivityNotFound
	}
	return nil
}

// FindByID 根据 ID 查找活动
func (r *PostgresActivityRepository) FindByID(id string) (*domain.Activity, error) {
	query := `SELECT
		id, title, description, speaker, start_time, end_time,
		input_language, target_languages, cover_url, status,
		viewer_url, created_at, updated_at
	FROM activities
	WHERE id = $1;`

	row := r.db.QueryRow(query, id)

	activity, err := scanActivity(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrActivityNotFound
		}
		return nil, err
	}
	return activity, nil
}

// FindAll 查找所有活动
func (r *PostgresActivityRepository) FindAll() ([]*domain.Activity, error) {
	query := `SELECT
		id, title, description, speaker, start_time, end_time,
		input_language, target_languages, cover_url, status,
		viewer_url, created_at, updated_at
	FROM activities
	ORDER BY created_at DESC;`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query activities: %w", err)
	}
	defer rows.Close()

	activities := make([]*domain.Activity, 0)
	for rows.Next() {
		activity, err := scanActivity(rows)
		if err != nil {
			return nil, err
		}
		activities = append(activities, activity)
	}
	return activities, rows.Err()
}

// FindByStatus 根据状态查找活动
func (r *PostgresActivityRepository) FindByStatus(status domain.ActivityStatus) ([]*domain.Activity, error) {
	query := `SELECT
		id, title, description, speaker, start_time, end_time,
		input_language, target_languages, cover_url, status,
		viewer_url, created_at, updated_at
	FROM activities
	WHERE status = $1
	ORDER BY start_time DESC;`

	rows, err := r.db.Query(query, status)
	if err != nil {
		return nil, fmt.Errorf("failed to query activities by status: %w", err)
	}
	defer rows.Close()

	activities := make([]*domain.Activity, 0)
	for rows.Next() {
		activity, err := scanActivity(rows)
		if err != nil {
			return nil, err
		}
		activities = append(activities, activity)
	}
	return activities, rows.Err()
}

func scanActivity(scanner interface {
	Scan(dest ...any) error
}) (*domain.Activity, error) {
	var (
		id            string
		title         string
		description   sql.NullString
		speaker       string
		startTime     time.Time
		endTime       sql.NullTime
		inputLanguage string
		targetJSON    []byte
		coverURL      sql.NullString
		status        string
		viewerURL     sql.NullString
		createdAt     time.Time
		updatedAt     time.Time
	)

	if err := scanner.Scan(
		&id,
		&title,
		&description,
		&speaker,
		&startTime,
		&endTime,
		&inputLanguage,
		&targetJSON,
		&coverURL,
		&status,
		&viewerURL,
		&createdAt,
		&updatedAt,
	); err != nil {
		return nil, fmt.Errorf("failed to scan activity: %w", err)
	}

	var targets []string
	if len(targetJSON) > 0 {
		if err := json.Unmarshal(targetJSON, &targets); err != nil {
			return nil, fmt.Errorf("failed to unmarshal target languages: %w", err)
		}
	}

	var endPtr *time.Time
	if endTime.Valid {
		endPtr = &endTime.Time
	}

	return &domain.Activity{
		ID:              id,
		Title:           title,
		Description:     description.String,
		Speaker:         speaker,
		StartTime:       startTime,
		EndTime:         endPtr,
		InputLanguage:   inputLanguage,
		TargetLanguages: targets,
		CoverURL:        coverURL.String,
		Status:          domain.ActivityStatus(status),
		ViewerURL:       viewerURL.String,
		CreatedAt:       createdAt,
		UpdatedAt:       updatedAt,
	}, nil
}
