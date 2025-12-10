package handler

import (
	"context"
	"net/http"
	"noti-service/internal/sse"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type HealthHandler struct {
	db         *gorm.DB
	redis      *redis.Client
	sseService *sse.SSEService
}

func NewHealthHandler(db *gorm.DB, redis *redis.Client, sseService *sse.SSEService) *HealthHandler {
	return &HealthHandler{
		db:         db,
		redis:      redis,
		sseService: sseService,
	}
}

func (h *HealthHandler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"service": "notification-service",
	})
}

func (h *HealthHandler) Ready(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	connections := make(map[string]string)

	// Check database
	sqlDB, err := h.db.DB()
	if err != nil {
		connections["database"] = "error: " + err.Error()
	} else if err := sqlDB.PingContext(ctx); err != nil {
		connections["database"] = "error: " + err.Error()
	} else {
		connections["database"] = "connected"
	}

	// Check Redis
	if h.redis != nil {
		if err := h.redis.Ping(ctx).Err(); err != nil {
			connections["redis"] = "error: " + err.Error()
		} else {
			connections["redis"] = "connected"
		}
	} else {
		connections["redis"] = "not configured"
	}

	// Check for errors
	hasError := false
	for _, status := range connections {
		if status != "connected" && status != "not configured" {
			hasError = true
			break
		}
	}

	status := http.StatusOK
	statusText := "ready"
	if hasError {
		status = http.StatusServiceUnavailable
		statusText = "not ready"
	}

	c.JSON(status, gin.H{
		"status":      statusText,
		"connections": connections,
		"sseClients":  h.sseService.GetConnectedClientsCount(),
	})
}

func (h *HealthHandler) Metrics(c *gin.Context) {
	sseClients := h.sseService.GetConnectedClientsCount()

	// Prometheus-compatible metrics
	metrics := "# HELP notification_service_sse_clients_total Total SSE clients connected\n"
	metrics += "# TYPE notification_service_sse_clients_total gauge\n"
	metrics += "notification_service_sse_clients_total " + string(rune(sseClients+'0')) + "\n"
	metrics += "# HELP notification_service_up Service is up\n"
	metrics += "# TYPE notification_service_up gauge\n"
	metrics += "notification_service_up 1\n"

	c.String(http.StatusOK, metrics)
}
