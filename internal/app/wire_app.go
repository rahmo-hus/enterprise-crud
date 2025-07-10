package app

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"enterprise-crud/internal/config"
	"enterprise-crud/internal/domain/user"
	"enterprise-crud/internal/infrastructure/database"
	httpHandlers "enterprise-crud/internal/presentation/http"
	"github.com/gin-gonic/gin"
	"github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	_ "enterprise-crud/docs"
)

// WireApp represents the application with Wire-injected dependencies
type WireApp struct {
	config      *config.Config
	server      *http.Server
	dbConn      *database.Connection
	userHandler *httpHandlers.UserHandler
}

// NewWireApp creates a new application with injected dependencies
func NewWireApp(
	cfg *config.Config,
	dbConn *database.Connection,
	userHandler *httpHandlers.UserHandler,
) *WireApp {
	return &WireApp{
		config:      cfg,
		dbConn:      dbConn,
		userHandler: userHandler,
	}
}

// Run starts the application with graceful shutdown
func (a *WireApp) Run() error {
	// Configure database connection pool
	sqlDB, err := a.dbConn.DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get sql.DB from GORM: %w", err)
	}

	sqlDB.SetMaxOpenConns(a.config.Database.MaxOpenConns)
	sqlDB.SetMaxIdleConns(a.config.Database.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(a.config.Database.ConnMaxLifetime)

	// Setup HTTP server
	router := a.setupRouter()
	
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

func (a *WireApp) setupRouter() *gin.Engine {
	if a.config.App.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// Health check endpoint
	// @Summary Health check endpoint
	// @Description Check if the service is running
	// @Tags health
	// @Produce json
	// @Success 200 {object} map[string]interface{} "Service is healthy"
	// @Router /health [get]
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
		a.userHandler.RegisterRoutes(v1)
	}

	return router
}

func (a *WireApp) waitForShutdown() error {
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

	// Close database connection
	if a.dbConn != nil {
		a.dbConn.Close()
	}

	log.Println("Server exited")
	return nil
}

// Dependencies injection interface
type Dependencies struct {
	Config      *config.Config
	DBConn      *database.Connection
	UserRepo    user.Repository
	UserService user.Service
	UserHandler *httpHandlers.UserHandler
}

// NewDependencies creates all application dependencies manually (alternative to Wire)
func NewDependencies(cfg *config.Config) (*Dependencies, error) {
	// Database connection
	dbConn, err := database.NewConnection()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Repositories
	userRepo := database.NewUserRepository(dbConn.DB)

	// Services
	userService := user.NewUserService(userRepo)

	// Handlers
	userHandler := httpHandlers.NewUserHandler(userService)

	return &Dependencies{
		Config:      cfg,
		DBConn:      dbConn,
		UserRepo:    userRepo,
		UserService: userService,
		UserHandler: userHandler,
	}, nil
}