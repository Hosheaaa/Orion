package app

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/hoshea/orion-backend/internal/domain"
)

const (
	defaultSpeakerTokenTTL = 24 * time.Hour
	defaultViewerTokenTTL  = 120 * time.Minute
	viewerInviteCodeLength = 6
)

// AccessRepository 定义令牌与入口的持久化接口
type AccessRepository interface {
	CreateToken(ctx context.Context, token *domain.ActivityToken) error
	ListTokens(ctx context.Context, activityID string) ([]*domain.ActivityToken, error)
	FindTokenByID(ctx context.Context, id string) (*domain.ActivityToken, error)
	FindToken(ctx context.Context, activityID string, tokenType domain.TokenType, value string) (*domain.ActivityToken, error)
	UpdateTokenStatus(ctx context.Context, id string, status domain.TokenStatus) error
	RevokeTokens(ctx context.Context, activityID string, tokenType domain.TokenType) error
	UpsertViewerEntry(ctx context.Context, entry *domain.ViewerEntry) error
	GetViewerEntry(ctx context.Context, activityID string) (*domain.ViewerEntry, error)
}

// AccessService 负责活动令牌与观众入口管理
type AccessService struct {
	activityRepo domain.ActivityRepository
	repo         AccessRepository
	viewerBase   string
}

// NewAccessService 创建访问控制服务
func NewAccessService(activityRepo domain.ActivityRepository, repo AccessRepository, viewerBaseURL string) *AccessService {
	base := strings.TrimRight(viewerBaseURL, "/")
	return &AccessService{
		activityRepo: activityRepo,
		repo:         repo,
		viewerBase:   base,
	}
}

// GenerateSpeakerToken 生成演讲者令牌
func (s *AccessService) GenerateSpeakerToken(activityID string) (*domain.ActivityToken, error) {
	if _, err := s.activityRepo.FindByID(activityID); err != nil {
		return nil, err
	}

	now := time.Now()
	token := &domain.ActivityToken{
		ID:         uuid.NewString(),
		ActivityID: activityID,
		Type:       domain.TokenTypeSpeaker,
		Value:      uuid.NewString(),
		CreatedAt:  now,
		ExpiresAt:  now.Add(defaultSpeakerTokenTTL),
		Status:     domain.TokenStatusActive,
	}

	if err := s.repo.CreateToken(context.Background(), token); err != nil {
		return nil, err
	}
	return cloneToken(token), nil
}

// RevokeSpeakerTokens 撤销活动下所有演讲者令牌
func (s *AccessService) RevokeSpeakerTokens(activityID string) error {
	if _, err := s.activityRepo.FindByID(activityID); err != nil {
		return err
	}

	return s.repo.RevokeTokens(context.Background(), activityID, domain.TokenTypeSpeaker)
}

// RevokeSpeakerToken 撤销单个演讲者令牌
func (s *AccessService) RevokeSpeakerToken(activityID, tokenID string) error {
	if tokenID == "" {
		return errors.New("令牌ID不能为空")
	}

	token, err := s.repo.FindTokenByID(context.Background(), tokenID)
	if err != nil {
		return err
	}
	if token == nil || token.ActivityID != activityID {
		return errors.New("演讲者令牌不存在")
	}
	if token.Type != domain.TokenTypeSpeaker {
		return errors.New("令牌类型不匹配")
	}
	if token.Status == domain.TokenStatusRevoked {
		return nil
	}

	if err := s.repo.UpdateTokenStatus(context.Background(), tokenID, domain.TokenStatusRevoked); err != nil {
		return err
	}
	return nil
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
	code := strings.ToUpper(generateInviteCode(viewerInviteCodeLength))
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
	entry := &domain.ViewerEntry{
		ActivityID: activityID,
		ShareURL:   shareURL,
		QRType:     "text",
		QRContent:  encodeTextAsDataURL(shareURL),
		Status:     domain.ViewerEntryStatusActive,
		UpdatedAt:  now,
	}

	ctx := context.Background()

	if err := s.repo.RevokeTokens(ctx, activityID, domain.TokenTypeViewer); err != nil {
		return nil, err
	}
	if err := s.repo.CreateToken(ctx, token); err != nil {
		return nil, err
	}
	if err := s.repo.UpsertViewerEntry(ctx, entry); err != nil {
		return nil, err
	}

	return cloneToken(token), nil
}

