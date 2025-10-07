package api

import (
	"context"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hoshea/orion-backend/internal/api/handler"
	"github.com/hoshea/orion-backend/internal/api/middleware"
	"github.com/hoshea/orion-backend/internal/app"
	"github.com/hoshea/orion-backend/internal/infra/config"
	"github.com/hoshea/orion-backend/internal/infra/repository"
)

// SetupRouter 设置路由
func SetupRouter(cfg *config.Config) *gin.Engine {
	// 初始化依赖
	activityRepo := repository.NewMemoryActivityRepository()
	activityService := app.NewActivityService(activityRepo, cfg)
	activityHandler := handler.NewActivityHandler(activityService)

	// 初始化翻译管线（如果 API Key 存在）
	var translationPipeline *app.TranslationPipeline
	var err error
	if cfg.Google.STTAPIKey != "" && cfg.Google.TranslateAPIKey != "" {
		translationPipeline, err = app.NewTranslationPipeline(
			context.Background(),
			cfg.Google.STTAPIKey,
			cfg.Google.TranslateAPIKey,
		)
		if err != nil {
			log.Printf("Warning: Failed to initialize translation pipeline: %v", err)
		} else {
			log.Println("Translation pipeline initialized successfully")
		}
	}

	// 初始化字幕广播器
	subtitleBroadcaster := app.NewSubtitleBroadcaster()

	// 初始化 WebSocket 处理器
	var speakerWSHandler *handler.SpeakerWebSocketHandler
	var viewerWSHandler *handler.ViewerWebSocketHandler

	if translationPipeline != nil {
		speakerWSHandler = handler.NewSpeakerWebSocketHandler(translationPipeline, subtitleBroadcaster)
		viewerWSHandler = handler.NewViewerWebSocketHandler(subtitleBroadcaster)
		log.Println("WebSocket handlers initialized")
	}
	// 根据环境设置 Gin 模式
	if cfg.Server.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()

	// 全局中间件
	router.Use(middleware.CORS())
	router.Use(middleware.RequestID())

	// 健康检查
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
			"env":    cfg.Server.Env,
		})
	})

	// API v1 路由组
	v1 := router.Group("/api/v1")
	{
		// 认证路由
		auth := v1.Group("/auth")
		{
			auth.POST("/login", handler.Login)
			auth.POST("/refresh", handler.RefreshToken)
		}

		// 活动路由（需要认证）
		activities := v1.Group("/activities")
		activities.Use(middleware.AuthRequired())
		{
			activities.GET("", activityHandler.ListActivities)
			activities.POST("", activityHandler.CreateActivity)
			activities.GET("/:id", activityHandler.GetActivity)
			activities.PUT("/:id", activityHandler.UpdateActivity)
			activities.DELETE("/:id", activityHandler.DeleteActivity)
			activities.POST("/:id/publish", activityHandler.PublishActivity)
			activities.POST("/:id/close", activityHandler.CloseActivity)
		}

		// 令牌路由
		tokens := v1.Group("/activities/:id/tokens")
		tokens.Use(middleware.AuthRequired())
		{
			tokens.POST("/speaker", handler.GenerateSpeakerToken)
			tokens.POST("/viewer", handler.GenerateViewerToken)
			tokens.GET("", handler.ListTokens)
		}

		// 观众入口路由
		viewerEntry := v1.Group("/activities/:id/viewer-entry")
		viewerEntry.Use(middleware.AuthRequired())
		{
			viewerEntry.GET("", handler.GetViewerEntry)
			viewerEntry.POST("/revoke", handler.RevokeViewerEntry)
			viewerEntry.POST("/activate", handler.ActivateViewerEntry)
		}

		// 文件上传
		uploads := v1.Group("/uploads")
		uploads.Use(middleware.AuthRequired())
		{
			uploads.POST("/cover", handler.UploadCover)
		}

		// 语言列表
		v1.GET("/languages", handler.GetLanguages)
	}

	// WebSocket 路由
	ws := router.Group("/ws")
	{
		if speakerWSHandler != nil && viewerWSHandler != nil {
			ws.GET("/speaker", speakerWSHandler.HandleSpeakerWebSocket)
			ws.GET("/viewer", viewerWSHandler.HandleViewerWebSocket)
		} else {
			ws.GET("/speaker", func(c *gin.Context) {
				c.JSON(http.StatusServiceUnavailable, gin.H{
					"error": "WebSocket service not available - Translation pipeline not initialized",
				})
			})
			ws.GET("/viewer", func(c *gin.Context) {
				c.JSON(http.StatusServiceUnavailable, gin.H{
					"error": "WebSocket service not available - Translation pipeline not initialized",
				})
			})
		}
	}

	return router
}
