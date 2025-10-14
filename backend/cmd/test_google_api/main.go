package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/hoshea/orion-backend/internal/infra/google"
)

func main() {
	ctx := context.Background()

	// 从环境变量获取 API Key
	sttAPIKey := os.Getenv("GOOGLE_STT_API_KEY")
	translateAPIKey := os.Getenv("GOOGLE_TRANSLATE_API_KEY")

	if sttAPIKey == "" || translateAPIKey == "" {
		log.Fatal("请设置 GOOGLE_STT_API_KEY 和 GOOGLE_TRANSLATE_API_KEY 环境变量")
	}

	fmt.Println("=== Google API 测试 ===")
	fmt.Printf("STT API Key: %s...%s\n", sttAPIKey[:10], sttAPIKey[len(sttAPIKey)-10:])
	fmt.Printf("Translation API Key: %s...%s\n\n", translateAPIKey[:10], translateAPIKey[len(translateAPIKey)-10:])

	// 测试翻译 API
	fmt.Println("=== 测试 Google Translation API ===")
	testTranslationAPI(ctx, translateAPIKey)

	// 测试 STT API
	fmt.Println("\n=== 测试 Google Speech-to-Text API ===")
	testSTTAPI(ctx, sttAPIKey)

	fmt.Println("\n=== 测试完成 ===")
	fmt.Println("\n提示：")
	fmt.Println("1. 如果遇到 403 错误，请确认已在 Google Cloud Console 中启用 API")
	fmt.Println("2. API 启用后可能需要等待 5-10 分钟才能生效")
	fmt.Println("3. 确认 API Key 有正确的权限和配额")
	fmt.Println("4. 如果是新项目，可能需要启用计费")
}

func testSTTAPI(ctx context.Context, apiKey string) {
	client, err := google.NewSTTClient(ctx, apiKey)
	if err != nil {
		log.Fatalf("创建 STT 客户端失败: %v", err)
	}
	defer client.Close()

	audioStream := make(chan []byte, 1)
	results := make(chan google.RecognitionResult, 10)

	// 构造 0.5 秒的静音音频 (16kHz, 16bit LINEAR16)
	silentSamples := make([]byte, 16000) // 8000 采样点 * 2 字节
	audioStream <- silentSamples
	close(audioStream)

	err = client.StreamingRecognize(
		ctx,
		audioStream,
		google.StreamingRecognizeConfig{
			LanguageCode:               "en-US",
			SampleRateHertz:            16000,
			EnableAutomaticPunctuation: true,
		},
		results,
	)

	close(results)

	if err != nil {
		log.Printf("语音识别调用失败: %v", err)
		log.Println("可能原因：API 未启用、凭据权限不足、或音频数据格式不符合要求。")
		return
	}

	var count int
	for res := range results {
		count++
		fmt.Printf("识别结果：%s (final=%v, confidence=%f)\n", res.Transcript, res.IsFinal, res.Confidence)
	}

	if count == 0 {
		fmt.Println("✅ STT 客户端调用完成（静音样本未产生识别结果，但连接成功）")
	}
}

func testTranslationAPI(ctx context.Context, apiKey string) {
	client, err := google.NewTranslationClient(ctx, apiKey)
	if err != nil {
		log.Fatalf("创建翻译客户端失败: %v", err)
	}
	defer client.Close()

	// 测试简单翻译
	text := "Hello"
	sourceLang := "en"
	targetLangs := []string{"zh-CN"}

	fmt.Printf("原文: %s\n", text)
	fmt.Printf("源语言: %s\n", sourceLang)
	fmt.Printf("目标语言: %v\n\n", targetLangs)
	fmt.Println("开始翻译...")

	results, err := client.Translate(ctx, text, sourceLang, targetLangs)
	if err != nil {
		log.Printf("翻译失败: %v\n", err)
		log.Println("\n可能的原因：")
		log.Println("- Cloud Translation API 未启用或刚启用（等待几分钟）")
		log.Println("- API Key 权限不足")
		log.Println("- 项目未启用计费")
		log.Printf("- 访问此链接启用 API: https://console.developers.google.com/apis/api/translate.googleapis.com/overview\n")
		return
	}

	fmt.Println("✅ 翻译成功！")
	fmt.Println("翻译结果:")
	for _, result := range results {
		fmt.Printf("  [%s] %s\n", result.Language, result.Text)
	}

	// 测试中文翻译
	fmt.Println("\n测试中文翻译:")
	text2 := "大家好，欢迎来到2024产品发布会"
	sourceLang2 := "zh-CN"
	targetLangs2 := []string{"en", "ja"}

	fmt.Printf("原文: %s\n", text2)
	results2, err := client.Translate(ctx, text2, sourceLang2, targetLangs2)
	if err != nil {
		log.Printf("翻译失败: %v\n", err)
		return
	}

	fmt.Println("翻译结果:")
	for _, result := range results2 {
		fmt.Printf("  [%s] %s\n", result.Language, result.Text)
	}
}
