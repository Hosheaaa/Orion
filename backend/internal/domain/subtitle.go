package domain

import "time"

// Subtitle 字幕实体
type Subtitle struct {
	ID           string            `json:"id"`           // 句子 ID
	ActivityID   string            `json:"activityId"`   // 活动 ID
	Original     string            `json:"original"`     // 原文
	SourceLang   string            `json:"sourceLang"`   // 源语言
	Translations map[string]string `json:"translations"` // 翻译结果 {语言代码: 翻译文本}
	Confidence   float32           `json:"confidence"`   // 置信度
	Timestamp    time.Time         `json:"timestamp"`    // 时间戳
}

// SubtitleForLanguage 特定语言的字幕
type SubtitleForLanguage struct {
	ID         string    `json:"id"`
	Original   string    `json:"original"`
	SourceLang string    `json:"sourceLang"`
	TargetLang string    `json:"targetLang"`
	Text       string    `json:"text"`
	Timestamp  time.Time `json:"timestamp"`
}
