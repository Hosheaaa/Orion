package domain

import "time"

// TokenType 令牌类型
type TokenType string

const (
	TokenTypeSpeaker TokenType = "speaker"
	TokenTypeViewer  TokenType = "viewer"
)

// TokenStatus 令牌状态
type TokenStatus string

const (
	TokenStatusActive  TokenStatus = "active"
	TokenStatusRevoked TokenStatus = "revoked"
	TokenStatusExpired TokenStatus = "expired"
)

// ActivityToken 活动令牌实体
type ActivityToken struct {
	ID          string      `json:"id"`
	ActivityID  string      `json:"activityId"`
	Type        TokenType   `json:"type"`
	Value       string      `json:"value"`
	ExpiresAt   time.Time   `json:"expiresAt"`
	MaxAudience *int        `json:"maxAudience,omitempty"`
	CreatedAt   time.Time   `json:"createdAt"`
	Status      TokenStatus `json:"status"`
}

// GenerateViewerTokenRequest 观众令牌生成请求
type GenerateViewerTokenRequest struct {
	MaxAudience int `json:"maxAudience" binding:"omitempty,min=1"`
	TTLMinutes  int `json:"ttlMinutes" binding:"omitempty,min=1"`
}

// ViewerEntryStatus 观众入口状态
type ViewerEntryStatus string

const (
	ViewerEntryStatusInactive ViewerEntryStatus = "inactive"
	ViewerEntryStatusActive   ViewerEntryStatus = "active"
	ViewerEntryStatusRevoked  ViewerEntryStatus = "revoked"
)

// ViewerEntry 观众入口信息
type ViewerEntry struct {
	ActivityID string            `json:"activityId"`
	ShareURL   string            `json:"shareUrl"`
	QRType     string            `json:"qrType"`
	QRContent  string            `json:"qrContent"`
	Status     ViewerEntryStatus `json:"status"`
	UpdatedAt  time.Time         `json:"updatedAt"`
}
