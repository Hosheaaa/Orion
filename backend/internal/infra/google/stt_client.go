package google

import (
	"context"
	"fmt"
	"io"

	speech "cloud.google.com/go/speech/apiv1"
	"cloud.google.com/go/speech/apiv1/speechpb"
	"google.golang.org/api/option"
)

// STTClient Google Speech-to-Text 客户端
type STTClient struct {
	client *speech.Client
	apiKey string
}

// NewSTTClient 创建 STT 客户端
func NewSTTClient(ctx context.Context, apiKey string) (*STTClient, error) {
	client, err := speech.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("failed to create STT client: %w", err)
	}

	return &STTClient{
		client: client,
		apiKey: apiKey,
	}, nil
}

// Close 关闭客户端
func (c *STTClient) Close() error {
	return c.client.Close()
}

// StreamingRecognizeConfig 流式识别配置
type StreamingRecognizeConfig struct {
	LanguageCode       string // 例如 "zh-CN", "en-US"
	SampleRateHertz    int32  // 采样率，例如 16000
	EnableAutomaticPunctuation bool   // 是否启用自动标点
}

// RecognitionResult 识别结果
type RecognitionResult struct {
	Transcript string  // 识别的文本
	IsFinal    bool    // 是否是最终结果
	Confidence float32 // 置信度 (0-1)
}

// StreamingRecognize 流式语音识别
// audioStream: 音频数据流 channel
// config: 识别配置
// results: 输出识别结果的 channel
func (c *STTClient) StreamingRecognize(
	ctx context.Context,
	audioStream <-chan []byte,
	config StreamingRecognizeConfig,
	results chan<- RecognitionResult,
) error {
	// 创建流式识别客户端
	stream, err := c.client.StreamingRecognize(ctx)
	if err != nil {
		return fmt.Errorf("failed to create streaming recognize: %w", err)
	}

	// 发送配置
	if err := stream.Send(&speechpb.StreamingRecognizeRequest{
		StreamingRequest: &speechpb.StreamingRecognizeRequest_StreamingConfig{
			StreamingConfig: &speechpb.StreamingRecognitionConfig{
				Config: &speechpb.RecognitionConfig{
					Encoding:                   speechpb.RecognitionConfig_LINEAR16,
					SampleRateHertz:           config.SampleRateHertz,
					LanguageCode:              config.LanguageCode,
					EnableAutomaticPunctuation: config.EnableAutomaticPunctuation,
				},
				InterimResults: true, // 启用中间结果
			},
		},
	}); err != nil {
		return fmt.Errorf("failed to send config: %w", err)
	}

	// 启动音频发送 goroutine
	go func() {
		defer func() {
			if err := stream.CloseSend(); err != nil {
				// 记录错误但不阻塞
			}
		}()

		for {
			select {
			case <-ctx.Done():
				return
			case audioData, ok := <-audioStream:
				if !ok {
					// 音频流关闭
					return
				}

				// 发送音频数据
				if err := stream.Send(&speechpb.StreamingRecognizeRequest{
					StreamingRequest: &speechpb.StreamingRecognizeRequest_AudioContent{
						AudioContent: audioData,
					},
				}); err != nil {
					// 发送失败，退出
					return
				}
			}
		}
	}()

	// 接收识别结果
	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			// 流正常结束
			break
		}
		if err != nil {
			return fmt.Errorf("failed to receive response: %w", err)
		}

		// 处理识别结果
		for _, result := range resp.Results {
			if len(result.Alternatives) > 0 {
				alt := result.Alternatives[0]

				recognitionResult := RecognitionResult{
					Transcript: alt.Transcript,
					IsFinal:    result.IsFinal,
					Confidence: alt.Confidence,
				}

				select {
				case results <- recognitionResult:
				case <-ctx.Done():
					return ctx.Err()
				}
			}
		}
	}

	return nil
}
