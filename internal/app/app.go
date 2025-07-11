package app

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	_ "enterprise-crud/docs"
	"enterprise-crud/internal/config"
	"enterprise-crud/internal/domain/event"
	"enterprise-crud/internal/domain/user"
	"enterprise-crud/internal/infrastructure/auth"
	"enterprise-crud/internal/infrastructure/database"
	httpHandlers "enterprise-crud/internal/presentation/http"
	"github.com/gin-gonic/gin"
	"github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type App struct {
	config *config.Config
	server *http.Server
}

func New(cfg *config.Config) *App {
	return &App{
		config: cfg,
	}
}

func (a *App) Run() error {
	// Setup database connection
	dbConn, err := database.NewConnection()
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer dbConn.Close()

	// Configure database connection pool
	sqlDB, err := dbConn.DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get sql.DB from GORM: %w", err)
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "default-secret-key-change-in-production"
	}

	jwtIssuer := os.Getenv("JWT_ISSUER")
	if jwtIssuer == "" {
		jwtIssuer = "enterprise-crud-api"
	}

	jwtExpirationHours := 720
	if envHours := os.Getenv("JWT_EXPIRATION_HOURS"); envHours != "" {
		if hours, err := strconv.Atoi(envHours); err == nil {
			jwtExpirationHours = hours
		}
	}

	sqlDB.SetMaxOpenConns(a.config.Database.MaxOpenConns)
	sqlDB.SetMaxIdleConns(a.config.Database.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(a.config.Database.ConnMaxLifetime)

	// Initialize dependencies
	userRepo := database.NewUserRepository(dbConn.DB)
	roleRepo := database.NewRoleRepository(dbConn.DB)
	venueRepo := database.NewVenueRepository(dbConn.DB)
	eventRepo := database.NewEventRepository(dbConn.DB)

	userService := user.NewUserService(userRepo, roleRepo)
	eventService := event.NewService(eventRepo, venueRepo)

	jwtService := auth.NewJWTService(jwtSecret, jwtIssuer, time.Duration(jwtExpirationHours)*time.Hour)
	userHandler := httpHandlers.NewUserHandler(userService, jwtService)
	eventHandler := httpHandlers.NewEventHandler(eventService, jwtService)

	// Setup HTTP server
	router := a.setupRouter(userHandler, eventHandler)

	a.server = &http.Server{
		Addr:         ":" + a.config.Server.Port,
		Handler:      router,
		ReadTimeout:  a.config.Server.ReadTimeout,
		WriteTimeout: a.config.Server.WriteTimeout,
		IdleTimeout:  a.config.Server.IdleTimeout,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Server starting on port %s", a.config.Server.Port)
		if err := a.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown
	return a.waitForShutdown()
}

func (a *App) setupRouter(userHandler *httpHandlers.UserHandler, eventHandler *httpHandlers.EventHandler) *gin.Engine {
	if a.config.App.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":      "healthy",
			"service":     a.config.App.Name,
			"version":     a.config.App.Version,
			"environment": a.config.App.Environment,
		})
	})

	// Swagger documentation
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// API routes
	v1 := router.Group("/api/v1")
	{
		userHandler.RegisterRoutes(v1)
		userHandler.RegisterAuthRoutes(v1)
		eventHandler.RegisterRoutes(v1)
	}

	return router
}

func (a *App) waitForShutdown() error {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit
	log.Println("Shutting down server...")

	// Create a context with timeout for graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Shutdown the server gracefully
	if err := a.server.Shutdown(ctx); err != nil {
		return fmt.Errorf("server forced to shutdown: %w", err)
	}

	log.Println("Server exited")
	return nil
}
