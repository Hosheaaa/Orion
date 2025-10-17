package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// SpeakerConsoleHandler 提供演讲者控制台的辅助数据
type SpeakerConsoleHandler struct{}

// NewSpeakerConsoleHandler 构造函数
func NewSpeakerConsoleHandler() *SpeakerConsoleHandler {
	return &SpeakerConsoleHandler{}
}

// HeroInsight 控制台关键指标
type HeroInsight struct {
	Label       string `json:"label"`
	Value       string `json:"value"`
	Trend       string `json:"trend"`
	DeltaText   string `json:"deltaText"`
	Description string `json:"description"`
	Accent      string `json:"accent"`
}

// GuidanceChecklistItem 控制台引导清单
type GuidanceChecklistItem struct {
	Title    string `json:"title"`
	Detail   string `json:"detail"`
	Emphasis string `json:"emphasis"`
}

// SubtitleHistoryItem 控制台字幕历史
type SubtitleHistoryItem struct {
	ID         string `json:"id"`
	Original   string `json:"original"`
	Translated string `json:"translated"`
	Timestamp  string `json:"timestamp"`
}

// GetHeroInsights 返回演讲者端关键指标
func (h *SpeakerConsoleHandler) GetHeroInsights(c *gin.Context) {
	items := []HeroInsight{
		{
			Label:       "实时观众",
			Value:       "128",
			Trend:       "up",
			DeltaText:   "+5.1% 较上场",
			Description: "包含桌面端与移动端，移动端占比 71%。",
			Accent:      "#10b981",
		},
		{
			Label:       "翻译延迟",
			Value:       "1.42s",
			Trend:       "down",
			DeltaText:   "-0.3s",
			Description: "端到端延迟稳定控制在 1.5 秒内。",
			Accent:      "#6366f1",
		},
		{
			Label:       "字幕准确率",
			Value:       "97.8%",
			Trend:       "stable",
			DeltaText:   "稳定在 ≥97%",
			Description: "术语表启用，关键名词识别率 >99%。",
			Accent:      "#f97316",
		},
	}
	c.JSON(http.StatusOK, items)
}

// GetSubtitleHistory 返回最近字幕历史数据
func (h *SpeakerConsoleHandler) GetSubtitleHistory(c *gin.Context) {
	now := time.Now()
	items := []SubtitleHistoryItem{
		{
			ID:         "sub-001",
			Original:   "大家下午好，欢迎来到 Orion 实时翻译发布会。",
			Translated: "Good afternoon, welcome to the Orion real-time translation launch.",
			Timestamp:  now.Add(-25 * time.Second).UTC().Format(time.RFC3339),
		},
		{
			ID:         "sub-002",
			Original:   "今天我们将展示最新的流式翻译引擎和运营方案。",
			Translated: "Today we present our latest streaming translation engine and operation plan.",
			Timestamp:  now.Add(-18 * time.Second).UTC().Format(time.RFC3339),
		},
		{
			ID:         "sub-003",
			Original:   "它支持 30+ 目标语种，并提供术语表管理能力。",
			Translated: "It supports over 30 target languages with glossary management capabilities.",
			Timestamp:  now.Add(-12 * time.Second).UTC().Format(time.RFC3339),
		},
	}
	c.JSON(http.StatusOK, items)
}

// GetGuidanceChecklist 返回演讲引导建议
func (h *SpeakerConsoleHandler) GetGuidanceChecklist(c *gin.Context) {
	items := []GuidanceChecklistItem{
		{
			Title:    "语速控制在 150 字/分钟以内",
			Detail:   "确保识别准确率和字幕滚动体验。",
			Emphasis: "primary",
		},
		{
			Title:    "段落之间留出 1.5 秒停顿",
			Detail:   "便于翻译引擎分段处理并降低延迟。",
			Emphasis: "success",
		},
		{
			Title:    "遇到专有名词先拼写后阐述",
			Detail:   "帮助术语库准确抓取并同步给译员。",
			Emphasis: "warning",
		},
	}
	c.JSON(http.StatusOK, items)
}
