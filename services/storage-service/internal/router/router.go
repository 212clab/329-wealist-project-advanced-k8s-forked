package router

import (
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
	"gorm.io/gorm"

	commonmw "github.com/OrangesCloud/wealist-advanced-go-pkg/middleware"
	"storage-service/internal/client"
	"storage-service/internal/handler"
	"storage-service/internal/middleware"
	"storage-service/internal/repository"
	"storage-service/internal/service"
)

// Config holds router configuration
type Config struct {
	DB         *gorm.DB
	Logger     *zap.Logger
	JWTSecret  string
	BasePath   string
	S3Client   *client.S3Client
	AuthClient *client.AuthClient
}

// Setup sets up the router with all routes
func Setup(cfg Config) *gin.Engine {
	r := gin.New()

	// Middleware (using common package)
	r.Use(commonmw.Recovery(cfg.Logger))
	r.Use(commonmw.Logger(cfg.Logger))
	r.Use(commonmw.DefaultCORS())
	r.Use(commonmw.Metrics())

	// Prometheus metrics endpoint
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// Health check routes
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "healthy", "service": "storage-service"})
	})
	r.GET("/ready", func(c *gin.Context) {
		if cfg.DB == nil {
			c.JSON(503, gin.H{"status": "not ready", "service": "storage-service"})
			return
		}
		sqlDB, err := cfg.DB.DB()
		if err != nil {
			c.JSON(503, gin.H{"status": "not ready", "service": "storage-service"})
			return
		}
		if err := sqlDB.Ping(); err != nil {
			c.JSON(503, gin.H{"status": "not ready", "service": "storage-service"})
			return
		}
		c.JSON(200, gin.H{"status": "ready", "service": "storage-service"})
	})

	// Initialize repositories
	folderRepo := repository.NewFolderRepository(cfg.DB)
	fileRepo := repository.NewFileRepository(cfg.DB)
	shareRepo := repository.NewShareRepository(cfg.DB)

	// Initialize services
	folderService := service.NewFolderService(folderRepo, fileRepo, cfg.Logger)
	fileService := service.NewFileService(fileRepo, folderRepo, cfg.S3Client, cfg.Logger)
	shareService := service.NewShareService(shareRepo, fileRepo, folderRepo, cfg.Logger)

	// Initialize handlers
	folderHandler := handler.NewFolderHandler(folderService, fileService)
	fileHandler := handler.NewFileHandler(fileService)
	shareHandler := handler.NewShareHandler(shareService)

	// API routes group
	api := r.Group(cfg.BasePath)

	// Auth middleware
	var authMiddleware gin.HandlerFunc
	if cfg.AuthClient != nil {
		authMiddleware = middleware.AuthWithValidator(cfg.AuthClient)
	} else {
		authMiddleware = middleware.Auth(cfg.JWTSecret)
	}

	// ============================================================
	// Storage routes (authenticated)
	// ============================================================
	storage := api.Group("/storage")
	storage.Use(authMiddleware)
	{
		// ============================================================
		// Folder routes
		// ============================================================
		folders := storage.Group("/folders")
		{
			folders.POST("", folderHandler.CreateFolder)
			folders.GET("/contents", folderHandler.GetFolderContents)
			folders.GET("/:folderId", folderHandler.GetFolder)
			folders.PUT("/:folderId", folderHandler.UpdateFolder)
			folders.DELETE("/:folderId", folderHandler.DeleteFolder)
			folders.POST("/:folderId/restore", folderHandler.RestoreFolder)
			folders.DELETE("/:folderId/permanent", folderHandler.PermanentDeleteFolder)

			// Folder shares
			folders.GET("/:folderId/shares", shareHandler.GetFolderShares)
		}

		// ============================================================
		// File routes
		// ============================================================
		files := storage.Group("/files")
		{
			files.POST("/upload-url", fileHandler.GenerateUploadURL)
			files.POST("/confirm", fileHandler.ConfirmUpload)
			files.GET("/:fileId", fileHandler.GetFile)
			files.GET("/:fileId/download", fileHandler.GetDownloadURL)
			files.PUT("/:fileId", fileHandler.UpdateFile)
			files.DELETE("/:fileId", fileHandler.DeleteFile)
			files.POST("/:fileId/restore", fileHandler.RestoreFile)
			files.DELETE("/:fileId/permanent", fileHandler.PermanentDeleteFile)

			// File shares
			files.GET("/:fileId/shares", shareHandler.GetFileShares)
		}

		// ============================================================
		// Share routes
		// ============================================================
		shares := storage.Group("/shares")
		{
			shares.POST("", shareHandler.CreateShare)
			shares.GET("/link/:link", shareHandler.GetShareByLink)
			shares.PUT("/:shareId", shareHandler.UpdateShare)
			shares.DELETE("/:shareId", shareHandler.DeleteShare)
		}

		// Shared with me
		storage.GET("/shared-with-me", shareHandler.GetSharedWithMe)

		// ============================================================
		// Workspace routes
		// ============================================================
		workspaces := storage.Group("/workspaces")
		{
			workspaces.GET("/:workspaceId/folders", folderHandler.GetWorkspaceFolders)
			workspaces.GET("/:workspaceId/files", fileHandler.GetWorkspaceFiles)
			workspaces.GET("/:workspaceId/files/search", fileHandler.SearchFiles)
			workspaces.GET("/:workspaceId/usage", fileHandler.GetStorageUsage)

			// Trash
			workspaces.GET("/:workspaceId/trash/folders", folderHandler.GetTrashFolders)
			workspaces.GET("/:workspaceId/trash/files", fileHandler.GetTrashFiles)
		}
	}

	// ============================================================
	// Public routes (no auth required for shared links)
	// ============================================================
	public := api.Group("/public/storage")
	{
		public.GET("/shares/link/:link", shareHandler.GetShareByLink)
	}

	return r
}
