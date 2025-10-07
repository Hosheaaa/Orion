package config

import (
	"fmt"
	"os"
	"strconv"
)

// Config 应用配置
type Config struct {
	Server   ServerConfig
	Auth     AuthConfig
	Google   GoogleConfig
	Redis    RedisConfig
	Cache    CacheConfig
	ViewerBaseURL string
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Port int
	Env  string
}

// AuthConfig 认证配置
type AuthConfig struct {
	JWTSecretPath string
}

// GoogleConfig Google API 配置
type GoogleConfig struct {
	CredentialsPath string
	ProjectID       string
	STTAPIKey       string // Speech-to-Text API Key
	TranslateAPIKey string // Translation API Key
}

// RedisConfig Redis 配置
type RedisConfig struct {
	URL string
}

// CacheConfig 缓存配置
type CacheConfig struct {
	HistoryTTL      string
	WSPingInterval  string
}

// Load 加载配置（从环境变量）
func Load() (*Config, error) {
	port := getEnvAsInt("APP_PORT", 8080)
	env := getEnv("APP_ENV", "development")

	return &Config{
		Server: ServerConfig{
			Port: port,
			Env:  env,
		},
		Auth: AuthConfig{
			JWTSecretPath: getEnv("JWT_SECRET_PATH", "./secrets/jwt_private.pem"),
		},
		Google: GoogleConfig{
			CredentialsPath: getEnv("GOOGLE_APPLICATION_CREDENTIALS", "./secrets/google-service-account.json"),
			ProjectID:       getEnv("GOOGLE_PROJECT_ID", ""),
			STTAPIKey:       getEnv("GOOGLE_STT_API_KEY", ""),
			TranslateAPIKey: getEnv("GOOGLE_TRANSLATE_API_KEY", ""),
		},
		Redis: RedisConfig{
			URL: getEnv("REDIS_URL", "redis://localhost:6379/0"),
		},
		Cache: CacheConfig{
			HistoryTTL:     getEnv("HISTORY_CACHE_TTL", "5m"),
			WSPingInterval: getEnv("WS_PING_INTERVAL", "30s"),
		},
		ViewerBaseURL: getEnv("VIEWER_BASE_URL", "http://localhost:3000"),
	}, nil
}

// 工具函数：获取环境变量
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// 工具函数：获取整型环境变量
func getEnvAsInt(key string, defaultValue int) int {
	valueStr := getEnv(key, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultValue
}

// Validate 验证配置
func (c *Config) Validate() error {
	if c.Server.Port <= 0 || c.Server.Port > 65535 {
		return fmt.Errorf("invalid port: %d", c.Server.Port)
	}
	return nil
}
