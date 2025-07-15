package config

import (
	"time"

	"github.com/spf13/viper"
)

// Config represents the main application configuration structure
// It aggregates all configuration sections including server, database, and app settings
// This struct is populated from environment variables, config files, or defaults
type Config struct {
	Server   ServerConfig   `mapstructure:"server"`   // HTTP server configuration settings
	Database DatabaseConfig `mapstructure:"database"` // Database connection and pool settings
	Redis    RedisConfig    `mapstructure:"redis"`    // Redis cache configuration settings
	App      AppConfig      `mapstructure:"app"`      // Application metadata and general settings
}

// ServerConfig configures the HTTP server behavior and timeouts
// These settings directly affect server performance and client experience
// Similar to Spring Boot's server.* properties in application.properties
type ServerConfig struct {
	Port         string        `mapstructure:"port"`          // HTTP server port (default: "8080")
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`  // Max time to read request (default: 15s)
	WriteTimeout time.Duration `mapstructure:"write_timeout"` // Max time to write response (default: 15s)
	IdleTimeout  time.Duration `mapstructure:"idle_timeout"`  // Max time for idle keep-alive connections (default: 60s)
}

// DatabaseConfig manages database connection pool settings
// These settings are critical for database performance and resource management
// Similar to Spring Boot's spring.datasource.* properties
type DatabaseConfig struct {
	URL             string        `mapstructure:"url"`               // Database connection string (PostgreSQL format)
	MaxOpenConns    int           `mapstructure:"max_open_conns"`    // Maximum number of open connections (default: 25)
	MaxIdleConns    int           `mapstructure:"max_idle_conns"`    // Maximum number of idle connections (default: 25)
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"` // Maximum connection lifetime (default: 5m)
}

// RedisConfig manages Redis connection and caching settings
// These settings control cache behavior and Redis connection parameters
// Similar to Spring Boot's spring.redis.* properties
type RedisConfig struct {
	Host         string        `mapstructure:"host"`           // Redis server host (default: "localhost")
	Port         string        `mapstructure:"port"`           // Redis server port (default: "6379")
	Password     string        `mapstructure:"password"`       // Redis password (optional, default: "")
	DB           int           `mapstructure:"db"`             // Redis database number (default: 0)
	PoolSize     int           `mapstructure:"pool_size"`      // Connection pool size (default: 10)
	MinIdleConns int           `mapstructure:"min_idle_conns"` // Minimum idle connections (default: 5)
	CacheTTL     time.Duration `mapstructure:"cache_ttl"`      // Default cache TTL for events (default: 5m)
}

// AppConfig contains application-level metadata and general settings
// These settings control application behavior and operational characteristics
// Similar to Spring Boot's spring.application.* properties
type AppConfig struct {
	Name        string `mapstructure:"name"`        // Application name for logging and monitoring (default: "enterprise-crud")
	Version     string `mapstructure:"version"`     // Application version for health checks and monitoring (default: "1.0.0")
	Environment string `mapstructure:"environment"` // Runtime environment: development, staging, production (default: "development")
	LogLevel    string `mapstructure:"log_level"`   // Logging level: debug, info, warn, error (default: "info")
}

// Load initializes and returns the application configuration
// It loads configuration from multiple sources in this priority order:
// 1. Default values (always applied first)
// 2. Config files (config.yaml from current dir, ./configs, or /etc/enterprise-crud)
// 3. Environment variables (prefixed with APP_, e.g., APP_SERVER_PORT)
//
// This is similar to Spring Boot's configuration loading mechanism
func Load() (*Config, error) {
	v := viper.New()

	// Set default values
	setDefaults(v)

	// Environment variable configuration
	v.SetEnvPrefix("APP")
	v.AutomaticEnv()

	// Config file configuration
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(".")
	v.AddConfigPath("./configs")
	v.AddConfigPath("/etc/enterprise-crud")

	// Read config file if it exists
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, err
		}
	}

	var config Config
	if err := v.Unmarshal(&config); err != nil {
		return nil, err
	}

	return &config, nil
}

// setDefaults configures default values for all configuration options
// These defaults ensure the application can run without external configuration
// Similar to Spring Boot's @ConfigurationProperties with default values
func setDefaults(v *viper.Viper) {
	// Server defaults
	v.SetDefault("server.port", "8080")
	v.SetDefault("server.read_timeout", "15s")
	v.SetDefault("server.write_timeout", "15s")
	v.SetDefault("server.idle_timeout", "60s")

	// Database defaults
	v.SetDefault("database.url", "postgres://postgres:postgres@localhost:5433/enterprise_crud?sslmode=disable")
	v.SetDefault("database.max_open_conns", 25)
	v.SetDefault("database.max_idle_conns", 25)
	v.SetDefault("database.conn_max_lifetime", "5m")

	// Redis defaults
	v.SetDefault("redis.host", "localhost")
	v.SetDefault("redis.port", "6379")
	v.SetDefault("redis.password", "")
	v.SetDefault("redis.db", 0)
	v.SetDefault("redis.pool_size", 10)
	v.SetDefault("redis.min_idle_conns", 5)
	v.SetDefault("redis.cache_ttl", "5m")

	// App defaults
	v.SetDefault("app.name", "enterprise-crud")
	v.SetDefault("app.version", "1.0.0")
	v.SetDefault("app.environment", "development")
	v.SetDefault("app.log_level", "info")
}
