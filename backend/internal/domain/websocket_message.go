package domain

import "time"

// WebSocket 消息类型定义

// MessageType 消息类型
type MessageType string

const (
	// 通用消息类型
	MessageTypeAuth    MessageType = "AUTH"    // 认证
	MessageTypePing    MessageType = "PING"    // 心跳请求
	MessageTypePong    MessageType = "PONG"    // 心跳响应
	MessageTypeState   MessageType = "STATE"   // 状态消息
	MessageTypeError   MessageType = "ERROR"   // 错误消息

	// 演讲者端消息类型
	MessageTypeAudio   MessageType = "AUDIO"   // 音频数据
	MessageTypeControl MessageType = "CONTROL" // 控制消息

	// 观众端消息类型
	MessageTypeSubtitle MessageType = "SUBTITLE" // 字幕消息
	MessageTypeHistory  MessageType = "HISTORY"  // 历史字幕
)

// WebSocketMessage WebSocket 消息基础结构
type WebSocketMessage struct {
	Type      MessageType     `json:"type"`
	Payload   interface{}     `json:"payload,omitempty"`
	Timestamp time.Time       `json:"timestamp"`
}

// AuthPayload 认证消息负载
type AuthPayload struct {
	Token      string `json:"token"`       // JWT Token
	ActivityID string `json:"activityId"`  // 活动 ID
	Language   string `json:"language"`    // 语言（演讲者：输入语种，观众：订阅语种）
}

// AudioPayload 音频消息负载
type AudioPayload struct {
	Chunk    string `json:"chunk"`    // Base64 编码的音频数据
	Sequence int    `json:"sequence"` // 序列号
}

// ControlPayload 控制消息负载
type ControlPayload struct {
	Action string `json:"action"` // 动作：START, STOP, PAUSE
}

// SubtitlePayload 字幕消息负载
type SubtitlePayload struct {
	ID         string    `json:"id"`         // 句子 ID
	Original   string    `json:"original"`   // 原文
	SourceLang string    `json:"sourceLang"` // 源语言
	TargetLang string    `json:"targetLang"` // 目标语言
	Text       string    `json:"text"`       // 翻译后的文本
	Timestamp  time.Time `json:"timestamp"`  // 时间戳
	Confidence float32   `json:"confidence"` // 置信度
}

// StatePayload 状态消息负载
type StatePayload struct {
	Status  string `json:"status"`            // 状态：READY, CONNECTED, DISCONNECTED, ERROR
	Message string `json:"message,omitempty"` // 状态描述
}

// ErrorPayload 错误消息负载
type ErrorPayload struct {
	Code    string `json:"code"`    // 错误代码
	Message string `json:"message"` // 错误描述
}

// HistoryPayload 历史字幕负载
type HistoryPayload struct {
	Subtitles []SubtitlePayload `json:"subtitles"` // 字幕列表
}
