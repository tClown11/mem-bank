package memory

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"mem_bank/internal/domain/memory"
	"mem_bank/internal/domain/user"
	"mem_bank/pkg/logger"
)

// Handler handles HTTP requests for memory operations
type Handler struct {
	service memory.Service
	logger  logger.Logger
}

// NewHandler creates a new memory HTTP handler
func NewHandler(service memory.Service, logger logger.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

// CreateMemoryRequest represents the JSON request for creating a memory
type CreateMemoryRequest struct {
	UserID     string                 `json:"user_id" binding:"required"`
	Content    string                 `json:"content" binding:"required"`
	Summary    string                 `json:"summary"`
	Importance int                    `json:"importance" binding:"min=1,max=10"`
	MemoryType string                 `json:"memory_type" binding:"required"`
	Tags       []string               `json:"tags"`
	Metadata   map[string]interface{} `json:"metadata"`
}

// UpdateMemoryRequest represents the JSON request for updating a memory
type UpdateMemoryRequest struct {
	Content    *string                `json:"content,omitempty"`
	Summary    *string                `json:"summary,omitempty"`
	Importance *int                   `json:"importance,omitempty"`
	MemoryType *string                `json:"memory_type,omitempty"`
	Tags       []string               `json:"tags,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// SearchMemoryRequest represents the JSON request for searching memories
type SearchMemoryRequest struct {
	Query      string   `json:"query"`
	Tags       []string `json:"tags"`
	MemoryType string   `json:"memory_type"`
	Limit      int      `json:"limit"`
	Offset     int      `json:"offset"`
}

// BatchCreateRequest represents batch memory creation request
type BatchCreateRequest struct {
	Memories []CreateMemoryRequest `json:"memories" binding:"required,dive"`
}

// StandardResponse represents a standardized API response
type StandardResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *ErrorInfo  `json:"error,omitempty"`
}

// ErrorInfo provides detailed error information
type ErrorInfo struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// PaginatedResponse represents a paginated response
type PaginatedResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data"`
	Meta    *PageMeta   `json:"meta"`
	Error   *ErrorInfo  `json:"error,omitempty"`
}

// PageMeta contains pagination metadata
type PageMeta struct {
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
	Total  int `json:"total,omitempty"`
}

func (h *Handler) CreateMemory(c *gin.Context) {
	var req CreateMemoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.sendErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid request data", err.Error())
		return
	}

	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		h.sendErrorResponse(c, http.StatusBadRequest, "INVALID_USER_ID", "Invalid user ID format", err.Error())
		return
	}

	// Convert to domain request
	createReq := memory.CreateRequest{
		UserID:     user.ID(userID),
		Content:    req.Content,
		Summary:    req.Summary,
		Importance: req.Importance,
		MemoryType: req.MemoryType,
		Tags:       req.Tags,
		Metadata:   req.Metadata,
	}

	m, err := h.service.CreateMemory(c.Request.Context(), createReq)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	h.sendSuccessResponse(c, http.StatusCreated, h.toResponse(m))
}

func (h *Handler) GetMemory(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid memory ID"})
		return
	}

	m, err := h.service.GetMemory(c.Request.Context(), memory.ID(id))
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	h.sendSuccessResponse(c, http.StatusOK, h.toResponse(m))
}

func (h *Handler) UpdateMemory(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid memory ID"})
		return
	}

	var req UpdateMemoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Convert to domain request
	updateReq := memory.UpdateRequest{
		Content:    req.Content,
		Summary:    req.Summary,
		Importance: req.Importance,
		MemoryType: req.MemoryType,
		Tags:       req.Tags,
		Metadata:   req.Metadata,
	}

	m, err := h.service.UpdateMemory(c.Request.Context(), memory.ID(id), updateReq)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	h.sendSuccessResponse(c, http.StatusOK, h.toResponse(m))
}

func (h *Handler) DeleteMemory(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid memory ID"})
		return
	}

	err = h.service.DeleteMemory(c.Request.Context(), memory.ID(id))
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	h.sendSuccessResponse(c, http.StatusNoContent, nil)
}

func (h *Handler) ListUserMemories(c *gin.Context) {
	userIDStr := c.Param("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	limitStr := c.DefaultQuery("limit", "20")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 20
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	memories, err := h.service.ListUserMemories(c.Request.Context(), user.ID(userID), limit, offset)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	response := make([]interface{}, len(memories))
	for i, m := range memories {
		response[i] = h.toResponse(m)
	}

	h.sendPaginatedResponse(c, response, &PageMeta{
		Limit:  limit,
		Offset: offset,
	})
}

func (h *Handler) SearchMemories(c *gin.Context) {
	userIDStr := c.Param("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var req SearchMemoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Limit <= 0 {
		req.Limit = 20
	}
	if req.Offset < 0 {
		req.Offset = 0
	}

	// Convert to domain request
	searchReq := memory.SearchRequest{
		UserID:     user.ID(userID),
		Query:      req.Query,
		Tags:       req.Tags,
		MemoryType: req.MemoryType,
		Limit:      req.Limit,
		Offset:     req.Offset,
	}

	memories, err := h.service.SearchMemories(c.Request.Context(), searchReq)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	response := make([]interface{}, len(memories))
	for i, m := range memories {
		response[i] = h.toResponse(m)
	}

	h.sendPaginatedResponse(c, response, &PageMeta{
		Limit:  req.Limit,
		Offset: req.Offset,
	})
}

func (h *Handler) SearchSimilarMemories(c *gin.Context) {
	userIDStr := c.Param("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	content := c.Query("content")
	if content == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Content parameter is required"})
		return
	}

	limitStr := c.DefaultQuery("limit", "10")
	thresholdStr := c.DefaultQuery("threshold", "0.8")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 10
	}

	threshold, err := strconv.ParseFloat(thresholdStr, 64)
	if err != nil || threshold <= 0 {
		threshold = 0.8
	}

	memories, err := h.service.SearchSimilarMemories(c.Request.Context(), content, user.ID(userID), limit, threshold)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	response := make([]interface{}, len(memories))
	for i, m := range memories {
		response[i] = h.toResponse(m)
	}

	h.sendSuccessResponse(c, http.StatusOK, map[string]interface{}{
		"memories":  response,
		"content":   content,
		"limit":     limit,
		"threshold": threshold,
	})
}

