package app

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/hoshea/orion-backend/internal/domain"
	"github.com/hoshea/orion-backend/internal/infra/config"
	"github.com/hoshea/orion-backend/internal/infra/repository"
)

type fakeAccessRepo struct {
	tokens map[string][]*domain.ActivityToken
	entry  map[string]*domain.ViewerEntry
}

func newFakeAccessRepo() *fakeAccessRepo {
	return &fakeAccessRepo{
		tokens: make(map[string][]*domain.ActivityToken),
		entry:  make(map[string]*domain.ViewerEntry),
	}
}

func (f *fakeAccessRepo) CreateToken(_ context.Context, token *domain.ActivityToken) error {
	cloned := *token
	f.tokens[token.ActivityID] = append(f.tokens[token.ActivityID], &cloned)
	return nil
}

func (f *fakeAccessRepo) ListTokens(_ context.Context, activityID string) ([]*domain.ActivityToken, error) {
	var result []*domain.ActivityToken
	for _, token := range f.tokens[activityID] {
		cloned := *token
		result = append(result, &cloned)
	}
	return result, nil
}

func (f *fakeAccessRepo) FindToken(_ context.Context, activityID string, tokenType domain.TokenType, value string) (*domain.ActivityToken, error) {
	for _, token := range f.tokens[activityID] {
		if token.Type == tokenType && strings.EqualFold(token.Value, value) {
			cloned := *token
			return &cloned, nil
		}
	}
	return nil, nil
}

func (f *fakeAccessRepo) FindTokenByID(_ context.Context, id string) (*domain.ActivityToken, error) {
	for _, tokens := range f.tokens {
		for _, token := range tokens {
			if token.ID == id {
				cloned := *token
				return &cloned, nil
			}
		}
	}
	return nil, nil
}

func (f *fakeAccessRepo) UpdateTokenStatus(_ context.Context, id string, status domain.TokenStatus) error {
	for _, tokens := range f.tokens {
		for _, token := range tokens {
			if token.ID == id {
				token.Status = status
				return nil
			}
		}
	}
	return domain.ErrActivityNotFound
}

func (f *fakeAccessRepo) RevokeTokens(_ context.Context, activityID string, tokenType domain.TokenType) error {
	for _, token := range f.tokens[activityID] {
		if token.Type == tokenType && token.Status == domain.TokenStatusActive {
			token.Status = domain.TokenStatusRevoked
		}
	}
	return nil
}

func (f *fakeAccessRepo) UpsertViewerEntry(_ context.Context, entry *domain.ViewerEntry) error {
	cloned := *entry
	f.entry[entry.ActivityID] = &cloned
	return nil
}

func (f *fakeAccessRepo) GetViewerEntry(_ context.Context, activityID string) (*domain.ViewerEntry, error) {
	entry, ok := f.entry[activityID]
	if !ok {
		return nil, nil
	}
	cloned := *entry
	return &cloned, nil
}

func TestAccessService_GenerateAndValidateViewerToken(t *testing.T) {
	activityRepo := repository.NewMemoryActivityRepository()
	cfg := &config.Config{ViewerBaseURL: "http://localhost:3000"}
	service := NewActivityService(activityRepo, cfg)

	activity, err := service.CreateActivity(&domain.CreateActivityRequest{
		Title:           "Demo",
		Description:     "Demo activity",
		Speaker:         "Tester",
		StartTime:       time.Now().Add(10 * time.Minute),
		InputLanguage:   "zh-CN",
		TargetLanguages: []string{"en"},
	})
	if err != nil {
		t.Fatalf("create activity failed: %v", err)
	}

	if _, err := service.PublishActivity(activity.ID); err != nil {
		t.Fatalf("publish activity failed: %v", err)
	}

	accessRepo := newFakeAccessRepo()
	accessService := NewAccessService(activityRepo, accessRepo, cfg.ViewerBaseURL)

	token, err := accessService.GenerateViewerToken(activity.ID, &domain.GenerateViewerTokenRequest{TTLMinutes: 5})
	if err != nil {
		t.Fatalf("generate viewer token failed: %v", err)
	}
	if token.Value == "" {
		t.Fatalf("expected token value not empty")
	}

	entry, err := accessService.GetViewerEntry(activity.ID)
	if err != nil {
		t.Fatalf("get viewer entry failed: %v", err)
	}
	if entry.Status != domain.ViewerEntryStatusActive {
		t.Fatalf("expected active entry, got %s", entry.Status)
	}

	if _, err := accessService.ValidateViewerSession(activity.ID, token.Value, "en"); err != nil {
		t.Fatalf("validate viewer session failed: %v", err)
	}

	// 模拟过期
	accessRepo.tokens[activity.ID][0].ExpiresAt = time.Now().Add(-time.Minute)
	if _, err := accessService.ValidateViewerSession(activity.ID, token.Value, "en"); err == nil {
		t.Fatalf("expected error for expired token")
	}
}

func TestAccessService_GenerateSpeakerToken(t *testing.T) {
	activityRepo := repository.NewMemoryActivityRepository()
	cfg := &config.Config{ViewerBaseURL: "http://localhost:3000"}
	service := NewActivityService(activityRepo, cfg)

	activity, err := service.CreateActivity(&domain.CreateActivityRequest{
		Title:           "Speaker session",
		Description:     "",
		Speaker:         "Speaker",
		StartTime:       time.Now(),
		InputLanguage:   "zh-CN",
		TargetLanguages: []string{"en"},
	})
	if err != nil {
		t.Fatalf("create activity failed: %v", err)
	}

	accessRepo := newFakeAccessRepo()
	accessService := NewAccessService(activityRepo, accessRepo, cfg.ViewerBaseURL)

	token, err := accessService.GenerateSpeakerToken(activity.ID)
	if err != nil {
		t.Fatalf("generate speaker token failed: %v", err)
	}
	if _, err := uuid.Parse(token.ID); err != nil {
		t.Fatalf("invalid token id: %v", err)
	}
	if _, err := accessService.ValidateSpeakerSession(activity.ID, token.Value, "zh-CN"); err != nil {
		t.Fatalf("validate speaker session failed: %v", err)
	}
}
