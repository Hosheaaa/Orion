package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config 应用配置
type Config struct {
	Server        ServerConfig
	Auth          AuthConfig
	Google        GoogleConfig
	Redis         RedisConfig
	Cache         CacheConfig
	ViewerBaseURL string
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Port           int
	Env            string
	AllowedOrigins []string
}

// AuthConfig 认证配置
type AuthConfig struct {
	JWTSecretPath   string
	AdminUsername   string
	AdminPassword   string
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
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
	HistoryTTL     string
	WSPingInterval string
}

// Load 加载配置（从环境变量）
func Load() (*Config, error) {
	port := getEnvAsInt("APP_PORT", 8080)
	env := getEnv("APP_ENV", "development")
	allowedOrigins := getEnvAsStringSlice("CORS_ALLOWED_ORIGINS", []string{"http://localhost:3000"})

	return &Config{
		Server: ServerConfig{
			Port:           port,
			Env:            env,
			AllowedOrigins: allowedOrigins,
		},
		Auth: AuthConfig{
			JWTSecretPath:   getEnv("JWT_SECRET_PATH", "./secrets/jwt_private.pem"),
			AdminUsername:   getEnv("ADMIN_USERNAME", "admin"),
			AdminPassword:   getEnv("ADMIN_PASSWORD", "admin123"),
			AccessTokenTTL:  getEnvAsDuration("ACCESS_TOKEN_TTL", 15*time.Minute),
			RefreshTokenTTL: getEnvAsDuration("REFRESH_TOKEN_TTL", 7*24*time.Hour),
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

// 工具函数：获取字符串切片环境变量（逗号分隔）
func getEnvAsStringSlice(key string, defaultValue []string) []string {
	valueStr := getEnv(key, "")
	if valueStr == "" {
		return defaultValue
	}
	parts := strings.Split(valueStr, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	if len(result) == 0 {
		return defaultValue
	}
	return result
}

// 工具函数：获取整型环境变量
func getEnvAsInt(key string, defaultValue int) int {
	valueStr := getEnv(key, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultValue
}

// 工具函数：获取 Duration 环境变量
func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	valueStr := getEnv(key, "")
	if valueStr == "" {
		return defaultValue
	}
	if d, err := time.ParseDuration(valueStr); err == nil {
		return d
	}
	return defaultValue
}

// Validate 验证配置
func (c *Config) Validate() error {
	if c.Server.Port <= 0 || c.Server.Port > 65535 {
		return fmt.Errorf("invalid port: %d", c.Server.Port)
	}
	if c.Auth.AdminUsername == "" || c.Auth.AdminPassword == "" {
		return fmt.Errorf("admin credentials must be configured")
	}
	if c.Auth.AccessTokenTTL <= 0 {
		return fmt.Errorf("ACCESS_TOKEN_TTL 必须大于 0")
	}
	if c.Auth.RefreshTokenTTL <= 0 {
		return fmt.Errorf("REFRESH_TOKEN_TTL 必须大于 0")
	}
	return nil
}
