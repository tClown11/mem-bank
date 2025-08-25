package app

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	memoryDao "mem_bank/internal/dao/memory"
	userDao "mem_bank/internal/dao/user"
	memoryHandler "mem_bank/internal/handler/http/memory"
	userHandler "mem_bank/internal/handler/http/user"
	memoryService "mem_bank/internal/service/memory"
	userService "mem_bank/internal/service/user"
	"mem_bank/pkg/logger"
)

// App represents the application
type App struct {
	server *http.Server
	db     *gorm.DB
	logger logger.Logger
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
func NewApp(db *gorm.DB, appLogger *logger.Logger, config Config) *App {
	return &App{
		db:     db,
		logger: *appLogger,
	}
}

// Start starts the application
func (a *App) Start(ctx context.Context, config Config) error {
	// Wire up dependencies using constructor-based dependency injection

	// DAOs (Data Access Objects)
	userRepository := userDao.NewPostgresRepository(a.db)
	memoryRepository := memoryDao.NewPostgresRepository(a.db)

	// Services
	userSvc := userService.NewService(userRepository)
	memorySvc := memoryService.NewService(memoryRepository, userRepository)

	// Handlers
	userHandler := userHandler.NewHandler(userSvc)
	memoryHandler := memoryHandler.NewHandler(memorySvc)

	// Setup router
	gin.SetMode(config.Mode)
	router := gin.New()

	// Add middleware
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

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

	a.logger.WithField("address", a.server.Addr).Info("Starting HTTP server")

	return a.server.ListenAndServe()
}

// setupRoutes configures all application routes
func (a *App) setupRoutes(router *gin.Engine, userHandler *userHandler.Handler, memoryHandler *memoryHandler.Handler) {
	api := router.Group("/api/v1")

	// Health check endpoint
	api.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
			"time":   time.Now().UTC(),
		})
	})

	// User routes
	users := api.Group("/users")
	{
		users.POST("", userHandler.CreateUser)
		users.GET("/:id", userHandler.GetUser)
		users.GET("/username/:username", userHandler.GetUserByUsername)
		users.GET("/search", userHandler.GetUserByEmail) // ?email=...
		users.PUT("/:id", userHandler.UpdateUser)
		users.DELETE("/:id", userHandler.DeleteUser)
		users.GET("", userHandler.ListUsers)
		users.GET("/stats", userHandler.GetUserStats)
		users.POST("/:id/login", userHandler.UpdateLastLogin)
	}

	// Memory routes
	memories := api.Group("/memories")
	{
		memories.POST("", memoryHandler.CreateMemory)
		memories.GET("/:id", memoryHandler.GetMemory)
		memories.PUT("/:id", memoryHandler.UpdateMemory)
		memories.DELETE("/:id", memoryHandler.DeleteMemory)
		memories.GET("/users/:user_id", memoryHandler.ListUserMemories)
		memories.POST("/users/:user_id/search", memoryHandler.SearchMemories)
		memories.GET("/users/:user_id/similar", memoryHandler.SearchSimilarMemories)
		memories.GET("/users/:user_id/stats", memoryHandler.GetMemoryStats)
	}
}

// Shutdown gracefully shuts down the application
func (a *App) Shutdown(ctx context.Context) error {
	a.logger.Info("Shutting down server...")

	if a.server != nil {
		return a.server.Shutdown(ctx)
	}

	return nil
}
