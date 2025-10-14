package app

import (
	"encoding/base64"
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/hoshea/orion-backend/internal/domain"
)

const (
	defaultSpeakerTokenTTL = 2 * time.Hour
	defaultViewerTokenTTL  = 120 * time.Minute
	viewerInviteCodeLength = 6
)

// AccessService 负责活动令牌与观众入口管理（内存实现）
type AccessService struct {
	activityRepo domain.ActivityRepository
	viewerBase   string
	mu           sync.RWMutex
	tokens       map[string][]*domain.ActivityToken
	viewerEntry  map[string]*domain.ViewerEntry
}

// NewAccessService 创建访问控制服务
func NewAccessService(activityRepo domain.ActivityRepository, viewerBaseURL string) *AccessService {
	base := strings.TrimRight(viewerBaseURL, "/")
	return &AccessService{
		activityRepo: activityRepo,
		viewerBase:   base,
		tokens:       make(map[string][]*domain.ActivityToken),
		viewerEntry:  make(map[string]*domain.ViewerEntry),
	}
}

// GenerateSpeakerToken 生成演讲者令牌
func (s *AccessService) GenerateSpeakerToken(activityID string) (*domain.ActivityToken, error) {
	if _, err := s.activityRepo.FindByID(activityID); err != nil {
		return nil, err
	}

	now := time.Now()
	value := uuid.NewString()
	token := &domain.ActivityToken{
		ID:         uuid.NewString(),
		ActivityID: activityID,
		Type:       domain.TokenTypeSpeaker,
		Value:      value,
		CreatedAt:  now,
		ExpiresAt:  now.Add(defaultSpeakerTokenTTL),
		Status:     domain.TokenStatusActive,
	}

	s.mu.Lock()
	s.tokens[activityID] = append(s.tokens[activityID], token)
	s.mu.Unlock()

	return cloneToken(token), nil
}

// GenerateViewerToken 生成观众邀请码
func (s *AccessService) GenerateViewerToken(activityID string, req *domain.GenerateViewerTokenRequest) (*domain.ActivityToken, error) {
	if _, err := s.activityRepo.FindByID(activityID); err != nil {
		return nil, err
	}

	ttl := defaultViewerTokenTTL
	if req != nil && req.TTLMinutes > 0 {
		ttl = time.Duration(req.TTLMinutes) * time.Minute
	}

	now := time.Now()
	code := generateInviteCode(viewerInviteCodeLength)
	token := &domain.ActivityToken{
		ID:         uuid.NewString(),
		ActivityID: activityID,
		Type:       domain.TokenTypeViewer,
		Value:      code,
		CreatedAt:  now,
		ExpiresAt:  now.Add(ttl),
		Status:     domain.TokenStatusActive,
	}
	if req != nil && req.MaxAudience > 0 {
		token.MaxAudience = ptr(req.MaxAudience)
	}

	shareURL := s.buildShareURL(activityID, code)
	viewerEntry := &domain.ViewerEntry{
		ActivityID: activityID,
		ShareURL:   shareURL,
		QRType:     "text",
		QRContent:  encodeTextAsDataURL(shareURL),
		Status:     domain.ViewerEntryStatusActive,
		UpdatedAt:  now,
	}

	s.mu.Lock()
	// 将旧的观众令牌标记为撤销
	for _, existing := range s.tokens[activityID] {
		if existing.Type == domain.TokenTypeViewer && existing.Status == domain.TokenStatusActive {
			existing.Status = domain.TokenStatusRevoked
		}
	}
	s.tokens[activityID] = append(s.tokens[activityID], token)
	s.viewerEntry[activityID] = viewerEntry
	s.mu.Unlock()

	return cloneToken(token), nil
}

