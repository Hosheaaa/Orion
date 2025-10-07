package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hoshea/orion-backend/internal/app"
	"github.com/hoshea/orion-backend/internal/domain"
)

// ActivityHandler 活动处理器
type ActivityHandler struct {
	service *app.ActivityService
}

// NewActivityHandler 创建活动处理器
func NewActivityHandler(service *app.ActivityService) *ActivityHandler {
	return &ActivityHandler{
		service: service,
	}
}

// ListActivities 列出活动
// @Summary 获取活动列表
// @Description 根据状态过滤活动列表，不传 status 则返回所有活动
// @Tags activities
// @Accept json
// @Produce json
// @Param status query string false "活动状态" Enums(draft, published, closed)
// @Success 200 {array} domain.Activity
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/activities [get]
func (h *ActivityHandler) ListActivities(c *gin.Context) {
	// 获取查询参数
	statusParam := c.Query("status")

	var statusFilter *domain.ActivityStatus
	if statusParam != "" {
		status := domain.ActivityStatus(statusParam)
		// 验证状态值
		if status != domain.ActivityStatusDraft &&
		   status != domain.ActivityStatusPublished &&
		   status != domain.ActivityStatusClosed {
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Code:    "INVALID_STATUS",
				Message: "无效的活动状态",
				Data:    nil,
			})
			return
		}
		statusFilter = &status
	}

	// 调用服务层
	activities, err := h.service.ListActivities(statusFilter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "INTERNAL_ERROR",
			Message: "获取活动列表失败",
			Data:    nil,
		})
		return
	}

	c.JSON(http.StatusOK, activities)
}

// CreateActivity 创建活动
// @Summary 创建新活动
// @Description 创建一个新的活动
// @Tags activities
// @Accept json
// @Produce json
// @Param activity body domain.CreateActivityRequest true "活动信息"
// @Success 201 {object} domain.Activity
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/activities [post]
func (h *ActivityHandler) CreateActivity(c *gin.Context) {
	var req domain.CreateActivityRequest

	// 绑定并验证请求
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_REQUEST",
			Message: "请求数据格式错误: " + err.Error(),
			Data:    nil,
		})
		return
	}

	// 调用服务层创建活动
	activity, err := h.service.CreateActivity(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "CREATE_FAILED",
			Message: err.Error(),
			Data:    nil,
		})
		return
	}

	c.JSON(http.StatusCreated, activity)
}

// GetActivity 获取活动详情
// @Summary 获取活动详情
// @Description 根据 ID 获取活动的详细信息
// @Tags activities
// @Accept json
// @Produce json
// @Param id path string true "活动 ID"
// @Success 200 {object} domain.Activity
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/activities/{id} [get]
func (h *ActivityHandler) GetActivity(c *gin.Context) {
	id := c.Param("id")

	activity, err := h.service.GetActivity(id)
	if err != nil {
		if err == domain.ErrActivityNotFound {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Code:    "ACTIVITY_NOT_FOUND",
				Message: "活动不存在",
				Data:    nil,
			})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "INTERNAL_ERROR",
			Message: "获取活动失败",
			Data:    nil,
		})
		return
	}

	c.JSON(http.StatusOK, activity)
}

// UpdateActivity 更新活动
// @Summary 更新活动信息
// @Description 更新活动的信息（不能修改已关闭的活动）
// @Tags activities
// @Accept json
// @Produce json
// @Param id path string true "活动 ID"
// @Param activity body domain.UpdateActivityRequest true "更新的活动信息"
// @Success 200 {object} domain.Activity
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/activities/{id} [put]
func (h *ActivityHandler) UpdateActivity(c *gin.Context) {
	id := c.Param("id")

	var req domain.UpdateActivityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_REQUEST",
			Message: "请求数据格式错误: " + err.Error(),
			Data:    nil,
		})
		return
	}

	activity, err := h.service.UpdateActivity(id, &req)
	if err != nil {
		if err == domain.ErrActivityNotFound {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Code:    "ACTIVITY_NOT_FOUND",
				Message: "活动不存在",
				Data:    nil,
			})
			return
		}
		if err == domain.ErrActivityCannotBeModified {
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Code:    "ACTIVITY_CANNOT_BE_MODIFIED",
				Message: "活动无法修改",
				Data:    nil,
			})
			return
		}
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "UPDATE_FAILED",
			Message: err.Error(),
			Data:    nil,
		})
		return
	}

	c.JSON(http.StatusOK, activity)
}

// DeleteActivity 删除活动
// @Summary 删除活动
// @Description 删除活动（只能删除草稿状态的活动）
// @Tags activities
// @Accept json
// @Produce json
// @Param id path string true "活动 ID"
// @Success 204 "删除成功"
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/activities/{id} [delete]
func (h *ActivityHandler) DeleteActivity(c *gin.Context) {
	id := c.Param("id")

	err := h.service.DeleteActivity(id)
	if err != nil {
		if err == domain.ErrActivityNotFound {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Code:    "ACTIVITY_NOT_FOUND",
				Message: "活动不存在",
				Data:    nil,
			})
			return
		}
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "DELETE_FAILED",
			Message: err.Error(),
			Data:    nil,
		})
		return
	}

	c.Status(http.StatusNoContent)
}

// PublishActivity 发布活动
// @Summary 发布活动
// @Description 将活动从草稿状态发布为已发布状态
// @Tags activities
// @Accept json
// @Produce json
// @Param id path string true "活动 ID"
// @Success 200 {object} domain.Activity
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/activities/{id}/publish [post]
func (h *ActivityHandler) PublishActivity(c *gin.Context) {
	id := c.Param("id")

	activity, err := h.service.PublishActivity(id)
	if err != nil {
		if err == domain.ErrActivityNotFound {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Code:    "ACTIVITY_NOT_FOUND",
				Message: "活动不存在",
				Data:    nil,
			})
			return
		}
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "PUBLISH_FAILED",
			Message: err.Error(),
			Data:    nil,
		})
		return
	}

	c.JSON(http.StatusOK, activity)
}

// CloseActivity 关闭活动
// @Summary 关闭活动
// @Description 关闭正在进行的活动
// @Tags activities
// @Accept json
// @Produce json
// @Param id path string true "活动 ID"
// @Success 200 {object} domain.Activity
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/activities/{id}/close [post]
func (h *ActivityHandler) CloseActivity(c *gin.Context) {
	id := c.Param("id")

	activity, err := h.service.CloseActivity(id)
	if err != nil {
		if err == domain.ErrActivityNotFound {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Code:    "ACTIVITY_NOT_FOUND",
				Message: "活动不存在",
				Data:    nil,
			})
			return
		}
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "CLOSE_FAILED",
			Message: err.Error(),
			Data:    nil,
		})
		return
	}

	c.JSON(http.StatusOK, activity)
}

// ErrorResponse 错误响应结构
type ErrorResponse struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}
