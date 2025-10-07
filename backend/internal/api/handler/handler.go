package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// 占位处理器 - 后续逐步实现

func Login(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"accessToken":  "placeholder-token",
		"refreshToken": "placeholder-refresh",
		"expiresIn":    7200,
	})
}

func RefreshToken(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"accessToken": "placeholder-new-token",
		"expiresIn":   7200,
	})
}

func ListActivities(c *gin.Context) {
	c.JSON(http.StatusOK, []interface{}{})
}

func CreateActivity(c *gin.Context) {
	c.JSON(http.StatusCreated, gin.H{
		"id":     "activity-1",
		"status": "draft",
	})
}

func GetActivity(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"id": c.Param("id"),
	})
}

func UpdateActivity(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"id": c.Param("id"),
	})
}

func DeleteActivity(c *gin.Context) {
	c.JSON(http.StatusNoContent, nil)
}

func PublishActivity(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"id":     c.Param("id"),
		"status": "published",
	})
}

func CloseActivity(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"id":     c.Param("id"),
		"status": "closed",
	})
}

func GenerateSpeakerToken(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"token":     "speaker-token",
		"expiresAt": "2024-12-31T23:59:59Z",
	})
}

func GenerateViewerToken(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"code":      "ABCDE",
		"expiresAt": "2024-12-31T23:59:59Z",
	})
}

func ListTokens(c *gin.Context) {
	c.JSON(http.StatusOK, []interface{}{})
}

func GetViewerEntry(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"shareUrl":  "https://viewer.example.com/activity/123",
		"qrType":    "png",
		"qrContent": "data:image/png;base64,placeholder",
		"status":    "active",
	})
}

func RevokeViewerEntry(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "revoked",
	})
}

func ActivateViewerEntry(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "active",
		"qrContent": "data:image/png;base64,placeholder",
	})
}

func UploadCover(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"url": "https://example.com/cover.png",
	})
}

func GetLanguages(c *gin.Context) {
	c.JSON(http.StatusOK, []gin.H{
		{"code": "zh-CN", "name": "中文"},
		{"code": "en", "name": "English"},
		{"code": "ja", "name": "日本語"},
		{"code": "es", "name": "Español"},
	})
}

func HandleSpeakerWebSocket(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{
		"message": "WebSocket not implemented yet",
	})
}

func HandleViewerWebSocket(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{
		"message": "WebSocket not implemented yet",
	})
}
