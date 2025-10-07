package app

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/hoshea/orion-backend/internal/domain"
	"github.com/hoshea/orion-backend/internal/infra/config"
)

// ActivityService 活动管理服务
type ActivityService struct {
	repo          domain.ActivityRepository
	viewerBaseURL string
}

// NewActivityService 创建活动服务
func NewActivityService(repo domain.ActivityRepository, cfg *config.Config) *ActivityService {
	return &ActivityService{
		repo:          repo,
		viewerBaseURL: cfg.ViewerBaseURL,
	}
}

// CreateActivity 创建活动
func (s *ActivityService) CreateActivity(req *domain.CreateActivityRequest) (*domain.Activity, error) {
	// 生成活动 ID
	activityID := uuid.New().String()

	// 生成观众端访问链接（初期简化版）
	viewerURL := fmt.Sprintf("%s/activity/%s", s.viewerBaseURL, activityID)

	// 创建活动实体
	now := time.Now()
	activity := &domain.Activity{
		ID:              activityID,
		Title:           req.Title,
		Description:     req.Description,
		Speaker:         req.Speaker,
		StartTime:       req.StartTime,
		InputLanguage:   req.InputLanguage,
		TargetLanguages: req.TargetLanguages,
		CoverURL:        req.CoverURL,
		Status:          domain.ActivityStatusDraft,
		ViewerURL:       viewerURL,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	// 验证活动数据
	if err := activity.Validate(); err != nil {
		return nil, fmt.Errorf("活动数据验证失败: %w", err)
	}

	// 保存到仓储
	if err := s.repo.Create(activity); err != nil {
		return nil, fmt.Errorf("创建活动失败: %w", err)
	}

	return activity, nil
}

// UpdateActivity 更新活动
func (s *ActivityService) UpdateActivity(id string, req *domain.UpdateActivityRequest) (*domain.Activity, error) {
	// 查找活动
	activity, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}

	// 检查活动状态（已关闭的活动不允许修改）
	if activity.Status == domain.ActivityStatusClosed {
		return nil, domain.ErrActivityCannotBeModified
	}

	// 更新字段
	if req.Title != nil {
		activity.Title = *req.Title
	}
	if req.Description != nil {
		activity.Description = *req.Description
	}
	if req.Speaker != nil {
		activity.Speaker = *req.Speaker
	}
	if req.StartTime != nil {
		activity.StartTime = *req.StartTime
	}
	if req.InputLanguage != nil {
		activity.InputLanguage = *req.InputLanguage
	}
	if req.TargetLanguages != nil {
		activity.TargetLanguages = req.TargetLanguages
	}
	if req.CoverURL != nil {
		activity.CoverURL = *req.CoverURL
	}

	activity.UpdatedAt = time.Now()

	// 验证更新后的数据
	if err := activity.Validate(); err != nil {
		return nil, fmt.Errorf("活动数据验证失败: %w", err)
	}

	// 保存更新
	if err := s.repo.Update(activity); err != nil {
		return nil, fmt.Errorf("更新活动失败: %w", err)
	}

	return activity, nil
}

// GetActivity 获取活动详情
func (s *ActivityService) GetActivity(id string) (*domain.Activity, error) {
	return s.repo.FindByID(id)
}

// ListActivities 列出活动
func (s *ActivityService) ListActivities(status *domain.ActivityStatus) ([]*domain.Activity, error) {
	if status != nil {
		return s.repo.FindByStatus(*status)
	}
	return s.repo.FindAll()
}

// PublishActivity 发布活动
func (s *ActivityService) PublishActivity(id string) (*domain.Activity, error) {
	activity, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}

	// 发布活动
	if err := activity.Publish(); err != nil {
		return nil, err
	}

	// 保存状态变更
	if err := s.repo.Update(activity); err != nil {
		return nil, fmt.Errorf("发布活动失败: %w", err)
	}

	return activity, nil
}

// CloseActivity 关闭活动
func (s *ActivityService) CloseActivity(id string) (*domain.Activity, error) {
	activity, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}

	// 关闭活动
	if err := activity.Close(); err != nil {
		return nil, err
	}

	// 保存状态变更
	if err := s.repo.Update(activity); err != nil {
		return nil, fmt.Errorf("关闭活动失败: %w", err)
	}

	// TODO: 关闭活动时自动失效二维码

	return activity, nil
}

// DeleteActivity 删除活动
func (s *ActivityService) DeleteActivity(id string) error {
	// 软删除策略：只允许删除草稿状态的活动
	activity, err := s.repo.FindByID(id)
	if err != nil {
		return err
	}

	if activity.Status != domain.ActivityStatusDraft {
		return fmt.Errorf("只能删除草稿状态的活动")
	}

	return s.repo.Delete(id)
}
