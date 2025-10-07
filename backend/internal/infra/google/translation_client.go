package google

import (
	"context"
	"fmt"

	"cloud.google.com/go/translate"
	"golang.org/x/text/language"
	"google.golang.org/api/option"
)

// TranslationClient Google Translation API 客户端
type TranslationClient struct {
	client *translate.Client
	apiKey string
}

// NewTranslationClient 创建翻译客户端
func NewTranslationClient(ctx context.Context, apiKey string) (*TranslationClient, error) {
	client, err := translate.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("failed to create translation client: %w", err)
	}

	return &TranslationClient{
		client: client,
		apiKey: apiKey,
	}, nil
}

// Close 关闭客户端
func (c *TranslationClient) Close() error {
	return c.client.Close()
}

// TranslationResult 翻译结果
type TranslationResult struct {
	Language string // 目标语言代码
	Text     string // 翻译后的文本
}

// Translate 翻译文本到多个目标语言
func (c *TranslationClient) Translate(
	ctx context.Context,
	text string,
	sourceLang string,
	targetLangs []string,
) ([]TranslationResult, error) {
	if text == "" {
		return nil, fmt.Errorf("text cannot be empty")
	}
	if len(targetLangs) == 0 {
		return nil, fmt.Errorf("target languages cannot be empty")
	}

	results := make([]TranslationResult, 0, len(targetLangs))

	// 遍历每个目标语言进行翻译
	for _, targetLang := range targetLangs {
		// 跳过与源语言相同的目标语言
		if targetLang == sourceLang {
			results = append(results, TranslationResult{
				Language: targetLang,
				Text:     text, // 直接使用原文
			})
			continue
		}

		// 解析语言代码
		targetLangTag, err := language.Parse(targetLang)
		if err != nil {
			return nil, fmt.Errorf("invalid target language %s: %w", targetLang, err)
		}

		sourceLangTag, err := language.Parse(sourceLang)
		if err != nil {
			return nil, fmt.Errorf("invalid source language %s: %w", sourceLang, err)
		}

		// 调用翻译 API
		translations, err := c.client.Translate(ctx,
			[]string{text},
			targetLangTag,
			&translate.Options{
				Source: sourceLangTag,
			},
		)
		if err != nil {
			return nil, fmt.Errorf("failed to translate to %s: %w", targetLang, err)
		}

		if len(translations) == 0 {
			return nil, fmt.Errorf("no translation result for %s", targetLang)
		}

		results = append(results, TranslationResult{
			Language: targetLang,
			Text:     translations[0].Text,
		})
	}

	return results, nil
}

// GetSupportedLanguages 获取支持的语言列表
func (c *TranslationClient) GetSupportedLanguages(ctx context.Context) ([]string, error) {
	displayLang := language.English
	langs, err := c.client.SupportedLanguages(ctx, displayLang)
	if err != nil {
		return nil, fmt.Errorf("failed to get supported languages: %w", err)
	}

	codes := make([]string, len(langs))
	for i, lang := range langs {
		codes[i] = lang.Tag.String()
	}

	return codes, nil
}
