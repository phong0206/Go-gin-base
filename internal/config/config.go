package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/phong0206/go_gin_restful/internal/common/constants"
)

type Config struct {
	App      AppConfig
	HTTP     HTTPConfig
	CORS     CORSConfig
	Cookie   CookieConfig
	Database DatabaseConfig
	JWT      JWTConfig
	Redis    RedisConfig
}

type AppConfig struct {
	Name        string
	Environment string
}

type HTTPConfig struct {
	Host       string
	Port       int
	APIPrefix  string
	StorageDir string
}

func (c HTTPConfig) Address() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

type CORSConfig struct {
	AllowedOrigins   []string
	AllowCredentials bool
}

type CookieConfig struct {
	Domain string
	Secure bool
}

type DatabaseConfig struct {
	Host            string
	Port            int
	Name            string
	User            string
	Password        string
	SSLMode         string
	URL             string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
	ConnectTimeout  time.Duration
}

type JWTConfig struct {
	AccessSecret  string
	RefreshSecret string
	AccessTTL     time.Duration
	RefreshTTL    time.Duration
}

type RedisConfig struct {
	Host     string
	Port     int
	User     string
	Password string
}

// Load reads application configuration from environment variables.
func Load() (*Config, error) {
	httpPort, err := envInt("HTTP_PORT", 8080)
	if err != nil {
		return nil, err
	}
	databasePort, err := envInt("DATABASE_PORT", 5432)
	if err != nil {
		return nil, err
	}
	accessTTL, err := envDuration("JWT_ACCESS_TTL", 15*time.Minute)
	if err != nil {
		return nil, err
	}
	refreshTTL, err := envDuration("JWT_REFRESH_TTL", 7*24*time.Hour)
	if err != nil {
		return nil, err
	}
	maxOpenConns, err := envInt("DATABASE_MAX_OPEN_CONNS", 25)
	if err != nil {
		return nil, err
	}
	maxIdleConns, err := envInt("DATABASE_MAX_IDLE_CONNS", 5)
	if err != nil {
		return nil, err
	}
	connMaxLifetime, err := envDuration("DATABASE_CONN_MAX_LIFETIME", 30*time.Minute)
	if err != nil {
		return nil, err
	}
	connMaxIdleTime, err := envDuration("DATABASE_CONN_MAX_IDLE_TIME", 5*time.Minute)
	if err != nil {
		return nil, err
	}
	connectTimeout, err := envDuration("DATABASE_CONNECT_TIMEOUT", 5*time.Second)
	if err != nil {
		return nil, err
	}
	redisPort, err := envInt("REDIS_PORT", 6379)
	if err != nil {
		return nil, err
	}

	cfg := &Config{
		App: AppConfig{
			Name:        envString("APP_NAME", "ticket-service"),
			Environment: envString("APP_ENV", "local"),
		},
		HTTP: HTTPConfig{
			Host:       envString("HTTP_HOST", "0.0.0.0"),
			Port:       httpPort,
			APIPrefix:  envString("API_PREFIX", constants.DefaultAPIPrefix),
			StorageDir: envString("STORAGE_DIR", constants.DefaultStorageDir),
		},
		CORS: CORSConfig{
			AllowedOrigins:   envCSV("CORS_ALLOWED_ORIGINS", []string{"http://localhost:3000", "http://localhost:5173"}),
			AllowCredentials: envBool("CORS_ALLOW_CREDENTIALS", true),
		},
		Cookie: CookieConfig{
			Domain: envString("COOKIE_DOMAIN", ""),
			Secure: envBool("COOKIE_SECURE", false),
		},
		Database: DatabaseConfig{
			Host:            envString("DATABASE_HOST", "localhost"),
			Port:            databasePort,
			Name:            envString("DATABASE_NAME", "ticket_service"),
			User:            envString("DATABASE_USER", "postgres"),
			Password:        envString("DATABASE_PASSWORD", "postgres"),
			SSLMode:         envString("DATABASE_SSLMODE", "disable"),
			URL:             os.Getenv("DATABASE_URL"),
			MaxOpenConns:    maxOpenConns,
			MaxIdleConns:    maxIdleConns,
			ConnMaxLifetime: connMaxLifetime,
			ConnMaxIdleTime: connMaxIdleTime,
			ConnectTimeout:  connectTimeout,
		},
		JWT: JWTConfig{
			AccessSecret:  envString("JWT_ACCESS_SECRET", "local-access-secret-change-me"),
			RefreshSecret: envString("JWT_REFRESH_SECRET", "local-refresh-secret-change-me"),
			AccessTTL:     accessTTL,
			RefreshTTL:    refreshTTL,
		},
		Redis: RedisConfig{
			Host:     envString("REDIS_HOST", "localhost"),
			Port:     redisPort,
			User:     envString("REDIS_USER", ""),
			Password: envString("REDIS_PASSWORD", ""),
		},
	}

	if cfg.Database.URL == "" {
		cfg.Database.URL = fmt.Sprintf(
			"postgres://%s:%s@%s:%d/%s?sslmode=%s",
			cfg.Database.User,
			cfg.Database.Password,
			cfg.Database.Host,
			cfg.Database.Port,
			cfg.Database.Name,
			cfg.Database.SSLMode,
		)
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (c *Config) validate() error {
	if c.App.Name == "" {
		return fmt.Errorf("APP_NAME must not be empty")
	}
	if c.HTTP.Host == "" {
		return fmt.Errorf("HTTP_HOST must not be empty")
	}
	if c.HTTP.Port < 1 || c.HTTP.Port > 65535 {
		return fmt.Errorf("HTTP_PORT must be between 1 and 65535")
	}
	if c.Database.Port < 1 || c.Database.Port > 65535 {
		return fmt.Errorf("DATABASE_PORT must be between 1 and 65535")
	}
	if c.Database.MaxOpenConns < 1 || c.Database.MaxIdleConns < 0 || c.Database.MaxIdleConns > c.Database.MaxOpenConns {
		return fmt.Errorf("database pool settings are invalid")
	}
	if c.Database.ConnMaxLifetime <= 0 || c.Database.ConnMaxIdleTime <= 0 || c.Database.ConnectTimeout <= 0 {
		return fmt.Errorf("database pool durations must be greater than zero")
	}
	if c.JWT.AccessTTL <= 0 || c.JWT.RefreshTTL <= 0 {
		return fmt.Errorf("JWT token TTL must be greater than zero")
	}
	if c.CORS.AllowCredentials {
		for _, origin := range c.CORS.AllowedOrigins {
			if origin == "*" {
				return fmt.Errorf("CORS_ALLOWED_ORIGINS cannot contain * when credentials are enabled")
			}
		}
	}
	if c.App.Environment == "production" {
		if isDevelopmentSecret(c.JWT.AccessSecret) || isDevelopmentSecret(c.JWT.RefreshSecret) {
			return fmt.Errorf("JWT secrets must be explicitly configured in production")
		}
	}
	return nil
}

func envString(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}

func envInt(key string, fallback int) (int, error) {
	value := envString(key, strconv.Itoa(fallback))
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("%s must be an integer: %w", key, err)
	}
	return parsed, nil
}

func envDuration(key string, fallback time.Duration) (time.Duration, error) {
	value := envString(key, fallback.String())
	parsed, err := time.ParseDuration(value)
	if err != nil {
		return 0, fmt.Errorf("%s must be a valid duration: %w", key, err)
	}
	return parsed, nil
}

func envBool(key string, fallback bool) bool {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func envCSV(key string, fallback []string) []string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}

	items := strings.Split(value, ",")
	result := make([]string, 0, len(items))
	for _, item := range items {
		if item = strings.TrimSpace(item); item != "" {
			result = append(result, item)
		}
	}
	if len(result) == 0 {
		return fallback
	}
	return result
}

func isDevelopmentSecret(value string) bool {
	return value == "local-access-secret-change-me" || value == "local-refresh-secret-change-me" || strings.HasPrefix(value, "replace-with-")
}
