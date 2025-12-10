package router

import (
	"noti-service/internal/config"
	"noti-service/internal/handler"
	"noti-service/internal/middleware"
	"noti-service/internal/repository"
	"noti-service/internal/service"
	"noti-service/internal/sse"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	// swaggerFiles "github.com/swaggo/files"
	// ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/zap"
	"gorm.io/gorm"

	commonmw "github.com/OrangesCloud/wealist-advanced-go-pkg/middleware"
)

func Setup(cfg *config.Config, db *gorm.DB, redisClient *redis.Client, logger *zap.Logger) *gin.Engine {
	if cfg.Server.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()

	// Middleware (using common package)
	r.Use(commonmw.Recovery(logger))
	r.Use(commonmw.Logger(logger))
	r.Use(commonmw.DefaultCORS())

	// Initialize services
	notificationRepo := repository.NewNotificationRepository(db)
	sseService := sse.NewSSEService(redisClient, logger)
	notificationService := service.NewNotificationService(notificationRepo, redisClient, cfg, logger)

	// Initialize handlers
	validator := middleware.NewAuthServiceValidator(cfg.Auth.ServiceURL, cfg.Auth.SecretKey, logger)
	notificationHandler := handler.NewNotificationHandler(notificationService, sseService, logger)
	healthHandler := handler.NewHealthHandler(db, redisClient, sseService)

	// Health endpoints (no auth)
	r.GET("/health", healthHandler.Health)
	r.GET("/ready", healthHandler.Ready)
	r.GET("/metrics", healthHandler.Metrics)

	// Swagger documentation (disabled for faster builds)
	// r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// API routes
	api := r.Group("/api")
	{
		// SSE stream endpoint (uses query param token because EventSource doesn't support headers)
		api.GET("/notifications/stream", middleware.SSEAuthMiddleware(validator), notificationHandler.StreamNotifications)

		// Notification routes (require auth via Authorization header)
		notifications := api.Group("/notifications")
		notifications.Use(middleware.AuthMiddleware(validator))
		notifications.Use(middleware.WorkspaceMiddleware())
		{
			notifications.GET("", middleware.RequireWorkspace(), notificationHandler.GetNotifications)
			notifications.GET("/unread-count", middleware.RequireWorkspace(), notificationHandler.GetUnreadCount)
			notifications.PATCH("/:id/read", notificationHandler.MarkAsRead)
			notifications.POST("/read-all", middleware.RequireWorkspace(), notificationHandler.MarkAllAsRead)
			notifications.DELETE("/:id", notificationHandler.DeleteNotification)
		}

		// Internal API routes (require API key)
		internal := api.Group("/internal")
		internal.Use(middleware.InternalAuthMiddleware(cfg.Auth.InternalAPIKey))
		{
			internal.POST("/notifications", notificationHandler.CreateNotification)
			internal.POST("/notifications/bulk", notificationHandler.CreateBulkNotifications)
		}
	}

	return r
}
