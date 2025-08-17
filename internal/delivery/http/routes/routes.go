package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"mem_bank/internal/delivery/http/handlers"
	"mem_bank/internal/delivery/http/middleware"
)

func SetupRoutes(
	router *gin.Engine,
	logger *logrus.Logger,
	memoryHandler *handlers.MemoryHandler,
	userHandler *handlers.UserHandler,
) {
	router.Use(middleware.CORS())
	router.Use(middleware.Logger(logger))
	router.Use(gin.Recovery())

	api := router.Group("/api/v1")
	{
		health := api.Group("/health")
		{
			health.GET("", func(c *gin.Context) {
				c.JSON(200, gin.H{
					"status": "ok",
					"service": "mem_bank",
					"version": "1.0.0",
				})
			})
		}

		users := api.Group("/users")
		{
			users.POST("", userHandler.CreateUser)
			users.GET("", userHandler.ListUsers)
			users.GET("/stats", userHandler.GetUserStats)
			users.GET("/search", userHandler.GetUserByEmail)
			users.GET("/:id", userHandler.GetUser)
			users.PUT("/:id", userHandler.UpdateUser)
			users.DELETE("/:id", userHandler.DeleteUser)
			users.POST("/:id/login", userHandler.UpdateLastLogin)
			users.PUT("/:id/settings", userHandler.UpdateUserSettings)
			users.PUT("/:id/profile", userHandler.UpdateUserProfile)
			users.GET("/username/:username", userHandler.GetUserByUsername)
		}

		memories := api.Group("/memories")
		{
			memories.POST("", memoryHandler.CreateMemory)
			memories.GET("/:id", memoryHandler.GetMemory)
			memories.PUT("/:id", memoryHandler.UpdateMemory)
			memories.DELETE("/:id", memoryHandler.DeleteMemory)
			memories.GET("/user/:user_id", memoryHandler.GetUserMemories)
			memories.GET("/user/:user_id/stats", memoryHandler.GetMemoryStats)
			memories.POST("/search/similar", memoryHandler.SearchSimilar)
			memories.GET("/search/content", memoryHandler.SearchByContent)
			memories.POST("/search/tags", memoryHandler.SearchByTags)
		}
	}
}