// ListTokens 列出活动的所有令牌（自动刷新过期状态）
func (s *AccessService) ListTokens(activityID string) ([]*domain.ActivityToken, error) {
	if _, err := s.activityRepo.FindByID(activityID); err != nil {
		return nil, err
	}

	ctx := context.Background()
	tokens, err := s.repo.ListTokens(ctx, activityID)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	result := make([]*domain.ActivityToken, 0, len(tokens))
	for _, token := range tokens {
		if token.Status == domain.TokenStatusActive && now.After(token.ExpiresAt) {
			_ = s.repo.UpdateTokenStatus(ctx, token.ID, domain.TokenStatusExpired)
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

	entry, err := s.repo.GetViewerEntry(context.Background(), activityID)
	if err != nil {
		return nil, err
	}
	if entry != nil {
		return cloneViewerEntry(entry), nil
	}

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

	ctx := context.Background()
	entry, err := s.repo.GetViewerEntry(ctx, activityID)
	if err != nil {
		return nil, err
	}
	if entry == nil {
		return nil, errors.New("观众入口尚未生成")
	}

	entry.Status = domain.ViewerEntryStatusRevoked
	entry.QRContent = ""
	entry.UpdatedAt = time.Now()

	if err := s.repo.RevokeTokens(ctx, activityID, domain.TokenTypeViewer); err != nil {
		return nil, err
	}
	if err := s.repo.UpsertViewerEntry(ctx, entry); err != nil {
		return nil, err
	}
	return cloneViewerEntry(entry), nil
}

// ActivateViewerEntry 重新启用观众入口
func (s *AccessService) ActivateViewerEntry(activityID string) (*domain.ViewerEntry, error) {
	if _, err := s.activityRepo.FindByID(activityID); err != nil {
		return nil, err
	}

	ctx := context.Background()
	entry, err := s.repo.GetViewerEntry(ctx, activityID)
	if err != nil {
		return nil, err
	}
	if entry == nil {
		return nil, errors.New("观众入口尚未生成")
	}

	tokens, err := s.repo.ListTokens(ctx, activityID)
	if err != nil {
		return nil, err
	}
	var latest *domain.ActivityToken
	for _, t := range tokens {
		if t.Type == domain.TokenTypeViewer {
			if latest == nil || t.CreatedAt.After(latest.CreatedAt) {
				latest = t
			}
		}
	}
	if latest == nil {
		return nil, errors.New("请先生成观众邀请码")
	}

	if time.Now().After(latest.ExpiresAt) {
		_ = s.repo.UpdateTokenStatus(ctx, latest.ID, domain.TokenStatusExpired)
		return nil, errors.New("最新观众邀请码已过期，请重新生成")
	}

	entry.Status = domain.ViewerEntryStatusActive
	entry.ShareURL = s.buildShareURL(activityID, latest.Value)
	entry.QRType = "text"
	entry.QRContent = encodeTextAsDataURL(entry.ShareURL)
	entry.UpdatedAt = time.Now()

	if err := s.repo.UpsertViewerEntry(ctx, entry); err != nil {
		return nil, err
	}
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

	ctx := context.Background()
	token, err := s.repo.FindToken(ctx, activityID, domain.TokenTypeSpeaker, tokenValue)
	if err != nil {
		return nil, err
	}
	if token == nil {
		return nil, errors.New("演讲者令牌无效")
	}

	now := time.Now()
	if token.Status == domain.TokenStatusActive && now.After(token.ExpiresAt) {
		_ = s.repo.UpdateTokenStatus(ctx, token.ID, domain.TokenStatusExpired)
		token.Status = domain.TokenStatusExpired
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

	ctx := context.Background()
	entry, err := s.repo.GetViewerEntry(ctx, activityID)
	if err != nil {
		return nil, err
	}
	if entry == nil || entry.Status != domain.ViewerEntryStatusActive {
		return nil, errors.New("观众入口未启用，请联系主办方")
	}

	token, err := s.repo.FindToken(ctx, activityID, domain.TokenTypeViewer, normalizedCode)
	if err != nil {
		return nil, err
	}
	if token == nil {
		return nil, errors.New("观众令牌无效")
	}

	now := time.Now()
	if token.Status == domain.TokenStatusActive && now.After(token.ExpiresAt) {
		_ = s.repo.UpdateTokenStatus(ctx, token.ID, domain.TokenStatusExpired)
		token.Status = domain.TokenStatusExpired
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
