package handler

import (
	"encoding/base64"
	"encoding/json"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/hoshea/orion-backend/internal/app"
	"github.com/hoshea/orion-backend/internal/domain"
	ws "github.com/hoshea/orion-backend/internal/infra/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // 生产环境应该检查 Origin
	},
}

// SpeakerWebSocketHandler 演讲者 WebSocket 处理器
type SpeakerWebSocketHandler struct {
	pipeline    *app.TranslationPipeline
	broadcaster *app.SubtitleBroadcaster
}

// NewSpeakerWebSocketHandler 创建演讲者处理器
func NewSpeakerWebSocketHandler(pipeline *app.TranslationPipeline, broadcaster *app.SubtitleBroadcaster) *SpeakerWebSocketHandler {
	return &SpeakerWebSocketHandler{
		pipeline:    pipeline,
		broadcaster: broadcaster,
	}
}

// HandleSpeakerWebSocket 处理演讲者 WebSocket 连接
func (h *SpeakerWebSocketHandler) HandleSpeakerWebSocket(c *gin.Context) {
	// 升级为 WebSocket 连接
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("Failed to upgrade connection: %v", err)
		return
	}

	// 创建连接封装
	connectionID := uuid.New().String()
	wsConn := ws.NewConnection(connectionID, conn)

	log.Printf("Speaker WebSocket connected: %s", connectionID)

	// 等待认证消息
	authPayload, err := h.authenticate(wsConn, c)
	if err != nil {
		log.Printf("Authentication failed: %v", err)
		wsConn.SendJSON(domain.MessageTypeError, domain.ErrorPayload{
			Code:    "AUTH_FAILED",
			Message: "认证失败: " + err.Error(),
		})
		wsConn.Close()
		return
	}

	// 启动翻译会话
	session, err := h.pipeline.StartSession(
		authPayload.ActivityID,
		authPayload.Language,
		[]string{"en", "ja", "es"}, // TODO: 从活动配置中获取目标语言
	)
	if err != nil {
		log.Printf("Failed to start translation session: %v", err)
		wsConn.SendJSON(domain.MessageTypeError, domain.ErrorPayload{
			Code:    "SESSION_FAILED",
			Message: "启动翻译会话失败: " + err.Error(),
		})
		wsConn.Close()
		return
	}

	// 注册活动到广播器
	h.broadcaster.RegisterActivity(authPayload.ActivityID)

	// 发送就绪状态
	wsConn.SendJSON(domain.MessageTypeState, domain.StatePayload{
		Status:  "READY",
		Message: "已连接，准备接收音频",
	})

	// 启动字幕转发 goroutine
	go h.forwardSubtitles(wsConn, session, authPayload.ActivityID)

	// 启动写入 pump
	go wsConn.WritePump()

	// 读取音频数据
	wsConn.ReadPump(func(message []byte) {
		h.handleSpeakerMessage(wsConn, session, message)
	})

	// 连接关闭，停止会话
	h.pipeline.StopSession(authPayload.ActivityID)
	h.broadcaster.UnregisterActivity(authPayload.ActivityID)
	log.Printf("Speaker disconnected: %s", connectionID)
}

// authenticate 认证
func (h *SpeakerWebSocketHandler) authenticate(conn *ws.Connection, c *gin.Context) (*domain.AuthPayload, error) {
	// 从查询参数获取认证信息
	token := c.Query("token")
	activityID := c.Query("activityId")
	language := c.Query("language")

	if token == "" || activityID == "" || language == "" {
		return nil, http.ErrAbortHandler
	}

	// TODO: 验证 JWT token

	return &domain.AuthPayload{
		Token:      token,
		ActivityID: activityID,
		Language:   language,
	}, nil
}

// handleSpeakerMessage 处理演讲者消息
func (h *SpeakerWebSocketHandler) handleSpeakerMessage(conn *ws.Connection, session *app.PipelineSession, message []byte) {
	var msg domain.WebSocketMessage
	if err := json.Unmarshal(message, &msg); err != nil {
		log.Printf("Failed to parse message: %v", err)
		return
	}

	switch msg.Type {
	case domain.MessageTypeAudio:
		// 处理音频数据
		h.handleAudio(session, msg.Payload)

	case domain.MessageTypeControl:
		// 处理控制消息
		h.handleControl(conn, msg.Payload)

	case domain.MessageTypePong:
		// 心跳响应，不需要处理
		break

	default:
		log.Printf("Unknown message type: %s", msg.Type)
	}
}

// handleAudio 处理音频数据
func (h *SpeakerWebSocketHandler) handleAudio(session *app.PipelineSession, payload interface{}) {
	audioPayload, ok := payload.(map[string]interface{})
	if !ok {
		log.Printf("Invalid audio payload")
		return
	}

	chunkStr, ok := audioPayload["chunk"].(string)
	if !ok {
		log.Printf("Invalid audio chunk")
		return
	}

	// 解码 Base64 音频数据
	audioData, err := base64.StdEncoding.DecodeString(chunkStr)
	if err != nil {
		log.Printf("Failed to decode audio: %v", err)
		return
	}

	// 发送音频到翻译管线
	if err := session.SendAudio(audioData); err != nil {
		log.Printf("Failed to send audio: %v", err)
	}
}

// handleControl 处理控制消息
func (h *SpeakerWebSocketHandler) handleControl(conn *ws.Connection, payload interface{}) {
	controlPayload, ok := payload.(map[string]interface{})
	if !ok {
		return
	}

	action, ok := controlPayload["action"].(string)
	if !ok {
		return
	}

	log.Printf("Control action: %s", action)

	// TODO: 根据 action 执行相应操作
	switch action {
	case "START":
		conn.SendJSON(domain.MessageTypeState, domain.StatePayload{
			Status:  "STREAMING",
			Message: "开始接收音频",
		})
	case "STOP":
		conn.SendJSON(domain.MessageTypeState, domain.StatePayload{
			Status:  "STOPPED",
			Message: "停止接收音频",
		})
	}
}

// forwardSubtitles 转发字幕到广播器
func (h *SpeakerWebSocketHandler) forwardSubtitles(conn *ws.Connection, session *app.PipelineSession, activityID string) {
	for subtitle := range session.SubtitleOutput {
		// 广播字幕给所有观众
		h.broadcaster.BroadcastSubtitle(activityID, subtitle)

		// 同时也发送给演讲者（显示原文和翻译）
		conn.SendJSON(domain.MessageTypeSubtitle, domain.SubtitlePayload{
			ID:         subtitle.ID,
			Original:   subtitle.Original,
			SourceLang: subtitle.SourceLang,
			Timestamp:  subtitle.Timestamp,
			Confidence: subtitle.Confidence,
		})
	}
}
