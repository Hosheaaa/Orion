package app

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/hoshea/orion-backend/internal/domain"
	"github.com/hoshea/orion-backend/internal/infra/google"
)

// TranslationPipeline 翻译管线服务
// 负责协调 STT 识别和 Translation 翻译
type TranslationPipeline struct {
	sttClient         *google.STTClient
	translationClient *google.TranslationClient
	mu                sync.RWMutex
	sessions          map[string]*PipelineSession // activityID -> session
}

// PipelineSession 翻译会话
type PipelineSession struct {
	ActivityID      string
	SourceLanguage  string
	TargetLanguages []string
	AudioInput      chan []byte                       // 音频输入
	SubtitleOutput  chan *domain.Subtitle             // 字幕输出（包含所有语言翻译）
	cancel          context.CancelFunc
	ctx             context.Context
}

// NewTranslationPipeline 创建翻译管线
func NewTranslationPipeline(ctx context.Context, sttAPIKey, translateAPIKey string) (*TranslationPipeline, error) {
	// 创建 STT 客户端
	sttClient, err := google.NewSTTClient(ctx, sttAPIKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create STT client: %w", err)
	}

	// 创建翻译客户端
	translationClient, err := google.NewTranslationClient(ctx, translateAPIKey)
	if err != nil {
		sttClient.Close()
		return nil, fmt.Errorf("failed to create translation client: %w", err)
	}

	return &TranslationPipeline{
		sttClient:         sttClient,
		translationClient: translationClient,
		sessions:          make(map[string]*PipelineSession),
	}, nil
}

// Close 关闭管线
func (p *TranslationPipeline) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	// 停止所有会话
	for _, session := range p.sessions {
		session.cancel()
		close(session.AudioInput)
		close(session.SubtitleOutput)
	}

	// 关闭客户端
	if err := p.sttClient.Close(); err != nil {
		return err
	}
	if err := p.translationClient.Close(); err != nil {
		return err
	}

	return nil
}

// StartSession 开始翻译会话
func (p *TranslationPipeline) StartSession(
	activityID string,
	sourceLanguage string,
	targetLanguages []string,
) (*PipelineSession, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	// 检查会话是否已存在
	if _, exists := p.sessions[activityID]; exists {
		return nil, fmt.Errorf("session already exists for activity %s", activityID)
	}

	// 创建会话上下文
	ctx, cancel := context.WithCancel(context.Background())

	session := &PipelineSession{
		ActivityID:      activityID,
		SourceLanguage:  sourceLanguage,
		TargetLanguages: targetLanguages,
		AudioInput:      make(chan []byte, 100),      // 缓冲音频数据
		SubtitleOutput:  make(chan *domain.Subtitle, 50), // 缓冲字幕
		ctx:             ctx,
		cancel:          cancel,
	}

	p.sessions[activityID] = session

	// 启动处理 goroutine
	go p.processSession(session)

	log.Printf("Started translation session for activity %s", activityID)
	return session, nil
}

// StopSession 停止翻译会话
func (p *TranslationPipeline) StopSession(activityID string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	session, exists := p.sessions[activityID]
	if !exists {
		return fmt.Errorf("session not found for activity %s", activityID)
	}

	// 取消上下文
	session.cancel()

	// 关闭 channels
	close(session.AudioInput)
	close(session.SubtitleOutput)

	// 删除会话
	delete(p.sessions, activityID)

	log.Printf("Stopped translation session for activity %s", activityID)
	return nil
}

// GetSession 获取会话
func (p *TranslationPipeline) GetSession(activityID string) (*PipelineSession, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	session, exists := p.sessions[activityID]
	if !exists {
		return nil, fmt.Errorf("session not found for activity %s", activityID)
	}

	return session, nil
}

// processSession 处理会话的核心逻辑
func (p *TranslationPipeline) processSession(session *PipelineSession) {
	// 创建 STT 结果 channel
	sttResults := make(chan google.RecognitionResult, 50)

	// 启动 STT 识别
	go func() {
		defer close(sttResults)

		config := google.StreamingRecognizeConfig{
			LanguageCode:               session.SourceLanguage,
			SampleRateHertz:            16000,
			EnableAutomaticPunctuation: true,
		}

		if err := p.sttClient.StreamingRecognize(
			session.ctx,
			session.AudioInput,
			config,
			sttResults,
		); err != nil {
			log.Printf("STT error for activity %s: %v", session.ActivityID, err)
		}
	}()

	// 处理 STT 结果并翻译
	for result := range sttResults {
		// 只处理最终结果
		if !result.IsFinal {
			continue
		}

		// 翻译到目标语言
		translations, err := p.translationClient.Translate(
			session.ctx,
			result.Transcript,
			session.SourceLanguage,
			session.TargetLanguages,
		)
		if err != nil {
			log.Printf("Translation error for activity %s: %v", session.ActivityID, err)
			continue
		}

		// 构建翻译映射
		translationMap := make(map[string]string)
		for _, t := range translations {
			translationMap[t.Language] = t.Text
		}

		// 创建字幕
		subtitle := &domain.Subtitle{
			ID:           uuid.New().String(),
			ActivityID:   session.ActivityID,
			Original:     result.Transcript,
			SourceLang:   session.SourceLanguage,
			Translations: translationMap,
			Confidence:   result.Confidence,
			Timestamp:    time.Now(),
		}

		// 发送字幕到输出 channel
		select {
		case session.SubtitleOutput <- subtitle:
			log.Printf("Subtitle created for activity %s: %s", session.ActivityID, result.Transcript)
		case <-session.ctx.Done():
			return
		default:
			// 输出 channel 已满，跳过
			log.Printf("Warning: subtitle output buffer full for activity %s", session.ActivityID)
		}
	}
}

// SendAudio 发送音频数据到会话
func (s *PipelineSession) SendAudio(audioData []byte) error {
	select {
	case s.AudioInput <- audioData:
		return nil
	case <-s.ctx.Done():
		return fmt.Errorf("session closed")
	default:
		return fmt.Errorf("audio buffer full")
	}
}
