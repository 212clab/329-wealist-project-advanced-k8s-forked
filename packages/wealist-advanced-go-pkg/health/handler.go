package health

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// HealthChecker provides standardized health check endpoints for Kubernetes
type HealthChecker struct {
	db    *gorm.DB
	redis *redis.Client
}

// NewHealthChecker creates a new HealthChecker instance
// db: Required database connection (can be nil if no DB)
// redis: Optional Redis connection (can be nil)
func NewHealthChecker(db *gorm.DB, redis *redis.Client) *HealthChecker {
	return &HealthChecker{
		db:    db,
		redis: redis,
	}
}

// RegisterRoutes registers health check endpoints at root level and under basePath
// This ensures health checks work both for direct pod access and through ingress
func (h *HealthChecker) RegisterRoutes(router gin.IRouter, basePath string) {
	// Root level endpoints (for K8s probes on base deployment)
	router.GET("/health/live", h.Liveness)
	router.GET("/health/ready", h.Readiness)
	router.GET("/health", h.Readiness) // Backwards compatibility

	// Under basePath (for ingress routing with SERVER_BASE_PATH)
	if basePath != "" {
		group := router.Group(basePath)
		group.GET("/health/live", h.Liveness)
		group.GET("/health/ready", h.Readiness)
		group.GET("/health", h.Readiness)
	}
}

// Liveness checks if the application process is running
// This is for Kubernetes liveness probe - does NOT check dependencies
// Returns 200 if process is alive (should always return 200)
func (h *HealthChecker) Liveness(c *gin.Context) {
	c.JSON(200, gin.H{
		"status": "alive",
	})
}

// Readiness checks if the service is ready to accept traffic
// This is for Kubernetes readiness probe - checks DB and Redis connections
// Returns 200 if ready, 503 if not ready
func (h *HealthChecker) Readiness(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	checks := gin.H{}
	isReady := true

	// Check database connection
	if h.db != nil {
		sqlDB, err := h.db.DB()
		if err != nil {
			checks["database"] = "error"
			isReady = false
		} else if err := sqlDB.PingContext(ctx); err != nil {
			checks["database"] = "disconnected"
			isReady = false
		} else {
			checks["database"] = "ok"
		}
	} else {
		checks["database"] = "not_configured"
	}

	// Check Redis connection (optional)
	if h.redis != nil {
		if err := h.redis.Ping(ctx).Err(); err != nil {
			checks["redis"] = "disconnected"
			isReady = false
		} else {
			checks["redis"] = "ok"
		}
	} else {
		checks["redis"] = "not_configured"
	}

	if isReady {
		c.JSON(200, gin.H{
			"status": "ready",
			"checks": checks,
		})
	} else {
		c.JSON(503, gin.H{
			"status": "not_ready",
			"checks": checks,
		})
	}
}
