//go:build integration
// +build integration

package integration

import (
	"os"
	"strconv"
	"time"

	"enterprise-crud/internal/app"
	"enterprise-crud/internal/config"
	"enterprise-crud/internal/domain/event"
	"enterprise-crud/internal/domain/order"
	"enterprise-crud/internal/domain/user"
	"enterprise-crud/internal/infrastructure/auth"
	"enterprise-crud/internal/infrastructure/database"
	httpHandlers "enterprise-crud/internal/presentation/http"
)

// CreateTestDependencies creates test dependencies with test database
func CreateTestDependencies(cfg *config.Config, dbConn *database.Connection) (*app.Dependencies, error) {
	// Create repositories using the test database
	userRepo := database.NewUserRepository(dbConn.DB)
	roleRepo := database.NewRoleRepository(dbConn.DB)
	venueRepo := database.NewVenueRepository(dbConn.DB)
	eventRepo := database.NewEventRepository(dbConn.DB)
	orderRepo := database.NewOrderRepository(dbConn.DB)

	// Create services
	userService := user.NewUserService(userRepo, roleRepo)
	eventService := event.NewService(eventRepo, venueRepo)
	orderService := order.NewOrderService(orderRepo, dbConn.DB)

	// JWT Service
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "test-secret-key"
	}

	jwtIssuer := os.Getenv("JWT_ISSUER")
	if jwtIssuer == "" {
		jwtIssuer = "test-issuer"
	}

	jwtExpirationHours := 720 // 30 days default
	if envHours := os.Getenv("JWT_EXPIRATION_HOURS"); envHours != "" {
		if hours, err := strconv.Atoi(envHours); err == nil {
			jwtExpirationHours = hours
		}
	}

	jwtService := auth.NewJWTService(jwtSecret, jwtIssuer, time.Duration(jwtExpirationHours)*time.Hour)

	// Create handlers
	userHandler := httpHandlers.NewUserHandler(userService, jwtService)
	eventHandler := httpHandlers.NewEventHandler(eventService, jwtService)
	orderHandler := httpHandlers.NewOrderHandler(orderService, jwtService)

	return &app.Dependencies{
		Config:       cfg,
		DBConn:       dbConn,
		UserRepo:     userRepo,
		RoleRepo:     roleRepo,
		UserService:  userService,
		EventService: eventService,
		OrderService: orderService,
		JWTService:   jwtService,
		UserHandler:  userHandler,
		EventHandler: eventHandler,
		OrderHandler: orderHandler,
	}, nil
}

// CreateTestConfig creates a test configuration
func CreateTestConfig() *config.Config {
	return &config.Config{
		App: config.AppConfig{
			Name:        "test-app",
			Version:     "1.0.0",
			Environment: "test",
		},
		Server: config.ServerConfig{
			Port:         "8080",
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 10 * time.Second,
			IdleTimeout:  120 * time.Second,
		},
		Database: config.DatabaseConfig{
			MaxOpenConns:    10,
			MaxIdleConns:    5,
			ConnMaxLifetime: time.Hour,
		},
	}
}
