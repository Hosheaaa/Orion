package app

import (
	"log"
	"sync"

	"github.com/hoshea/orion-backend/internal/domain"
)

// SubtitleBroadcaster 字幕广播服务
// 负责将字幕分发给订阅了特定语言的观众
type SubtitleBroadcaster struct {
	mu          sync.RWMutex
	activities  map[string]*ActivityBroadcast // activityID -> broadcast
}

// ActivityBroadcast 单个活动的广播器
type ActivityBroadcast struct {
	ActivityID string
	mu         sync.RWMutex
	viewers    map[string]*ViewerConnection // viewerID -> connection
}

// ViewerConnection 观众连接
type ViewerConnection struct {
	ID           string
	Language     string                      // 订阅的语言
	SendChannel  chan *domain.SubtitlePayload // 发送字幕的 channel
}

// NewSubtitleBroadcaster 创建字幕广播服务
func NewSubtitleBroadcaster() *SubtitleBroadcaster {
	return &SubtitleBroadcaster{
		activities: make(map[string]*ActivityBroadcast),
	}
}

// RegisterActivity 注册活动
func (b *SubtitleBroadcaster) RegisterActivity(activityID string) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if _, exists := b.activities[activityID]; !exists {
		b.activities[activityID] = &ActivityBroadcast{
			ActivityID: activityID,
			viewers:    make(map[string]*ViewerConnection),
		}
		log.Printf("Registered activity for broadcast: %s", activityID)
	}
}

// UnregisterActivity 注销活动
func (b *SubtitleBroadcaster) UnregisterActivity(activityID string) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if broadcast, exists := b.activities[activityID]; exists {
		// 关闭所有观众连接
		broadcast.mu.Lock()
		for _, viewer := range broadcast.viewers {
			close(viewer.SendChannel)
		}
		broadcast.mu.Unlock()

		delete(b.activities, activityID)
		log.Printf("Unregistered activity from broadcast: %s", activityID)
	}
}

// AddViewer 添加观众
func (b *SubtitleBroadcaster) AddViewer(activityID, viewerID, language string) (*ViewerConnection, error) {
	b.mu.RLock()
	broadcast, exists := b.activities[activityID]
	b.mu.RUnlock()

	if !exists {
		// 自动注册活动
		b.RegisterActivity(activityID)
		b.mu.RLock()
		broadcast = b.activities[activityID]
		b.mu.RUnlock()
	}

	viewer := &ViewerConnection{
		ID:          viewerID,
		Language:    language,
		SendChannel: make(chan *domain.SubtitlePayload, 100), // 缓冲 100 条字幕
	}

	broadcast.mu.Lock()
	broadcast.viewers[viewerID] = viewer
	broadcast.mu.Unlock()

	log.Printf("Added viewer %s to activity %s (language: %s)", viewerID, activityID, language)
	return viewer, nil
}

// RemoveViewer 移除观众
func (b *SubtitleBroadcaster) RemoveViewer(activityID, viewerID string) {
	b.mu.RLock()
	broadcast, exists := b.activities[activityID]
	b.mu.RUnlock()

	if !exists {
		return
	}

	broadcast.mu.Lock()
	if viewer, found := broadcast.viewers[viewerID]; found {
		close(viewer.SendChannel)
		delete(broadcast.viewers, viewerID)
		log.Printf("Removed viewer %s from activity %s", viewerID, activityID)
	}
	broadcast.mu.Unlock()
}

// BroadcastSubtitle 广播字幕
// 根据观众订阅的语言分发字幕
func (b *SubtitleBroadcaster) BroadcastSubtitle(activityID string, subtitle *domain.Subtitle) {
	b.mu.RLock()
	broadcast, exists := b.activities[activityID]
	b.mu.RUnlock()

	if !exists {
		log.Printf("Warning: No broadcast found for activity %s", activityID)
		return
	}

	broadcast.mu.RLock()
	defer broadcast.mu.RUnlock()

	// 遍历所有观众，发送对应语言的字幕
	for _, viewer := range broadcast.viewers {
		// 获取观众订阅语言的翻译
		translatedText, ok := subtitle.Translations[viewer.Language]
		if !ok {
			// 如果没有该语言的翻译，跳过
			continue
		}

		subtitlePayload := &domain.SubtitlePayload{
			ID:         subtitle.ID,
			Original:   subtitle.Original,
			SourceLang: subtitle.SourceLang,
			TargetLang: viewer.Language,
			Text:       translatedText,
			Timestamp:  subtitle.Timestamp,
			Confidence: subtitle.Confidence,
		}

		// 非阻塞发送
		select {
		case viewer.SendChannel <- subtitlePayload:
			// 发送成功
		default:
			// Channel 已满，跳过该观众（避免阻塞其他观众）
			log.Printf("Warning: viewer %s channel is full, skipping subtitle", viewer.ID)
		}
	}

	log.Printf("Broadcasted subtitle for activity %s to %d viewers", activityID, len(broadcast.viewers))
}

// GetViewerCount 获取活动的观众数量
func (b *SubtitleBroadcaster) GetViewerCount(activityID string) int {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if broadcast, exists := b.activities[activityID]; exists {
		broadcast.mu.RLock()
		defer broadcast.mu.RUnlock()
		return len(broadcast.viewers)
	}

	return 0
}

// GetViewersByLanguage 按语言统计观众数量
func (b *SubtitleBroadcaster) GetViewersByLanguage(activityID string) map[string]int {
	b.mu.RLock()
	defer b.mu.RUnlock()

	result := make(map[string]int)

	if broadcast, exists := b.activities[activityID]; exists {
		broadcast.mu.RLock()
		defer broadcast.mu.RUnlock()

		for _, viewer := range broadcast.viewers {
			result[viewer.Language]++
		}
	}

	return result
}
