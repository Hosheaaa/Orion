package domain

import (
	"errors"
	"time"
)

// ActivityStatus 活动状态
type ActivityStatus string

const (
	// ActivityStatusDraft 草稿状态
	ActivityStatusDraft ActivityStatus = "draft"
	// ActivityStatusPublished 已发布
	ActivityStatusPublished ActivityStatus = "published"
	// ActivityStatusClosed 已关闭
	ActivityStatusClosed ActivityStatus = "closed"
)

// Activity 活动实体
type Activity struct {
	ID              string         `json:"id"`
	Title           string         `json:"title"`
	Description     string         `json:"description"`
	Speaker         string         `json:"speaker"`
	StartTime       time.Time      `json:"startTime"`
	EndTime         *time.Time     `json:"endTime,omitempty"`
	InputLanguage   string         `json:"inputLanguage"`   // 输入语种，例如 "zh-CN"
	TargetLanguages []string       `json:"targetLanguages"` // 目标语种列表
	CoverURL        string         `json:"coverUrl,omitempty"`
	Status          ActivityStatus `json:"status"`
	ViewerURL       string         `json:"viewerUrl,omitempty"` // 观众端访问链接
	CreatedAt       time.Time      `json:"createdAt"`
	UpdatedAt       time.Time      `json:"updatedAt"`
}

// Validate 验证活动数据
func (a *Activity) Validate() error {
	if a.Title == "" {
		return errors.New("活动标题不能为空")
	}
	if len(a.Title) > 200 {
		return errors.New("活动标题不能超过200个字符")
	}
	if a.Speaker == "" {
		return errors.New("演讲者不能为空")
	}
	if a.InputLanguage == "" {
		return errors.New("输入语种不能为空")
	}
	if len(a.TargetLanguages) == 0 {
		return errors.New("至少需要一个目标语种")
	}
	if a.StartTime.IsZero() {
		return errors.New("开始时间不能为空")
	}
	return nil
}

// CanPublish 检查是否可以发布
func (a *Activity) CanPublish() bool {
	return a.Status == ActivityStatusDraft && a.Validate() == nil
}

// CanClose 检查是否可以关闭
func (a *Activity) CanClose() bool {
	return a.Status == ActivityStatusPublished
}

// Publish 发布活动
func (a *Activity) Publish() error {
	if !a.CanPublish() {
		return errors.New("活动无法发布，请检查状态和必填字段")
	}
	a.Status = ActivityStatusPublished
	a.UpdatedAt = time.Now()
	return nil
}

// Close 关闭活动
func (a *Activity) Close() error {
	if !a.CanClose() {
		return errors.New("只有已发布的活动才能关闭")
	}
	a.Status = ActivityStatusClosed
	now := time.Now()
	a.EndTime = &now
	a.UpdatedAt = now
	return nil
}

// CreateActivityRequest 创建活动请求
type CreateActivityRequest struct {
	Title           string    `json:"title" binding:"required,max=200"`
	Description     string    `json:"description" binding:"max=2000"`
	Speaker         string    `json:"speaker" binding:"required,max=100"`
	StartTime       time.Time `json:"startTime" binding:"required"`
	InputLanguage   string    `json:"inputLanguage" binding:"required"`
	TargetLanguages []string  `json:"targetLanguages" binding:"required,min=1"`
	CoverURL        string    `json:"coverUrl" binding:"omitempty,url"`
}

// UpdateActivityRequest 更新活动请求
type UpdateActivityRequest struct {
	Title           *string    `json:"title" binding:"omitempty,max=200"`
	Description     *string    `json:"description" binding:"omitempty,max=2000"`
	Speaker         *string    `json:"speaker" binding:"omitempty,max=100"`
	StartTime       *time.Time `json:"startTime"`
	InputLanguage   *string    `json:"inputLanguage"`
	TargetLanguages []string   `json:"targetLanguages" binding:"omitempty,min=1"`
	CoverURL        *string    `json:"coverUrl" binding:"omitempty,url"`
}

// ActivityRepository 活动仓储接口
type ActivityRepository interface {
	// Create 创建活动
	Create(activity *Activity) error
	// Update 更新活动
	Update(activity *Activity) error
	// Delete 删除活动
	Delete(id string) error
	// FindByID 根据 ID 查找活动
	FindByID(id string) (*Activity, error)
	// FindAll 查找所有活动
	FindAll() ([]*Activity, error)
	// FindByStatus 根据状态查找活动
	FindByStatus(status ActivityStatus) ([]*Activity, error)
}
