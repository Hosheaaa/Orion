package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hoshea/orion-backend/internal/app"
)

// AuthHandler 管理员认证接口
type AuthHandler struct {
	authService *app.AuthService
}

// NewAuthHandler 创建认证处理器
func NewAuthHandler(authService *app.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// LoginRequest 登录请求体
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// RefreshRequest 刷新令牌请求体
type RefreshRequest struct {
	RefreshToken string `json:"refreshToken" binding:"required"`
}

// Login 管理员登录
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "INVALID_REQUEST",
			"message": "请求参数错误",
			"data":    nil,
		})
		return
	}

	tokens, err := h.authService.Authenticate(req.Username, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    "UNAUTHORIZED",
			"message": err.Error(),
			"data":    nil,
		})
		return
	}

	c.JSON(http.StatusOK, tokens)
}

// Refresh 刷新访问令牌
func (h *AuthHandler) Refresh(c *gin.Context) {
	var req RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "INVALID_REQUEST",
			"message": "请求参数错误",
			"data":    nil,
		})
		return
	}

	tokens, err := h.authService.Refresh(req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    "UNAUTHORIZED",
			"message": err.Error(),
			"data":    nil,
		})
		return
	}

	c.JSON(http.StatusOK, tokens)
}
