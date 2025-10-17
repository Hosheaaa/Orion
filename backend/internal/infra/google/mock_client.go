package google

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// MockSTTClient 用于本地/测试环境的语音识别模拟
type MockSTTClient struct{}

// NewMockSTTClient 构造函数
func NewMockSTTClient() *MockSTTClient {
	return &MockSTTClient{}
}

// Close 实现接口
func (c *MockSTTClient) Close() error {
	return nil
}

// StreamingRecognize 模拟识别：每收到一个音频块输出一条文本
func (c *MockSTTClient) StreamingRecognize(
	ctx context.Context,
	audioStream <-chan []byte,
	config StreamingRecognizeConfig,
	results chan<- RecognitionResult,
) error {
	counter := 0
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case chunk, ok := <-audioStream:
			if !ok {
				return nil
			}
			if len(chunk) == 0 {
				continue
			}
			counter++
			text := fmt.Sprintf("模拟语音片段 %d（%s）", counter, time.Now().Format("15:04:05"))
			select {
			case results <- RecognitionResult{
				Transcript: text,
				IsFinal:    true,
				Confidence: 0.85,
			}:
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	}
}

// MockTranslationClient 用于本地/测试环境的翻译模拟
type MockTranslationClient struct{}

// NewMockTranslationClient 构造函数
func NewMockTranslationClient() *MockTranslationClient {
	return &MockTranslationClient{}
}

// Close 实现接口
func (c *MockTranslationClient) Close() error {
	return nil
}

// Translate 模拟翻译：将文本包装成「[lang] 原文」形式
func (c *MockTranslationClient) Translate(
	ctx context.Context,
	text string,
	sourceLang string,
	targetLangs []string,
) ([]TranslationResult, error) {
	results := make([]TranslationResult, 0, len(targetLangs))
	for _, lang := range targetLangs {
		if lang == "" {
			continue
		}
		if lang == sourceLang {
			results = append(results, TranslationResult{
				Language: lang,
				Text:     text,
			})
			continue
		}
		results = append(results, TranslationResult{
			Language: lang,
			Text:     fmt.Sprintf("[%s] %s", strings.ToUpper(lang), text),
		})
	}
	return results, nil
}
