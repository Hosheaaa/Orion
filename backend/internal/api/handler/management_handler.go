package handler

import (
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hoshea/orion-backend/internal/app"
	"github.com/hoshea/orion-backend/internal/domain"
)

// ManagementHandler 管理端辅助接口
type ManagementHandler struct {
	accessService *app.AccessService
}

// NewManagementHandler 创建管理端处理器
func NewManagementHandler(accessService *app.AccessService) *ManagementHandler {
	return &ManagementHandler{accessService: accessService}
}

// GenerateSpeakerToken 生成演讲者令牌
func (h *ManagementHandler) GenerateSpeakerToken(c *gin.Context) {
	activityID := c.Param("id")
	token, err := h.accessService.GenerateSpeakerToken(activityID)
	if err != nil {
		writeError(c, http.StatusBadRequest, "GENERATE_SPEAKER_TOKEN_FAILED", err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token":     token.Value,
		"expiresAt": token.ExpiresAt.UTC().Format(time.RFC3339),
	})
}

// RevokeSpeakerTokens 撤销演讲者令牌
func (h *ManagementHandler) RevokeSpeakerTokens(c *gin.Context) {
	activityID := c.Param("id")
	if err := h.accessService.RevokeSpeakerTokens(activityID); err != nil {
		writeError(c, http.StatusBadRequest, "REVOKE_SPEAKER_TOKEN_FAILED", err.Error())
		return
	}

	c.Status(http.StatusNoContent)
}

// RevokeSpeakerToken 撤销单个演讲者令牌
func (h *ManagementHandler) RevokeSpeakerToken(c *gin.Context) {
	activityID := c.Param("id")
	tokenID := c.Param("tokenId")
	if err := h.accessService.RevokeSpeakerToken(activityID, tokenID); err != nil {
		writeError(c, http.StatusBadRequest, "REVOKE_SPEAKER_TOKEN_FAILED", err.Error())
		return
	}

	c.Status(http.StatusNoContent)
}

// GenerateViewerToken 生成观众邀请码
func (h *ManagementHandler) GenerateViewerToken(c *gin.Context) {
	activityID := c.Param("id")
	var req domain.GenerateViewerTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		if !errors.Is(err, io.EOF) {
			writeError(c, http.StatusBadRequest, "INVALID_REQUEST", "请求参数错误")
			return
		}
	}

	token, err := h.accessService.GenerateViewerToken(activityID, &req)
	if err != nil {
		writeError(c, http.StatusBadRequest, "GENERATE_VIEWER_TOKEN_FAILED", err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":      token.Value,
		"expiresAt": token.ExpiresAt.UTC().Format(time.RFC3339),
	})
}

// ListTokens 列出活动令牌
func (h *ManagementHandler) ListTokens(c *gin.Context) {
	activityID := c.Param("id")
	tokens, err := h.accessService.ListTokens(activityID)
	if err != nil {
		writeError(c, http.StatusBadRequest, "LIST_TOKENS_FAILED", err.Error())
		return
	}

	response := make([]gin.H, 0, len(tokens))
	for _, token := range tokens {
		item := gin.H{
			"id":        token.ID,
			"type":      token.Type,
			"value":     token.Value,
			"status":    token.Status,
			"createdAt": token.CreatedAt.UTC().Format(time.RFC3339),
			"expiresAt": token.ExpiresAt.UTC().Format(time.RFC3339),
		}
		if token.MaxAudience != nil {
			item["maxAudience"] = *token.MaxAudience
		}
		response = append(response, item)
	}

	c.JSON(http.StatusOK, response)
}

// GetViewerEntry 获取观众入口详情
func (h *ManagementHandler) GetViewerEntry(c *gin.Context) {
	activityID := c.Param("id")
	entry, err := h.accessService.GetViewerEntry(activityID)
	if err != nil {
		writeError(c, http.StatusBadRequest, "GET_VIEWER_ENTRY_FAILED", err.Error())
		return
	}

	c.JSON(http.StatusOK, entry)
}

// RevokeViewerEntry 撤销观众入口
func (h *ManagementHandler) RevokeViewerEntry(c *gin.Context) {
	activityID := c.Param("id")
	entry, err := h.accessService.RevokeViewerEntry(activityID)
	if err != nil {
		writeError(c, http.StatusBadRequest, "REVOKE_VIEWER_ENTRY_FAILED", err.Error())
		return
	}

	c.JSON(http.StatusOK, entry)
}

// ActivateViewerEntry 重新启用观众入口
func (h *ManagementHandler) ActivateViewerEntry(c *gin.Context) {
	activityID := c.Param("id")
	entry, err := h.accessService.ActivateViewerEntry(activityID)
	if err != nil {
		writeError(c, http.StatusBadRequest, "ACTIVATE_VIEWER_ENTRY_FAILED", err.Error())
		return
	}

	c.JSON(http.StatusOK, entry)
}

// UploadCover 处理封面上传（占位实现）
func (h *ManagementHandler) UploadCover(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{
		"code":    "UPLOAD_NOT_IMPLEMENTED",
		"message": "封面上传功能尚未接入文件存储，请稍后重试",
	})
}

// GetLanguages 返回语言列表
func (h *ManagementHandler) GetLanguages(c *gin.Context) {
	c.JSON(http.StatusOK, []gin.H{
		{"code": "zh-CN", "name": "中文"},
		{"code": "en", "name": "English"},
		{"code": "ja", "name": "日本語"},
		{"code": "es", "name": "Español"},
	})
}

func writeError(c *gin.Context, status int, code, message string) {
	c.JSON(status, gin.H{
		"code":    code,
		"message": message,
		"data":    nil,
	})
}
