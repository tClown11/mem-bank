package app

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"mem_bank/configs"
	memoryDao "mem_bank/internal/dao/memory"
	userDao "mem_bank/internal/dao/user"
	memoryHandler "mem_bank/internal/handler/http/memory"
	userHandler "mem_bank/internal/handler/http/user"
	"mem_bank/internal/middleware"
	"mem_bank/internal/queue"
	embeddingService "mem_bank/internal/service/embedding"
	memoryService "mem_bank/internal/service/memory"
	userService "mem_bank/internal/service/user"
	"mem_bank/pkg/auth"
	"mem_bank/pkg/llm"
	"mem_bank/pkg/logger"
	"mem_bank/pkg/response"
)

// App represents the application
type App struct {
	server     *http.Server
	db         *gorm.DB
	redis      *redis.Client
	logger     logger.Logger
	config     *configs.Config
	jobQueue   queue.Queue
	jwtService *auth.JWTService
}

// Config holds application configuration
type Config struct {
	Host         string
	Port         int
	Mode         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

// NewApp creates a new application instance
func NewApp(db *gorm.DB, redisClient *redis.Client, appLogger logger.Logger, appConfig *configs.Config) *App {
	return &App{
		db:     db,
		redis:  redisClient,
		logger: appLogger,
		config: appConfig,
	}
}

// Start starts the application
func (a *App) Start(ctx context.Context, config Config) error {
	// Wire up dependencies using constructor-based dependency injection

	// Initialize JWT Service
	a.jwtService = auth.NewJWTService(
		a.config.Security.JWTSecret,
		"mem_bank",
		a.config.Security.JWTExpiry,
	)

	// Initialize LLM Provider
	llmConfig := &llm.Config{
		Provider:        a.config.LLM.Provider,
		APIKey:          a.config.LLM.APIKey,
		BaseURL:         a.config.LLM.BaseURL,
		EmbeddingModel:  a.config.LLM.EmbeddingModel,
		CompletionModel: a.config.LLM.CompletionModel,
		TimeoutSeconds:  a.config.LLM.TimeoutSeconds,
		MaxRetries:      a.config.LLM.MaxRetries,
		RateLimit:       a.config.LLM.RateLimit,
	}

	llmProvider := llm.NewOpenAIProvider(llmConfig)
	embeddingProvider := llmProvider // OpenAIProvider implements both interfaces

	// Initialize Embedding Service
	embeddingSvc := embeddingService.NewService(
		embeddingProvider,
		a.redis,
		a.logger,
		embeddingService.Config{
			MaxTextLength:   a.config.Embedding.MaxTextLength,
			CacheEnabled:    a.config.Embedding.CacheEnabled,
			CacheTTLMinutes: a.config.Embedding.CacheTTLMinutes,
			BatchSize:       a.config.Embedding.BatchSize,
			PreprocessingConfig: embeddingService.PreprocessingConfig{
				NormalizeWhitespace:    a.config.Embedding.NormalizeWhitespace,
				ToLowercase:            a.config.Embedding.ToLowercase,
				RemoveExtraPunctuation: a.config.Embedding.RemoveExtraPunctuation,
				ChunkSize:              a.config.Embedding.ChunkSize,
				ChunkOverlap:           a.config.Embedding.ChunkOverlap,
			},
		},
	)

	// Initialize Job Queue
	a.jobQueue = queue.NewRedisQueue(
		a.redis,
		a.logger,
		queue.Config{
			QueueName:           a.config.Queue.QueueName,
			MaxRetries:          a.config.Queue.MaxRetries,
			RetryDelay:          a.config.Queue.RetryDelay,
			JobTimeout:          a.config.Queue.JobTimeout,
			ResultTTL:           a.config.Queue.ResultTTL,
			DefaultConcurrency:  a.config.Queue.DefaultConcurrency,
			PollInterval:        a.config.Queue.PollInterval,
			CleanupInterval:     a.config.Queue.CleanupInterval,
			StatsEnabled:        a.config.Queue.StatsEnabled,
			StatsUpdateInterval: a.config.Queue.StatsUpdateInterval,
		},
	)

	// DAOs (Data Access Objects)
	userRepository := userDao.NewPostgresRepository(a.db)
	memoryRepository := memoryDao.NewPostgresRepository(a.db)

	// Optionally use Qdrant if enabled
	if a.config.Qdrant.Enabled {
		qdrantRepo, err := memoryDao.NewQdrantRepository(
			memoryDao.QdrantConfig{
				Host:           a.config.Qdrant.Host,
				Port:           a.config.Qdrant.Port,
				CollectionName: a.config.Qdrant.CollectionName,
				VectorSize:     a.config.Qdrant.VectorSize,
				UseHTTPS:       a.config.Qdrant.UseHTTPS,
				APIKey:         a.config.Qdrant.APIKey,
			},
			memoryRepository, // fallback to PostgreSQL
		)
		if err != nil {
			a.logger.WithError(err).Warn("Failed to initialize Qdrant repository, using PostgreSQL only")
		} else {
			memoryRepository = qdrantRepo
			a.logger.Info("Qdrant vector database initialized successfully")
		}
	}

	// Register job handlers
	generateEmbeddingHandler := queue.NewGenerateEmbeddingHandler(embeddingSvc, memoryRepository, a.logger)
	batchEmbeddingHandler := queue.NewBatchEmbeddingHandler(embeddingSvc, memoryRepository, a.logger)

	a.jobQueue.RegisterHandler("generate_embedding", generateEmbeddingHandler)
	a.jobQueue.RegisterHandler("batch_embedding", batchEmbeddingHandler)

	// Start job queue with concurrency
	if err := a.jobQueue.StartConsuming(ctx, a.config.Queue.DefaultConcurrency); err != nil {
		return fmt.Errorf("failed to start job queue consumer: %w", err)
	}

	// Services
	userSvc := userService.NewService(userRepository)

	// Create regular memory service
	regularMemorySvc := memoryService.NewService(memoryRepository, userRepository, embeddingSvc, a.logger)

	// Initialize AI Memory Service if needed
	// For now, we'll use the regular service
	enhancedMemorySvc := regularMemorySvc

	// Handlers
	userHandler := userHandler.NewHandler(userSvc)
	memoryHandler := memoryHandler.NewHandler(enhancedMemorySvc, a.logger)

	// Setup router
	gin.SetMode(config.Mode)
	router := gin.New()

	// Add essential middleware
	router.Use(middleware.RequestID())
	router.Use(middleware.ResponseTime())
	router.Use(middleware.RequestLogger(a.logger))
	router.Use(middleware.Recovery(a.logger))
	router.Use(middleware.ErrorHandler(a.logger))

	// Add performance monitoring middleware
	router.Use(middleware.SlowQueryDetector(2*time.Second, a.logger))

	// Add security middleware
	router.Use(middleware.SecurityHeaders())
	router.Use(middleware.XSSProtection())
	router.Use(middleware.SQLInjectionProtection())
	router.Use(middleware.RateLimit(a.config.Security.RateLimit, 60)) // Rate limit per minute

	// CORS middleware with proper configuration
	allowedOrigins := a.config.Security.AllowedOrigins
	if len(allowedOrigins) == 0 {
		// Default allowed origins for development
		if gin.IsDebugging() {
			allowedOrigins = []string{"http://localhost:3000", "http://127.0.0.1:3000"}
		} else {
			allowedOrigins = []string{"https://yourdomain.com"}
		}
	}

	router.Use(cors.New(cors.Config{
		AllowOrigins:     allowedOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Content-Type", "Authorization", "X-API-Key", "X-Request-ID"},
		ExposeHeaders:    []string{"X-Request-ID"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Request validation middleware
	router.Use(middleware.RequestSizeLimit(10 << 20)) // 10MB limit

	// Setup routes
	a.setupRoutes(router, userHandler, memoryHandler)

	// Setup server
	a.server = &http.Server{
		Addr:         fmt.Sprintf("%s:%d", config.Host, config.Port),
		Handler:      router,
		ReadTimeout:  config.ReadTimeout,
		WriteTimeout: config.WriteTimeout,
		IdleTimeout:  config.IdleTimeout,
	}

	a.logger.WithField("address", a.server.Addr).Info("Starting HTTP server with AI components")

	return a.server.ListenAndServe()
}

// setupRoutes configures all application routes
func (a *App) setupRoutes(router *gin.Engine, userHandler *userHandler.Handler, memoryHandler *memoryHandler.Handler) {
	// Public routes
	api := router.Group("/api/v1")

	// Health check endpoint (public)
	api.GET("/health", func(c *gin.Context) {
		response.Success(c, http.StatusOK, gin.H{
			"status":      "ok",
			"time":        time.Now().UTC(),
			"version":     "1.0.0",
			"environment": a.config.Server.Mode,
		})
	})

	// Public user creation endpoint
	api.POST("/users", userHandler.CreateUser)

	// Authentication endpoints
	auth := api.Group("/auth")
	{
		auth.POST("/login", a.handleLogin)
		auth.POST("/refresh", a.handleRefreshToken)
		auth.POST("/logout", middleware.OptionalJWTAuth(a.jwtService), a.handleLogout)
	}

	// Protected routes requiring authentication
	protected := api.Group("")
	protected.Use(middleware.OptionalJWTAuth(a.jwtService)) // Allow optional JWT auth

	// User routes
	users := protected.Group("/users")
	users.Use(middleware.ValidateJSON())
	{
		users.GET("/:id", middleware.ValidateUUID("id"), userHandler.GetUser)
		users.GET("/username/:username", userHandler.GetUserByUsername)
		users.GET("/search", userHandler.GetUserByEmail) // ?email=...
		users.PUT("/:id", middleware.ValidateUUID("id"), userHandler.UpdateUser)
		users.DELETE("/:id", middleware.ValidateUUID("id"), userHandler.DeleteUser)
		users.GET("", userHandler.ListUsers)
		users.GET("/stats", userHandler.GetUserStats)
		users.POST("/:id/login", middleware.ValidateUUID("id"), userHandler.UpdateLastLogin)
	}

	// Memory routes - these should be more strictly protected in production
	memories := protected.Group("/memories")
	memories.Use(middleware.ValidateJSON())
	{
		memories.POST("", memoryHandler.CreateMemory)
		memories.GET("/:id", middleware.ValidateUUID("id"), memoryHandler.GetMemory)
		memories.PUT("/:id", middleware.ValidateUUID("id"), memoryHandler.UpdateMemory)
		memories.DELETE("/:id", middleware.ValidateUUID("id"), memoryHandler.DeleteMemory)
		memories.GET("/users/:user_id", middleware.ValidateUUID("user_id"), memoryHandler.ListUserMemories)
		memories.POST("/users/:user_id/search", middleware.ValidateUUID("user_id"), memoryHandler.SearchMemories)
		memories.GET("/users/:user_id/similar", middleware.ValidateUUID("user_id"), memoryHandler.SearchSimilarMemories)
		memories.GET("/users/:user_id/stats", middleware.ValidateUUID("user_id"), memoryHandler.GetMemoryStats)
	}

	// Admin routes - require JWT authentication and admin role
	admin := api.Group("/admin")
	admin.Use(middleware.JWTAuth(a.jwtService))
	admin.Use(middleware.RequireRole("admin", "system"))
	{
		// Example admin endpoints
		admin.GET("/system/status", func(c *gin.Context) {
			// Return system status, queue stats, etc.
			response.Success(c, http.StatusOK, gin.H{
				"status": "admin access granted",
				"server": gin.H{
					"uptime":      time.Since(time.Now()),
					"environment": a.config.Server.Mode,
				},
			})
		})
	}
}

// LoginRequest represents a login request
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse represents a login response
type LoginResponse struct {
	Success      bool   `json:"success"`
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
	RefreshToken string `json:"refresh_token,omitempty"`
	User         struct {
		ID       string `json:"id"`
		Username string `json:"username"`
		Email    string `json:"email"`
		Role     string `json:"role"`
	} `json:"user"`
}

// RefreshTokenRequest represents a refresh token request
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// handleLogin handles user authentication and JWT token generation
func (a *App) handleLogin(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid_request", "Invalid request format")
		return
	}

	// TODO: Implement actual user authentication
	// For now, use a simple demonstration authentication
	if req.Username == "demo" && req.Password == "demo123" {
		userID := uuid.New()

		// Generate JWT token
		token, err := a.jwtService.GenerateToken(userID, req.Username, "demo@example.com", "user")
		if err != nil {
			a.logger.WithError(err).Error("Failed to generate JWT token")
			response.InternalError(c, "Failed to generate authentication token")
			return
		}

		loginResp := LoginResponse{
			Success:     true,
			AccessToken: token,
			TokenType:   "Bearer",
			ExpiresIn:   int64(a.config.Security.JWTExpiry.Seconds()),
		}
		loginResp.User.ID = userID.String()
		loginResp.User.Username = req.Username
		loginResp.User.Email = "demo@example.com"
		loginResp.User.Role = "user"

		response.Success(c, http.StatusOK, loginResp)
		return
	}

	// Authentication failed
	response.Unauthorized(c, "Invalid credentials")
}

// handleRefreshToken handles JWT token refresh
func (a *App) handleRefreshToken(c *gin.Context) {
	var req RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid_request", "Invalid request format")
		return
	}

	// Refresh the token
	newToken, err := a.jwtService.RefreshToken(req.RefreshToken)
	if err != nil {
		response.Unauthorized(c, err.Error())
		return
	}

	// Get claims from new token to return user info
	claims, err := a.jwtService.ValidateToken(newToken)
	if err != nil {
		a.logger.WithError(err).Error("Failed to validate refreshed token")
		response.InternalError(c, "Token validation failed")
		return
	}

	refreshResp := LoginResponse{
		Success:     true,
		AccessToken: newToken,
		TokenType:   "Bearer",
		ExpiresIn:   int64(a.config.Security.JWTExpiry.Seconds()),
	}
	refreshResp.User.ID = claims.UserID.String()
	refreshResp.User.Username = claims.Username
	refreshResp.User.Email = claims.Email
	refreshResp.User.Role = claims.Role

	response.Success(c, http.StatusOK, refreshResp)
}

// handleLogout handles user logout (token invalidation)
func (a *App) handleLogout(c *gin.Context) {
	// For stateless JWT, logout is handled client-side by removing the token
	// In a production system, you might want to implement a token blacklist
	response.Success(c, http.StatusOK, gin.H{
		"message": "Successfully logged out",
	})
}

// loadAPIKeysFromConfig loads API keys from configuration or environment variables
func (a *App) loadAPIKeysFromConfig() map[string]string {
	keys := make(map[string]string)

	// Try to load from environment variable first
	if envKeys := os.Getenv("API_KEYS"); envKeys != "" {
		// Expected format: "key1:user1,key2:user2"
		pairs := strings.Split(envKeys, ",")
		for _, pair := range pairs {
			kv := strings.SplitN(pair, ":", 2)
			if len(kv) == 2 {
				keys[strings.TrimSpace(kv[0])] = strings.TrimSpace(kv[1])
			}
		}
	}

	// Fallback to config file or default development key
	if len(keys) == 0 {
		a.logger.Warn("No API keys found in environment variables, using development fallback")
		if gin.IsDebugging() {
			// Only use development key in debug mode
			if devKey := os.Getenv("DEV_API_KEY"); devKey != "" {
				keys[devKey] = "system"
			} else {
				a.logger.Error("No API keys configured and DEV_API_KEY not set")
			}
		} else {
			a.logger.Error("No API keys configured for production mode")
		}
	}

	return keys
}

// Shutdown gracefully shuts down the application
func (a *App) Shutdown(ctx context.Context) error {
	a.logger.Info("Shutting down server...")

	// Stop job queue first
	if a.jobQueue != nil {
		if err := a.jobQueue.StopConsuming(); err != nil {
			a.logger.WithError(err).Error("Failed to stop job queue")
		}
		if err := a.jobQueue.Close(); err != nil {
			a.logger.WithError(err).Error("Failed to close job queue")
		}
	}

	// Shutdown HTTP server
	if a.server != nil {
		return a.server.Shutdown(ctx)
	}

	return nil
}
