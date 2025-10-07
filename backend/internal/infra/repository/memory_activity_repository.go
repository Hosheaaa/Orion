package repository

import (
	"sync"

	"github.com/hoshea/orion-backend/internal/domain"
)

// MemoryActivityRepository 基于内存的活动仓储实现
// 注意：这是开发阶段的简单实现，生产环境应使用数据库
type MemoryActivityRepository struct {
	mu         sync.RWMutex
	activities map[string]*domain.Activity
}

// NewMemoryActivityRepository 创建内存活动仓储
func NewMemoryActivityRepository() *MemoryActivityRepository {
	return &MemoryActivityRepository{
		activities: make(map[string]*domain.Activity),
	}
}

// Create 创建活动
func (r *MemoryActivityRepository) Create(activity *domain.Activity) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.activities[activity.ID]; exists {
		return domain.ErrActivityAlreadyExists
	}

	r.activities[activity.ID] = activity
	return nil
}

// Update 更新活动
func (r *MemoryActivityRepository) Update(activity *domain.Activity) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.activities[activity.ID]; !exists {
		return domain.ErrActivityNotFound
	}

	r.activities[activity.ID] = activity
	return nil
}

// Delete 删除活动
func (r *MemoryActivityRepository) Delete(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.activities[id]; !exists {
		return domain.ErrActivityNotFound
	}

	delete(r.activities, id)
	return nil
}

// FindByID 根据 ID 查找活动
func (r *MemoryActivityRepository) FindByID(id string) (*domain.Activity, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	activity, exists := r.activities[id]
	if !exists {
		return nil, domain.ErrActivityNotFound
	}

	// 返回副本，避免外部修改
	return copyActivity(activity), nil
}

// FindAll 查找所有活动
func (r *MemoryActivityRepository) FindAll() ([]*domain.Activity, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	activities := make([]*domain.Activity, 0, len(r.activities))
	for _, activity := range r.activities {
		activities = append(activities, copyActivity(activity))
	}

	return activities, nil
}

// FindByStatus 根据状态查找活动
func (r *MemoryActivityRepository) FindByStatus(status domain.ActivityStatus) ([]*domain.Activity, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	activities := make([]*domain.Activity, 0)
	for _, activity := range r.activities {
		if activity.Status == status {
			activities = append(activities, copyActivity(activity))
		}
	}

	return activities, nil
}

// copyActivity 复制活动对象（深拷贝）
func copyActivity(src *domain.Activity) *domain.Activity {
	if src == nil {
		return nil
	}

	dst := &domain.Activity{
		ID:              src.ID,
		Title:           src.Title,
		Description:     src.Description,
		Speaker:         src.Speaker,
		StartTime:       src.StartTime,
		InputLanguage:   src.InputLanguage,
		TargetLanguages: make([]string, len(src.TargetLanguages)),
		CoverURL:        src.CoverURL,
		Status:          src.Status,
		ViewerURL:       src.ViewerURL,
		CreatedAt:       src.CreatedAt,
		UpdatedAt:       src.UpdatedAt,
	}

	copy(dst.TargetLanguages, src.TargetLanguages)

	if src.EndTime != nil {
		endTime := *src.EndTime
		dst.EndTime = &endTime
	}

	return dst
}
