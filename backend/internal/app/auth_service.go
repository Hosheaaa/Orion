package app

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/hoshea/orion-backend/internal/infra/config"
)

// AuthService 负责管理员登录、令牌签发与刷新
type AuthService struct {
	adminUsername string
	adminPassword string
	signingKey    []byte
	accessTTL     time.Duration
	refreshTTL    time.Duration
	mu            sync.RWMutex
	refreshTokens map[string]*RefreshSession
}

// RefreshSession 刷新令牌会话信息
type RefreshSession struct {
	Token     string
	UserID    string
	Role      string
	ExpiresAt time.Time
}

// AuthTokens 登录或刷新后返回的令牌对
type AuthTokens struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	ExpiresIn    int64  `json:"expiresIn"`
}

// Claims JWT 声明
type Claims struct {
	UserID    string `json:"uid"`
	Role      string `json:"role"`
	Subject   string `json:"sub"`
	IssuedAt  int64  `json:"iat"`
	ExpiresAt int64  `json:"exp"`
	NotBefore int64  `json:"nbf"`
}

const defaultRoleAdmin = "admin"

// NewAuthService 创建 AuthService
func NewAuthService(cfg *config.Config) (*AuthService, error) {
	if cfg == nil {
		return nil, errors.New("config is required")
	}

	if cfg.Auth.AdminUsername == "" || cfg.Auth.AdminPassword == "" {
		return nil, errors.New("admin username/password must be configured")
	}

	signingKey, err := loadSigningKey(cfg.Auth.JWTSecretPath)
	if err != nil {
		return nil, err
	}

	return &AuthService{
		adminUsername: cfg.Auth.AdminUsername,
		adminPassword: cfg.Auth.AdminPassword,
		signingKey:    signingKey,
		accessTTL:     cfg.Auth.AccessTokenTTL,
		refreshTTL:    cfg.Auth.RefreshTokenTTL,
		refreshTokens: make(map[string]*RefreshSession),
	}, nil
}

// Authenticate 校验管理员账号并签发令牌
func (s *AuthService) Authenticate(username, password string) (*AuthTokens, error) {
	if username != s.adminUsername || password != s.adminPassword {
		return nil, errors.New("用户名或密码错误")
	}

	return s.issueTokens(s.adminUsername, defaultRoleAdmin)
}

// Refresh 使用 refresh token 刷新访问令牌
func (s *AuthService) Refresh(refreshToken string) (*AuthTokens, error) {
	if refreshToken == "" {
		return nil, errors.New("刷新令牌不能为空")
	}

	s.mu.RLock()
	session, ok := s.refreshTokens[refreshToken]
	s.mu.RUnlock()
	if !ok {
		return nil, errors.New("刷新令牌无效")
	}

	if time.Now().After(session.ExpiresAt) {
		s.DeleteRefreshToken(refreshToken)
		return nil, errors.New("刷新令牌已过期")
	}

	// 刷新时旋转刷新令牌
	s.DeleteRefreshToken(refreshToken)
	return s.issueTokens(session.UserID, session.Role)
}

// ValidateAccessToken 解析并验证访问令牌
func (s *AuthService) ValidateAccessToken(tokenString string) (*Claims, error) {
	if tokenString == "" {
		return nil, errors.New("缺少访问令牌")
	}

	parts := strings.Split(tokenString, ".")
	if len(parts) != 3 {
		return nil, errors.New("访问令牌格式错误")
	}

	unsigned := parts[0] + "." + parts[1]
	expectedSig := signToken(unsigned, s.signingKey)
	if !hmac.Equal([]byte(expectedSig), []byte(parts[2])) {
		return nil, errors.New("访问令牌签名无效")
	}

	payloadBytes, err := base64URLDecode(parts[1])
	if err != nil {
		return nil, errors.New("访问令牌载荷无法解析")
	}

	var claims Claims
	if err := json.Unmarshal(payloadBytes, &claims); err != nil {
		return nil, errors.New("访问令牌载荷格式错误")
	}

	if claims.NotBefore != 0 && time.Now().Unix() < claims.NotBefore {
		return nil, errors.New("访问令牌尚未生效")
	}
	if claims.ExpiresAt != 0 && time.Now().Unix() > claims.ExpiresAt {
		return nil, errors.New("访问令牌已过期")
	}

	return &claims, nil
}

// DeleteRefreshToken 移除刷新令牌
func (s *AuthService) DeleteRefreshToken(token string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.refreshTokens, token)
}

func (s *AuthService) issueTokens(userID, role string) (*AuthTokens, error) {
	accessToken, err := s.generateAccessToken(userID, role)
	if err != nil {
		return nil, err
	}

	refreshToken := uuid.New().String()
	expiresAt := time.Now().Add(s.refreshTTL)

	s.mu.Lock()
	s.refreshTokens[refreshToken] = &RefreshSession{
		Token:     refreshToken,
		UserID:    userID,
		Role:      role,
		ExpiresAt: expiresAt,
	}
	s.mu.Unlock()

	return &AuthTokens{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(s.accessTTL.Seconds()),
	}, nil
}

func (s *AuthService) generateAccessToken(userID, role string) (string, error) {
	now := time.Now()
	headerSegment := base64URLEncode([]byte(`{"alg":"HS256","typ":"JWT"}`))
	claims := Claims{
		UserID:    userID,
		Role:      role,
		Subject:   userID,
		IssuedAt:  now.Unix(),
		ExpiresAt: now.Add(s.accessTTL).Unix(),
		NotBefore: now.Unix(),
	}

	payloadBytes, err := json.Marshal(claims)
	if err != nil {
		return "", fmt.Errorf("序列化访问令牌载荷失败: %w", err)
	}

	payloadSegment := base64URLEncode(payloadBytes)
	unsigned := headerSegment + "." + payloadSegment
	signature := signToken(unsigned, s.signingKey)

	return unsigned + "." + signature, nil
}

func signToken(unsigned string, key []byte) string {
	mac := hmac.New(sha256.New, key)
	mac.Write([]byte(unsigned))
	return base64URLEncode(mac.Sum(nil))
}

func base64URLEncode(data []byte) string {
	return base64.RawURLEncoding.EncodeToString(data)
}

func base64URLDecode(s string) ([]byte, error) {
	return base64.RawURLEncoding.DecodeString(s)
}

func loadSigningKey(path string) ([]byte, error) {
	if strings.TrimSpace(path) == "" {
		return nil, errors.New("JWT_SECRET_PATH 未配置")
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("读取 JWT 私钥失败: %w", err)
	}

	key := strings.TrimSpace(string(content))
	if key == "" {
		return nil, errors.New("JWT 私钥文件内容为空")
	}
	if len(key) < 32 {
		return nil, errors.New("JWT 私钥长度至少 32 个字符")
	}

	return []byte(key), nil
}
