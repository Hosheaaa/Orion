package handler

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/hoshea/orion-backend/internal/app"
	"github.com/hoshea/orion-backend/internal/domain"
	ws "github.com/hoshea/orion-backend/internal/infra/websocket"
)

// ViewerWebSocketHandler 观众 WebSocket 处理器
type ViewerWebSocketHandler struct {
	broadcaster *app.SubtitleBroadcaster
}

// NewViewerWebSocketHandler 创建观众处理器
func NewViewerWebSocketHandler(broadcaster *app.SubtitleBroadcaster) *ViewerWebSocketHandler {
	return &ViewerWebSocketHandler{
		broadcaster: broadcaster,
	}
}

// HandleViewerWebSocket 处理观众 WebSocket 连接
func (h *ViewerWebSocketHandler) HandleViewerWebSocket(c *gin.Context) {
	// 升级为 WebSocket 连接
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("Failed to upgrade connection: %v", err)
		return
	}

	// 创建连接封装
	viewerID := uuid.New().String()
	wsConn := ws.NewConnection(viewerID, conn)

	log.Printf("Viewer WebSocket connected: %s", viewerID)

	// 等待认证消息
	authPayload, err := h.authenticateViewer(wsConn, c)
	if err != nil {
		log.Printf("Authentication failed: %v", err)
		wsConn.SendJSON(domain.MessageTypeError, domain.ErrorPayload{
			Code:    "AUTH_FAILED",
			Message: "认证失败: " + err.Error(),
		})
		wsConn.Close()
		return
	}

	// 添加观众到广播器
	viewerConn, err := h.broadcaster.AddViewer(authPayload.ActivityID, viewerID, authPayload.Language)
	if err != nil {
		log.Printf("Failed to add viewer: %v", err)
		wsConn.SendJSON(domain.MessageTypeError, domain.ErrorPayload{
			Code:    "ADD_VIEWER_FAILED",
			Message: "添加观众失败: " + err.Error(),
		})
		wsConn.Close()
		return
	}

	// 发送连接成功消息
	wsConn.SendJSON(domain.MessageTypeState, domain.StatePayload{
		Status:  "CONNECTED",
		Message: "已连接，准备接收字幕",
	})

	// TODO: 发送历史字幕
	// h.sendHistory(wsConn, authPayload.ActivityID, authPayload.Language)

	// 启动字幕转发 goroutine
	go h.forwardSubtitlesToViewer(wsConn, viewerConn)

	// 启动写入 pump
	go wsConn.WritePump()

	// 读取客户端消息（心跳等）
	wsConn.ReadPump(func(message []byte) {
		h.handleViewerMessage(wsConn, message)
	})

	// 连接关闭，移除观众
	h.broadcaster.RemoveViewer(authPayload.ActivityID, viewerID)
	log.Printf("Viewer disconnected: %s", viewerID)
}

// authenticateViewer 认证观众
func (h *ViewerWebSocketHandler) authenticateViewer(conn *ws.Connection, c *gin.Context) (*domain.AuthPayload, error) {
	// 从查询参数获取认证信息
	token := c.Query("token")
	activityID := c.Query("activityId")
	language := c.Query("language")

	if token == "" || activityID == "" || language == "" {
		return nil, http.ErrAbortHandler
	}

	// TODO: 验证 JWT token 和活动状态

	return &domain.AuthPayload{
		Token:      token,
		ActivityID: activityID,
		Language:   language,
	}, nil
}

// handleViewerMessage 处理观众消息
func (h *ViewerWebSocketHandler) handleViewerMessage(conn *ws.Connection, message []byte) {
	var msg domain.WebSocketMessage
	if err := json.Unmarshal(message, &msg); err != nil {
		log.Printf("Failed to parse message: %v", err)
		return
	}

	switch msg.Type {
	case domain.MessageTypePong:
		// 心跳响应，不需要处理
		break

	default:
		log.Printf("Unknown viewer message type: %s", msg.Type)
	}
}

// forwardSubtitlesToViewer 转发字幕给观众
func (h *ViewerWebSocketHandler) forwardSubtitlesToViewer(conn *ws.Connection, viewerConn *app.ViewerConnection) {
	for subtitle := range viewerConn.SendChannel {
		if conn.IsClosed() {
			return
		}

		// 发送字幕给观众
		if err := conn.SendJSON(domain.MessageTypeSubtitle, subtitle); err != nil {
			log.Printf("Failed to send subtitle to viewer: %v", err)
			return
		}
	}
}

// sendHistory 发送历史字幕（TODO）
func (h *ViewerWebSocketHandler) sendHistory(conn *ws.Connection, activityID, language string) {
	// TODO: 从缓存中获取历史字幕
	historyPayload := domain.HistoryPayload{
		Subtitles: []domain.SubtitlePayload{},
	}

	conn.SendJSON(domain.MessageTypeHistory, historyPayload)
}