func (h *Handler) GetMemoryStats(c *gin.Context) {
	userIDStr := c.Param("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	stats, err := h.service.GetMemoryStats(c.Request.Context(), user.ID(userID))
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	h.sendSuccessResponse(c, http.StatusOK, stats)
}

// Response helper methods
func (h *Handler) sendSuccessResponse(c *gin.Context, status int, data interface{}) {
	c.JSON(status, StandardResponse{
		Success: true,
		Data:    data,
	})
}

func (h *Handler) sendPaginatedResponse(c *gin.Context, data interface{}, meta *PageMeta) {
	c.JSON(http.StatusOK, PaginatedResponse{
		Success: true,
		Data:    data,
		Meta:    meta,
	})
}

func (h *Handler) sendErrorResponse(c *gin.Context, status int, code, message, details string) {
	h.logger.WithFields(map[string]interface{}{
		"error_code": code,
		"message":    message,
		"details":    details,
		"path":       c.Request.URL.Path,
		"method":     c.Request.Method,
	}).Error("API error occurred")

	c.JSON(status, StandardResponse{
		Success: false,
		Error: &ErrorInfo{
			Code:    code,
			Message: message,
			Details: details,
		},
	})
}

func (h *Handler) handleServiceError(c *gin.Context, err error) {
	// Check for custom service errors
	var serviceErr *memory.ServiceError
	if errors.As(err, &serviceErr) {
		status := h.getStatusFromErrorCode(serviceErr.Code)
		h.sendErrorResponse(c, status, serviceErr.Code, serviceErr.Message, serviceErr.Error())
		return
	}

	// Check for validation errors
	var validationErr *memory.ValidationError
	if errors.As(err, &validationErr) {
		h.sendErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR",
			validationErr.Message, validationErr.Field)
		return
	}

	// Handle domain errors
	switch err {
	case memory.ErrNotFound:
		h.sendErrorResponse(c, http.StatusNotFound, "NOT_FOUND", "Memory not found", "")
	case memory.ErrInvalidID:
		h.sendErrorResponse(c, http.StatusBadRequest, "INVALID_ID", "Invalid memory ID", "")
	case memory.ErrInvalidUserID:
		h.sendErrorResponse(c, http.StatusBadRequest, "INVALID_USER_ID", "Invalid user ID", "")
	case memory.ErrInvalidContent:
		h.sendErrorResponse(c, http.StatusBadRequest, "INVALID_CONTENT", "Invalid memory content", "")
	case memory.ErrInvalidImportance:
		h.sendErrorResponse(c, http.StatusBadRequest, "INVALID_IMPORTANCE", "Invalid importance level", "")
	case memory.ErrInvalidMemoryType:
		h.sendErrorResponse(c, http.StatusBadRequest, "INVALID_MEMORY_TYPE", "Invalid memory type", "")
	case memory.ErrEmbeddingFailed:
		h.sendErrorResponse(c, http.StatusInternalServerError, "EMBEDDING_ERROR",
			"Failed to process memory content", "")
	default:
		h.logger.WithError(err).Error("Unhandled service error")
		h.sendErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR",
			"Internal server error", "")
	}
}

func (h *Handler) getStatusFromErrorCode(code string) int {
	switch code {
	case memory.ErrCodeNotFound:
		return http.StatusNotFound
	case memory.ErrCodeInvalidInput:
		return http.StatusBadRequest
	case memory.ErrCodePermissionDenied:
		return http.StatusForbidden
	case memory.ErrCodeExternalService:
		return http.StatusBadGateway
	default:
		return http.StatusInternalServerError
	}
}

func (h *Handler) toResponse(m *memory.Memory) interface{} {
	return map[string]interface{}{
		"id":            m.ID,
		"user_id":       m.UserID,
		"content":       m.Content,
		"summary":       m.Summary,
		"importance":    m.Importance,
		"memory_type":   m.MemoryType,
		"tags":          m.Tags,
		"metadata":      m.Metadata,
		"created_at":    m.CreatedAt,
		"updated_at":    m.UpdatedAt,
		"last_accessed": m.LastAccessed,
		"access_count":  m.AccessCount,
	}
}
