package handler

import (
	"context"
	"encoding/base64"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	"github.com/hoshea/orion-backend/internal/app"
	"github.com/hoshea/orion-backend/internal/domain"
	"github.com/hoshea/orion-backend/internal/infra/config"
	memrepo "github.com/hoshea/orion-backend/internal/infra/repository"
)

type wsTestAccessRepo struct {
	tokens map[string][]*domain.ActivityToken
	entry  map[string]*domain.ViewerEntry
}

func newWsTestAccessRepo() *wsTestAccessRepo {
	return &wsTestAccessRepo{
		tokens: make(map[string][]*domain.ActivityToken),
		entry:  make(map[string]*domain.ViewerEntry),
	}
}

func (r *wsTestAccessRepo) CreateToken(_ context.Context, token *domain.ActivityToken) error {
	cloned := *token
	r.tokens[token.ActivityID] = append(r.tokens[token.ActivityID], &cloned)
	return nil
}

func (r *wsTestAccessRepo) ListTokens(_ context.Context, activityID string) ([]*domain.ActivityToken, error) {
	var result []*domain.ActivityToken
	for _, token := range r.tokens[activityID] {
		cloned := *token
		result = append(result, &cloned)
	}
	return result, nil
}

func (r *wsTestAccessRepo) FindToken(_ context.Context, activityID string, tokenType domain.TokenType, value string) (*domain.ActivityToken, error) {
	for _, token := range r.tokens[activityID] {
		if token.Type == tokenType && strings.EqualFold(token.Value, value) {
			cloned := *token
			return &cloned, nil
		}
	}
	return nil, nil
}

func (r *wsTestAccessRepo) FindTokenByID(_ context.Context, id string) (*domain.ActivityToken, error) {
	for _, tokens := range r.tokens {
		for _, token := range tokens {
			if token.ID == id {
				cloned := *token
				return &cloned, nil
			}
		}
	}
	return nil, nil
}

func (r *wsTestAccessRepo) UpdateTokenStatus(_ context.Context, id string, status domain.TokenStatus) error {
	for _, tokens := range r.tokens {
		for _, token := range tokens {
			if token.ID == id {
				token.Status = status
				return nil
			}
		}
	}
	return nil
}

func (r *wsTestAccessRepo) RevokeTokens(_ context.Context, activityID string, tokenType domain.TokenType) error {
	for _, token := range r.tokens[activityID] {
		if token.Type == tokenType && token.Status == domain.TokenStatusActive {
			token.Status = domain.TokenStatusRevoked
		}
	}
	return nil
}

func (r *wsTestAccessRepo) UpsertViewerEntry(_ context.Context, entry *domain.ViewerEntry) error {
	cloned := *entry
	r.entry[entry.ActivityID] = &cloned
	return nil
}

func (r *wsTestAccessRepo) GetViewerEntry(_ context.Context, activityID string) (*domain.ViewerEntry, error) {
	entry, ok := r.entry[activityID]
	if !ok {
		return nil, nil
	}
	cloned := *entry
	return &cloned, nil
}

func TestSpeakerWebSocket_WithMockPipeline(t *testing.T) {
	gin.SetMode(gin.TestMode)

	activityRepo := memrepo.NewMemoryActivityRepository()
	cfg := &config.Config{ViewerBaseURL: "http://localhost:3000"}
	activityService := app.NewActivityService(activityRepo, cfg)

	activity, err := activityService.CreateActivity(&domain.CreateActivityRequest{
		Title:           "测试活动",
		Description:     "",
		Speaker:         "演讲者",
		StartTime:       time.Now(),
		InputLanguage:   "zh-CN",
		TargetLanguages: []string{"en"},
	})
	if err != nil {
		t.Fatalf("create activity failed: %v", err)
	}

	accessRepo := newWsTestAccessRepo()
	accessService := app.NewAccessService(activityRepo, accessRepo, cfg.ViewerBaseURL)

	token, err := accessService.GenerateSpeakerToken(activity.ID)
	if err != nil {
		t.Fatalf("generate speaker token failed: %v", err)
	}

	pipeline := app.NewMockTranslationPipeline()
	broadcaster := app.NewSubtitleBroadcaster()
	handler := NewSpeakerWebSocketHandler(pipeline, broadcaster, accessService)

	router := gin.New()
	router.GET("/ws/speaker", handler.HandleSpeakerWebSocket)

	server := httptest.NewServer(router)
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws/speaker?activityId=" + activity.ID + "&token=" + token.Value + "&language=zh-CN"
	dialer := websocket.Dialer{}
	conn, _, err := dialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("dial websocket failed: %v", err)
	}
	defer conn.Close()

	// 读取连接成功后的 READY 状态消息
	mt, message, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("read state message failed: %v", err)
	}
	if mt != websocket.TextMessage {
		t.Fatalf("expected text message, got %v", mt)
	}
	if !strings.Contains(string(message), `"status":"READY"`) {
		t.Fatalf("unexpected initial message: %s", message)
	}

	// 发送音频数据，验证不会触发错误
	payload := base64.StdEncoding.EncodeToString([]byte{0, 1, 2, 3})
	err = conn.WriteMessage(websocket.TextMessage, []byte(`{"type":"AUDIO","payload":{"chunk":"`+payload+`","sequence":1}}`))
	if err != nil {
		t.Fatalf("send audio chunk failed: %v", err)
	}
}
