package memory

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"mem_bank/internal/domain/memory"
	"mem_bank/internal/domain/user"
)

// Handler handles HTTP requests for memory operations
type Handler struct {
	service memory.Service
}

// NewHandler creates a new memory HTTP handler
func NewHandler(service memory.Service) *Handler {
	return &Handler{
		service: service,
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

func (h *Handler) CreateMemory(c *gin.Context) {
	var req CreateMemoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
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
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, h.toResponse(m))
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
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, h.toResponse(m))
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
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, h.toResponse(m))
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
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusNoContent, nil)
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
		h.handleError(c, err)
		return
	}

	response := make([]interface{}, len(memories))
	for i, m := range memories {
		response[i] = h.toResponse(m)
	}

	c.JSON(http.StatusOK, gin.H{
		"memories": response,
		"limit":    limit,
		"offset":   offset,
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
		h.handleError(c, err)
		return
	}

	response := make([]interface{}, len(memories))
	for i, m := range memories {
		response[i] = h.toResponse(m)
	}

	c.JSON(http.StatusOK, gin.H{
		"memories": response,
		"query":    req.Query,
		"limit":    req.Limit,
		"offset":   req.Offset,
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
		h.handleError(c, err)
		return
	}

	response := make([]interface{}, len(memories))
	for i, m := range memories {
		response[i] = h.toResponse(m)
	}

	c.JSON(http.StatusOK, gin.H{
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
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, stats)
}

func (h *Handler) handleError(c *gin.Context, err error) {
	switch err {
	case memory.ErrNotFound:
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case memory.ErrInvalidID, memory.ErrInvalidUserID, memory.ErrInvalidContent,
		memory.ErrInvalidImportance, memory.ErrInvalidMemoryType:
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case memory.ErrEmbeddingFailed:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process memory content"})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
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