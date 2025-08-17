package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"mem_bank/configs"
	"mem_bank/internal/delivery/http/handlers"
	"mem_bank/internal/delivery/http/routes"
	"mem_bank/internal/repository"
	"mem_bank/internal/usecase"
	"mem_bank/pkg/database"
	"mem_bank/pkg/logger"
)

func main() {
	config, err := configs.LoadConfig("")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	appLogger, err := logger.NewLogger(&config.Logging)
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}

	db, err := database.NewGormConnection(&config.Database)
	if err != nil {
		appLogger.WithError(err).Fatal("Failed to connect to database")
	}
	defer db.Close()

	appLogger.Info("Database connection established")

	memoryRepo := repository.NewMemoryRepository(db.DB)
	userRepo := repository.NewUserRepository(db.DB)

	memoryUsecase := usecase.NewMemoryUsecase(memoryRepo, userRepo)
	userUsecase := usecase.NewUserUsecase(userRepo, memoryRepo)

	memoryHandler := handlers.NewMemoryHandler(memoryUsecase)
	userHandler := handlers.NewUserHandler(userUsecase)

	gin.SetMode(config.Server.Mode)
	router := gin.New()

	routes.SetupRoutes(router, appLogger.Logger, memoryHandler, userHandler)

	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", config.Server.Host, config.Server.Port),
		Handler:      router,
		ReadTimeout:  config.Server.ReadTimeout,
		WriteTimeout: config.Server.WriteTimeout,
		IdleTimeout:  config.Server.IdleTimeout,
	}

	go func() {
		appLogger.WithField("address", server.Addr).Info("Starting HTTP server")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			appLogger.WithError(err).Fatal("Failed to start server")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	appLogger.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		appLogger.WithError(err).Fatal("Server forced to shutdown")
	}

	appLogger.Info("Server exiting")
}