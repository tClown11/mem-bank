package app

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
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
	"mem_bank/pkg/llm"
	"mem_bank/pkg/logger"
)

// App represents the application
type App struct {
	server      *http.Server
	db          *gorm.DB
	redis       *redis.Client
	logger      logger.Logger
	config      *configs.Config
	jobQueue    queue.Queue
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
	
	// Initialize LLM Provider
	llmConfig := &llm.Config{
		Provider:         a.config.LLM.Provider,
		APIKey:           a.config.LLM.APIKey,
		BaseURL:          a.config.LLM.BaseURL,
		EmbeddingModel:   a.config.LLM.EmbeddingModel,
		CompletionModel:  a.config.LLM.CompletionModel,
		TimeoutSeconds:   a.config.LLM.TimeoutSeconds,
		MaxRetries:       a.config.LLM.MaxRetries,
		RateLimit:        a.config.LLM.RateLimit,
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
	
	// Initialize AI Memory Service
	aiMemorySvc := memoryService.NewAIService(
		memoryRepository,
		userRepository,
		embeddingSvc,
		a.jobQueue,
		a.logger,
		memoryService.AIServiceConfig{
			AsyncEmbedding:             true,
			DefaultSimilarityThreshold: 0.8,
			EmbeddingJobPriority:       5,
		},
	)
	
	// Use AI service as the enhanced memory service
	enhancedMemorySvc := aiMemorySvc

	// Handlers
	userHandler := userHandler.NewHandler(userSvc)
	memoryHandler := memoryHandler.NewHandler(enhancedMemorySvc)

	// Setup router
	gin.SetMode(config.Mode)
	router := gin.New()

	// Add essential middleware
	router.Use(middleware.RequestID())
	router.Use(middleware.RequestLogger(a.logger))
	router.Use(middleware.Recovery(a.logger))
	router.Use(middleware.ErrorHandler(a.logger))

	// CORS middleware
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"}, // Configure this properly for production
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"*"},
		ExposeHeaders:    []string{"X-Request-ID"},
		AllowCredentials: false,
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
		c.JSON(http.StatusOK, gin.H{
			"status":      "ok",
			"time":        time.Now().UTC(),
			"version":     "1.0.0",
			"environment": a.config.Server.Mode,
		})
	})

	// Public user creation endpoint
	api.POST("/users", userHandler.CreateUser)

	// Protected routes requiring authentication
	// For now, using simple API key auth - can be extended to JWT/OAuth
	validApiKeys := map[string]string{
		"dev-api-key-123": "system", // Example API key for development
	}
	
	protected := api.Group("")
	protected.Use(middleware.OptionalAuth(validApiKeys)) // Allow optional auth for some endpoints
	
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

	// Admin routes (if needed in the future)
	admin := api.Group("/admin")
	admin.Use(middleware.ApiKeyAuth(validApiKeys))
	{
		// Example admin endpoints
		admin.GET("/system/status", func(c *gin.Context) {
			// Return system status, queue stats, etc.
			c.JSON(http.StatusOK, gin.H{
				"status": "admin access granted",
			})
		})
	}
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
