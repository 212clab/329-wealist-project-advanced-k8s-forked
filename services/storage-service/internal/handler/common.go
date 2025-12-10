package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

// ErrorDetail contains error details
type ErrorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// SuccessResponse represents a success response
type SuccessResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

// respondWithError sends an error response
func respondWithError(c *gin.Context, code int, errorCode, message string) {
	c.JSON(code, ErrorResponse{
		Error: ErrorDetail{
			Code:    errorCode,
			Message: message,
		},
	})
}

// respondWithSuccess sends a success response
func respondWithSuccess(c *gin.Context, code int, message string, data interface{}) {
	c.JSON(code, SuccessResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// respondWithData sends a success response with data only
func respondWithData(c *gin.Context, code int, data interface{}) {
	c.JSON(code, SuccessResponse{
		Success: true,
		Data:    data,
	})
}

// getUserID extracts user ID from context
func getUserID(c *gin.Context) (uuid.UUID, bool) {
	userID, exists := c.Get("user_id")
	if !exists {
		return uuid.Nil, false
	}
	uid, ok := userID.(uuid.UUID)
	return uid, ok
}

// parseUUID parses a UUID from string
func parseUUID(idStr string) (uuid.UUID, error) {
	return uuid.Parse(idStr)
}

// getQueryInt gets an integer query parameter with default value
func getQueryInt(c *gin.Context, key string, defaultValue int) int {
	value := c.Query(key)
	if value == "" {
		return defaultValue
	}

	var result int
	if _, err := c.GetQuery(key); err {
		return defaultValue
	}

	if n, err := parseIntQuery(value); err == nil {
		result = n
	} else {
		result = defaultValue
	}

	return result
}

// parseIntQuery parses int from query string
func parseIntQuery(s string) (int, error) {
	var n int
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0, nil
		}
		n = n*10 + int(c-'0')
	}
	return n, nil
}

// handleNotFound handles 404 response
func handleNotFound(c *gin.Context, message string) {
	respondWithError(c, http.StatusNotFound, "NOT_FOUND", message)
}

// handleBadRequest handles 400 response
func handleBadRequest(c *gin.Context, message string) {
	respondWithError(c, http.StatusBadRequest, "BAD_REQUEST", message)
}

// handleUnauthorized handles 401 response
func handleUnauthorized(c *gin.Context, message string) {
	respondWithError(c, http.StatusUnauthorized, "UNAUTHORIZED", message)
}

// handleForbidden handles 403 response
func handleForbidden(c *gin.Context, message string) {
	respondWithError(c, http.StatusForbidden, "FORBIDDEN", message)
}

// handleInternalError handles 500 response
func handleInternalError(c *gin.Context, message string) {
	respondWithError(c, http.StatusInternalServerError, "INTERNAL_ERROR", message)
}
