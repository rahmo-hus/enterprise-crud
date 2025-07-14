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
	"enterprise-crud/internal/domain/order"
	"enterprise-crud/internal/domain/role"
	"enterprise-crud/internal/domain/user"
	"enterprise-crud/internal/domain/venue"
	"enterprise-crud/internal/infrastructure/auth"
	"enterprise-crud/internal/infrastructure/database"
	httpHandlers "enterprise-crud/internal/presentation/http"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// WireApp represents the application with Wire-injected dependencies
type WireApp struct {
	config       *config.Config
	server       *http.Server
	dbConn       *database.Connection
	userHandler  *httpHandlers.UserHandler
	eventHandler *httpHandlers.EventHandler
	orderHandler *httpHandlers.OrderHandler
	venueHandler *httpHandlers.VenueHandler
}

// NewWireApp creates a new application with injected dependencies
func NewWireApp(
	cfg *config.Config,
	dbConn *database.Connection,
	userHandler *httpHandlers.UserHandler,
	eventHandler *httpHandlers.EventHandler,
	orderHandler *httpHandlers.OrderHandler,
	venueHandler *httpHandlers.VenueHandler,
) *WireApp {
	return &WireApp{
		config:       cfg,
		dbConn:       dbConn,
		userHandler:  userHandler,
		eventHandler: eventHandler,
		orderHandler: orderHandler,
		venueHandler: venueHandler,
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
	router := a.SetupRouter()

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

// SetupRouter creates and configures the HTTP router
func (a *WireApp) SetupRouter() *gin.Engine {
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
		a.userHandler.RegisterAuthRoutes(v1)
		a.eventHandler.RegisterRoutes(v1)
		a.orderHandler.RegisterRoutes(v1)
		a.venueHandler.RegisterRoutes(v1)
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
	Config       *config.Config
	DBConn       *database.Connection
	UserRepo     user.Repository
	RoleRepo     role.Repository
	UserService  user.Service
	EventService event.Service
	OrderService order.Service
	VenueService venue.Service
	JWTService   *auth.JWTService
	UserHandler  *httpHandlers.UserHandler
	EventHandler *httpHandlers.EventHandler
	OrderHandler *httpHandlers.OrderHandler
	VenueHandler *httpHandlers.VenueHandler
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
	roleRepo := database.NewRoleRepository(dbConn.DB)
	venueRepo := database.NewVenueRepository(dbConn.DB)
	eventRepo := database.NewEventRepository(dbConn.DB)
	orderRepo := database.NewOrderRepository(dbConn.DB)

	// Services
	userService := user.NewUserService(userRepo, roleRepo)
	venueService := venue.NewVenueService(venueRepo)
	eventService := event.NewService(eventRepo, venueRepo)
	orderService := order.NewOrderService(orderRepo, dbConn.DB)

	// JWT Service
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "default-secret-key-change-in-production"
	}

	jwtIssuer := os.Getenv("JWT_ISSUER")
	if jwtIssuer == "" {
		jwtIssuer = "enterprise-crud-api"
	}

	jwtExpirationHours := 720 // 30 days default
	if envHours := os.Getenv("JWT_EXPIRATION_HOURS"); envHours != "" {
		if hours, err := strconv.Atoi(envHours); err == nil {
			jwtExpirationHours = hours
		}
	}

	jwtService := auth.NewJWTService(jwtSecret, jwtIssuer, time.Duration(jwtExpirationHours)*time.Hour)

	// Handlers
	userHandler := httpHandlers.NewUserHandler(userService, jwtService)
	eventHandler := httpHandlers.NewEventHandler(eventService, jwtService)
	orderHandler := httpHandlers.NewOrderHandler(orderService, jwtService)
	venueHandler := httpHandlers.NewVenueHandler(venueService, jwtService)

	return &Dependencies{
		Config:       cfg,
		DBConn:       dbConn,
		UserRepo:     userRepo,
		RoleRepo:     roleRepo,
		UserService:  userService,
		EventService: eventService,
		OrderService: orderService,
		VenueService: venueService,
		JWTService:   jwtService,
		UserHandler:  userHandler,
		EventHandler: eventHandler,
		OrderHandler: orderHandler,
		VenueHandler: venueHandler,
	}, nil
}
