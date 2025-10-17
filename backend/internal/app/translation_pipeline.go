package app

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/hoshea/orion-backend/internal/domain"
	"github.com/hoshea/orion-backend/internal/infra/google"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// speechClient 抽象语音识别客户端
type speechClient interface {
	StreamingRecognize(ctx context.Context, audioStream <-chan []byte, config google.StreamingRecognizeConfig, results chan<- google.RecognitionResult) error
	Close() error
}

// translateClient 抽象翻译客户端
type translateClient interface {
	Translate(ctx context.Context, text, sourceLang string, targetLangs []string) ([]google.TranslationResult, error)
	Close() error
}

// TranslationPipeline 翻译管线服务，负责协调 STT 识别和 Translation 翻译
type TranslationPipeline struct {
	sttClient         speechClient
	translationClient translateClient
	mu                sync.RWMutex
	sessions          map[string]*PipelineSession // activityID -> session
}

// PipelineSession 翻译会话
type PipelineSession struct {
	ActivityID      string
	SourceLanguage  string
	TargetLanguages []string
	AudioInput      chan []byte           // 音频输入
	SubtitleOutput  chan *domain.Subtitle // 字幕输出（包含所有语言翻译）
	cancel          context.CancelFunc
	ctx             context.Context
}

const (
	// streamRestartInterval 单个流的最长持续时间，必须小于 Google 官方 5 分钟限制
	streamRestartInterval = 4*time.Minute + 30*time.Second
	// streamErrorBackoff 遇到异常时的简单退避
	streamErrorBackoff = time.Second
)

// NewTranslationPipeline 创建真实的翻译管线
func NewTranslationPipeline(ctx context.Context, sttAPIKey, translateAPIKey string) (*TranslationPipeline, error) {
	sttClient, err := google.NewSTTClient(ctx, sttAPIKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create STT client: %w", err)
	}

	translationClient, err := google.NewTranslationClient(ctx, translateAPIKey)
	if err != nil {
		_ = sttClient.Close()
		return nil, fmt.Errorf("failed to create translation client: %w", err)
	}

	return &TranslationPipeline{
		sttClient:         sttClient,
		translationClient: translationClient,
		sessions:          make(map[string]*PipelineSession),
	}, nil
}

// NewMockTranslationPipeline 创建 Mock 管线（无需真实 API Key）
func NewMockTranslationPipeline() *TranslationPipeline {
	return &TranslationPipeline{
		sttClient:         google.NewMockSTTClient(),
		translationClient: google.NewMockTranslationClient(),
		sessions:          make(map[string]*PipelineSession),
	}
}

// Close 关闭管线
func (p *TranslationPipeline) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	for _, session := range p.sessions {
		session.cancel()
		close(session.AudioInput)
		close(session.SubtitleOutput)
	}
	p.sessions = make(map[string]*PipelineSession)

	if err := p.sttClient.Close(); err != nil {
		return err
	}
	if err := p.translationClient.Close(); err != nil {
		return err
	}
	return nil
}

// StartSession 开始翻译会话
func (p *TranslationPipeline) StartSession(activityID, sourceLanguage string, targetLanguages []string) (*PipelineSession, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if _, exists := p.sessions[activityID]; exists {
		return nil, fmt.Errorf("session already exists for activity %s", activityID)
	}

	ctx, cancel := context.WithCancel(context.Background())
	session := &PipelineSession{
		ActivityID:      activityID,
		SourceLanguage:  sourceLanguage,
		TargetLanguages: targetLanguages,
		AudioInput:      make(chan []byte, 100),
		SubtitleOutput:  make(chan *domain.Subtitle, 50),
		ctx:             ctx,
		cancel:          cancel,
	}

	p.sessions[activityID] = session
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

	session.cancel()
	close(session.AudioInput)
	close(session.SubtitleOutput)
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

func (p *TranslationPipeline) processSession(session *PipelineSession) {
	sttResults := make(chan google.RecognitionResult, 50)
	go p.streamRecognitionWithRestart(session, sttResults)

	var lastFinalTranscript string
	for result := range sttResults {
		if !result.IsFinal {
			continue
		}
		if result.Transcript == "" || result.Transcript == lastFinalTranscript {
			continue
		}
		lastFinalTranscript = result.Transcript

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

		translationMap := make(map[string]string)
		for _, t := range translations {
			translationMap[t.Language] = t.Text
		}

		subtitle := &domain.Subtitle{
			ID:           uuid.New().String(),
			ActivityID:   session.ActivityID,
			Original:     result.Transcript,
			SourceLang:   session.SourceLanguage,
			Translations: translationMap,
			Confidence:   result.Confidence,
			Timestamp:    time.Now(),
		}

		select {
		case session.SubtitleOutput <- subtitle:
			log.Printf("Subtitle created for activity %s: %s", session.ActivityID, result.Transcript)
		case <-session.ctx.Done():
			return
		default:
			log.Printf("Warning: subtitle output buffer full for activity %s", session.ActivityID)
		}
	}
}

func (p *TranslationPipeline) streamRecognitionWithRestart(session *PipelineSession, results chan<- google.RecognitionResult) {
	defer close(results)

	config := google.StreamingRecognizeConfig{
		LanguageCode:               session.SourceLanguage,
		SampleRateHertz:            16000,
		EnableAutomaticPunctuation: true,
	}

	for {
		if session.ctx.Err() != nil {
			return
		}

		streamCtx, cancel := context.WithCancel(session.ctx)
		errCh := make(chan error, 1)

		go func() {
			errCh <- p.sttClient.StreamingRecognize(
				streamCtx,
				session.AudioInput,
				config,
				results,
			)
		}()

		timer := time.NewTimer(streamRestartInterval)
		var err error

		select {
		case <-session.ctx.Done():
			cancel()
			err = <-errCh
			timer.Stop()
			return
		case err = <-errCh:
			timer.Stop()
		case <-timer.C:
			cancel()
			err = <-errCh
		}

		if !timer.Stop() {
			select {
			case <-timer.C:
			default:
			}
		}
		cancel()

		switch {
		case err == nil:
			return
		case errors.Is(err, context.Canceled):
			continue
		case status.Code(err) == codes.Canceled:
			continue
		default:
			log.Printf("STT stream error for activity %s: %v", session.ActivityID, err)
			select {
			case <-time.After(streamErrorBackoff):
			case <-session.ctx.Done():
				return
			}
		}
	}
}