// ListTokens 列出活动的所有令牌
func (s *AccessService) ListTokens(activityID string) ([]*domain.ActivityToken, error) {
	if _, err := s.activityRepo.FindByID(activityID); err != nil {
		return nil, err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	tokens := s.tokens[activityID]
	result := make([]*domain.ActivityToken, 0, len(tokens))
	now := time.Now()
	for _, token := range tokens {
		if token.Status == domain.TokenStatusActive && now.After(token.ExpiresAt) {
			token.Status = domain.TokenStatusExpired
		}
		result = append(result, cloneToken(token))
	}

	return result, nil
}

// GetViewerEntry 获取观众入口信息
func (s *AccessService) GetViewerEntry(activityID string) (*domain.ViewerEntry, error) {
	activity, err := s.activityRepo.FindByID(activityID)
	if err != nil {
		return nil, err
	}

	s.mu.RLock()
	entry, ok := s.viewerEntry[activityID]
	s.mu.RUnlock()
	if ok {
		return cloneViewerEntry(entry), nil
	}

	// 默认返回活动链接
	defaultEntry := &domain.ViewerEntry{
		ActivityID: activityID,
		ShareURL:   activity.ViewerURL,
		QRType:     "text",
		QRContent:  encodeTextAsDataURL(activity.ViewerURL),
		Status:     domain.ViewerEntryStatusInactive,
		UpdatedAt:  time.Now(),
	}
	return defaultEntry, nil
}

// RevokeViewerEntry 失效观众入口
func (s *AccessService) RevokeViewerEntry(activityID string) (*domain.ViewerEntry, error) {
	if _, err := s.activityRepo.FindByID(activityID); err != nil {
		return nil, err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	entry, ok := s.viewerEntry[activityID]
	if !ok {
		return nil, errors.New("观众入口尚未生成")
	}
	entry.Status = domain.ViewerEntryStatusRevoked
	entry.UpdatedAt = time.Now()
	entry.QRContent = ""

	// 同时撤销所有观众令牌
	for _, token := range s.tokens[activityID] {
		if token.Type == domain.TokenTypeViewer && token.Status == domain.TokenStatusActive {
			token.Status = domain.TokenStatusRevoked
		}
	}

	return cloneViewerEntry(entry), nil
}

// ActivateViewerEntry 重新启用观众入口
func (s *AccessService) ActivateViewerEntry(activityID string) (*domain.ViewerEntry, error) {
	if _, err := s.activityRepo.FindByID(activityID); err != nil {
		return nil, err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	entry, ok := s.viewerEntry[activityID]
	if !ok {
		return nil, errors.New("观众入口尚未生成")
	}

	var latest *domain.ActivityToken
	for i := len(s.tokens[activityID]) - 1; i >= 0; i-- {
		t := s.tokens[activityID][i]
		if t.Type == domain.TokenTypeViewer {
			latest = t
			break
		}
	}

	if latest == nil {
		return nil, errors.New("请先生成观众邀请码")
	}

	if time.Now().After(latest.ExpiresAt) {
		latest.Status = domain.TokenStatusExpired
		return nil, errors.New("最新观众邀请码已过期，请重新生成")
	}

	entry.Status = domain.ViewerEntryStatusActive
	entry.ShareURL = s.buildShareURL(activityID, latest.Value)
	entry.QRType = "text"
	entry.QRContent = encodeTextAsDataURL(entry.ShareURL)
	entry.UpdatedAt = time.Now()
	return cloneViewerEntry(entry), nil
}

func (s *AccessService) buildShareURL(activityID, code string) string {
	return fmt.Sprintf("%s/activity/%s?code=%s", s.viewerBase, activityID, code)
}

// ValidateSpeakerSession 校验演讲者接入令牌与语言
func (s *AccessService) ValidateSpeakerSession(activityID, tokenValue, language string) (*domain.Activity, error) {
	tokenValue = strings.TrimSpace(tokenValue)
	if tokenValue == "" {
		return nil, errors.New("演讲者令牌不能为空")
	}

	activity, err := s.activityRepo.FindByID(activityID)
	if err != nil {
		return nil, err
	}
	if activity.Status == domain.ActivityStatusClosed {
		return nil, errors.New("活动已关闭，无法继续推流")
	}

	if language != "" && !strings.EqualFold(language, activity.InputLanguage) {
		return nil, fmt.Errorf("演讲语言与活动配置不一致: %s", language)
	}

	now := time.Now()

	s.mu.Lock()
	defer s.mu.Unlock()

	tokens := s.tokens[activityID]
	for _, token := range tokens {
		if token.Status == domain.TokenStatusActive && now.After(token.ExpiresAt) {
			token.Status = domain.TokenStatusExpired
		}

		if token.Type != domain.TokenTypeSpeaker {
			continue
		}
		if token.Value != tokenValue {
			continue
		}

		switch token.Status {
		case domain.TokenStatusActive:
			return activity, nil
		case domain.TokenStatusRevoked:
			return nil, errors.New("演讲者令牌已被撤销")
		case domain.TokenStatusExpired:
			return nil, errors.New("演讲者令牌已过期")
		default:
			return nil, errors.New("演讲者令牌状态异常")
		}
	}

	return nil, errors.New("演讲者令牌无效")
}

// ValidateViewerSession 校验观众接入令牌与语言
func (s *AccessService) ValidateViewerSession(activityID, tokenValue, language string) (*domain.Activity, error) {
	tokenValue = strings.TrimSpace(tokenValue)
	if tokenValue == "" {
		return nil, errors.New("观众令牌不能为空")
	}
	if language == "" {
		return nil, errors.New("观众订阅语言不能为空")
	}

	normalizedCode := strings.ToUpper(tokenValue)

	activity, err := s.activityRepo.FindByID(activityID)
	if err != nil {
		return nil, err
	}
	if activity.Status != domain.ActivityStatusPublished {
		return nil, errors.New("活动尚未发布，暂不支持观众接入")
	}

	if !supportsLanguage(activity, language) {
		return nil, fmt.Errorf("活动未启用语言: %s", language)
	}

	now := time.Now()

	s.mu.Lock()
	defer s.mu.Unlock()

	entry, ok := s.viewerEntry[activityID]
	if !ok || entry.Status != domain.ViewerEntryStatusActive {
		return nil, errors.New("观众入口未启用，请联系主办方")
	}

	tokens := s.tokens[activityID]
	for _, token := range tokens {
		if token.Status == domain.TokenStatusActive && now.After(token.ExpiresAt) {
			token.Status = domain.TokenStatusExpired
		}

		if token.Type != domain.TokenTypeViewer {
			continue
		}
		if strings.ToUpper(token.Value) != normalizedCode {
			continue
		}

		switch token.Status {
		case domain.TokenStatusActive:
			return activity, nil
		case domain.TokenStatusRevoked:
			return nil, errors.New("观众令牌已被撤销")
		case domain.TokenStatusExpired:
			return nil, errors.New("观众令牌已过期")
		default:
			return nil, errors.New("观众令牌状态异常")
		}
	}

	return nil, errors.New("观众令牌无效")
}

func cloneToken(src *domain.ActivityToken) *domain.ActivityToken {
	if src == nil {
		return nil
	}
	clone := *src
	return &clone
}

func cloneViewerEntry(src *domain.ViewerEntry) *domain.ViewerEntry {
	if src == nil {
		return nil
	}
	clone := *src
	return &clone
}

func ptr(value int) *int {
	return &value
}

func generateInviteCode(length int) string {
	const letters = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func encodeTextAsDataURL(content string) string {
	encoded := base64.StdEncoding.EncodeToString([]byte(content))
	return "data:text/plain;base64," + encoded
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

func supportsLanguage(activity *domain.Activity, language string) bool {
	if strings.EqualFold(language, activity.InputLanguage) {
		return true
	}
	for _, target := range activity.TargetLanguages {
		if strings.EqualFold(target, language) {
			return true
		}
	}
	return false
}
