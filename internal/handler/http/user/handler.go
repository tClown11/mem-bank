package user

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"mem_bank/internal/domain/user"
)

// Handler handles HTTP requests for user operations
type Handler struct {
	service user.Service
}

// NewHandler creates a new user HTTP handler
func NewHandler(service user.Service) *Handler {
	return &Handler{
		service: service,
	}
}

// CreateUserRequest represents the JSON request for creating a user
type CreateUserRequest struct {
	Username string        `json:"username" binding:"required"`
	Email    string        `json:"email" binding:"required,email"`
	Profile  user.Profile  `json:"profile"`
	Settings user.Settings `json:"settings"`
}

// UpdateUserRequest represents the JSON request for updating a user
type UpdateUserRequest struct {
	Username *string        `json:"username,omitempty"`
	Email    *string        `json:"email,omitempty"`
	Profile  *user.Profile  `json:"profile,omitempty"`
	Settings *user.Settings `json:"settings,omitempty"`
	IsActive *bool          `json:"is_active,omitempty"`
}

func (h *Handler) CreateUser(c *gin.Context) {
	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Convert to domain request
	createReq := user.CreateRequest{
		Username: req.Username,
		Email:    req.Email,
		Profile:  req.Profile,
		Settings: req.Settings,
	}

	u, err := h.service.CreateUser(c.Request.Context(), createReq)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, h.toResponse(u))
}

func (h *Handler) GetUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	u, err := h.service.GetUser(c.Request.Context(), user.ID(id))
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, h.toResponse(u))
}

func (h *Handler) GetUserByUsername(c *gin.Context) {
	username := c.Param("username")
	if username == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Username is required"})
		return
	}

	u, err := h.service.GetUserByUsername(c.Request.Context(), username)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, h.toResponse(u))
}

func (h *Handler) GetUserByEmail(c *gin.Context) {
	email := c.Query("email")
	if email == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email parameter is required"})
		return
	}

	u, err := h.service.GetUserByEmail(c.Request.Context(), email)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, h.toResponse(u))
}

func (h *Handler) UpdateUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Convert to domain request
	updateReq := user.UpdateRequest{
		Username: req.Username,
		Email:    req.Email,
		Profile:  req.Profile,
		Settings: req.Settings,
		IsActive: req.IsActive,
	}

	u, err := h.service.UpdateUser(c.Request.Context(), user.ID(id), updateReq)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, h.toResponse(u))
}

func (h *Handler) DeleteUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	err = h.service.DeleteUser(c.Request.Context(), user.ID(id))
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

func (h *Handler) ListUsers(c *gin.Context) {
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

	users, err := h.service.ListUsers(c.Request.Context(), limit, offset)
	if err != nil {
		h.handleError(c, err)
		return
	}

	response := make([]interface{}, len(users))
	for i, u := range users {
		response[i] = h.toResponse(u)
	}

	c.JSON(http.StatusOK, gin.H{
		"users":  response,
		"limit":  limit,
		"offset": offset,
	})
}

func (h *Handler) GetUserStats(c *gin.Context) {
	stats, err := h.service.GetUserStats(c.Request.Context())
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, stats)
}

func (h *Handler) UpdateLastLogin(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	err = h.service.UpdateLastLogin(c.Request.Context(), user.ID(id))
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Last login updated successfully"})
}

func (h *Handler) handleError(c *gin.Context, err error) {
	switch err {
	case user.ErrNotFound:
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case user.ErrAlreadyExists, user.ErrEmailTaken, user.ErrUsernameTaken:
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	case user.ErrInvalidEmail, user.ErrInvalidUsername, user.ErrInvalidID:
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case user.ErrInactive:
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
	}
}

func (h *Handler) toResponse(u *user.User) interface{} {
	return map[string]interface{}{
		"id":         u.ID,
		"username":   u.Username,
		"email":      u.Email,
		"profile":    u.Profile,
		"settings":   u.Settings,
		"created_at": u.CreatedAt,
		"updated_at": u.UpdatedAt,
		"last_login": u.LastLogin,
		"is_active":  u.IsActive,
	}
}