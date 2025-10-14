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
	// 初始化认证服务
	authService, err := app.NewAuthService(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize auth service: %v", err)
	}
	authHandler := handler.NewAuthHandler(authService)

	// 初始化依赖
	activityRepo := repository.NewMemoryActivityRepository()
	activityService := app.NewActivityService(activityRepo, cfg)
	activityHandler := handler.NewActivityHandler(activityService)
	accessService := app.NewAccessService(activityRepo, cfg.ViewerBaseURL)
	managementHandler := handler.NewManagementHandler(accessService)

	// 初始化翻译管线（如果 API Key 存在）
	var translationPipeline *app.TranslationPipeline
	if cfg.Google.STTAPIKey != "" && cfg.Google.TranslateAPIKey != "" {
		tp, err := app.NewTranslationPipeline(
			context.Background(),
			cfg.Google.STTAPIKey,
			cfg.Google.TranslateAPIKey,
		)
		if err != nil {
			log.Printf("Warning: Failed to initialize translation pipeline: %v", err)
		} else {
			translationPipeline = tp
			log.Println("Translation pipeline initialized successfully")
		}
	}

	// 初始化字幕广播器
	subtitleBroadcaster := app.NewSubtitleBroadcaster()

	// 初始化 WebSocket 处理器
	var speakerWSHandler *handler.SpeakerWebSocketHandler
	var viewerWSHandler *handler.ViewerWebSocketHandler

	if translationPipeline != nil {
		speakerWSHandler = handler.NewSpeakerWebSocketHandler(translationPipeline, subtitleBroadcaster, accessService)
		viewerWSHandler = handler.NewViewerWebSocketHandler(subtitleBroadcaster, accessService)
		log.Println("WebSocket handlers initialized")
	}
	// 根据环境设置 Gin 模式
	if cfg.Server.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()

	// 全局中间件
	router.Use(middleware.CORS(cfg.Server.AllowedOrigins))
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
			auth.POST("/login", authHandler.Login)
			auth.POST("/refresh", authHandler.Refresh)
		}

		// 活动路由（需要认证）
		activities := v1.Group("/activities")
		activities.Use(middleware.AuthRequired(authService))
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
		tokens.Use(middleware.AuthRequired(authService))
		{
			tokens.POST("/speaker", managementHandler.GenerateSpeakerToken)
			tokens.POST("/viewer", managementHandler.GenerateViewerToken)
			tokens.GET("", managementHandler.ListTokens)
		}

		// 观众入口路由
		viewerEntry := v1.Group("/activities/:id/viewer-entry")
		viewerEntry.Use(middleware.AuthRequired(authService))
		{
			viewerEntry.GET("", managementHandler.GetViewerEntry)
			viewerEntry.POST("/revoke", managementHandler.RevokeViewerEntry)
			viewerEntry.POST("/activate", managementHandler.ActivateViewerEntry)
		}

		// 文件上传
		uploads := v1.Group("/uploads")
		uploads.Use(middleware.AuthRequired(authService))
		{
			uploads.POST("/cover", managementHandler.UploadCover)
		}

		// 语言列表
		v1.GET("/languages", managementHandler.GetLanguages)
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
