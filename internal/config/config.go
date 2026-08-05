package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"gocart/internal/logger"

	"github.com/joho/godotenv"
)

type Config struct {
	ServerPort string
	Env        string

	DatabaseHost     string
	DatabasePort     string
	DatabaseUser     string
	DatabasePassword string
	DatabaseName     string
	DatabaseSSLMode  string

	JWTSecret            string
	JWTExpiry            time.Duration
	TokenDurationMinutes int

	SeedAdminEmail    string
	SeedAdminPassword string

	UploadDir     string
	MaxUploadSize int64

	// MinIO
	MinioEndpoint  string
	MinioAccessKey string
	MinioSecretKey string
	MinioBucket    string
	MinioUseSSL    bool
}

var CFG *Config

func LoadConfig() (*Config, error) {
	loadEnv()

	cfg := &Config{
		UploadDir:        optionalEnv("UPLOAD_DIR", "./uploads"),
		SeedAdminEmail:    optionalEnv("SEED_ADMIN_EMAIL", ""),
		SeedAdminPassword: optionalEnv("SEED_ADMIN_PASSWORD", ""),
		MinioUseSSL:      optionalEnv("MINIO_USE_SSL", "false") == "true",
	}

	for _, r := range []struct {
		key string
		dst *string
	}{
		{"SERVER_PORT", &cfg.ServerPort},
		{"ENV", &cfg.Env},
		{"DB_HOST", &cfg.DatabaseHost},
		{"DB_PORT", &cfg.DatabasePort},
		{"DB_USER", &cfg.DatabaseUser},
		{"DB_PASSWORD", &cfg.DatabasePassword},
		{"DB_NAME", &cfg.DatabaseName},
		{"DB_SSL_MODE", &cfg.DatabaseSSLMode},
		{"JWT_SECRET", &cfg.JWTSecret},
		{"MINIO_ENDPOINT", &cfg.MinioEndpoint},
		{"MINIO_ACCESS_KEY", &cfg.MinioAccessKey},
		{"MINIO_SECRET_KEY", &cfg.MinioSecretKey},
		{"MINIO_BUCKET", &cfg.MinioBucket},
	} {
		value, err := requiredEnv(r.key)
		if err != nil {
			return nil, err
		}
		*r.dst = value
	}

	cfg.ServerPort = ":" + cfg.ServerPort

	jwtExpiry, err := durationEnv("JWT_EXPIRY")
	if err != nil {
		return nil, err
	}
	cfg.JWTExpiry = jwtExpiry

	maxUploadSize, err := intEnv("MAX_UPLOAD_SIZE")
	if err != nil {
		return nil, err
	}
	cfg.MaxUploadSize = int64(maxUploadSize)

	tokenMinutes, err := strconv.Atoi(optionalEnv("TOKEN_DURATION_MINUTES", "60"))
	if err != nil {
		return nil, fmt.Errorf("invalid TOKEN_DURATION_MINUTES: %w", err)
	}
	cfg.TokenDurationMinutes = tokenMinutes

	CFG = cfg
	return cfg, nil
}

func loadEnv() {
	if os.Getenv("GO_MODE") != "release" {
		if err := godotenv.Load(); err != nil {
			logger.Log.Warn().Err(err).Msg("warning: .env file not found")
		}
	}
}

func (c *Config) GetDSN() string {
	return "postgres://" + c.DatabaseUser + ":" + c.DatabasePassword +
		"@" + c.DatabaseHost + ":" + c.DatabasePort +
		"/" + c.DatabaseName + "?sslmode=" + c.DatabaseSSLMode
}

func requiredEnv(key string) (string, error) {
	value := os.Getenv(key)
	if value == "" {
		return "", fmt.Errorf("required environment variable %s is not set", key)
	}
	return value, nil
}

func optionalEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func intEnv(key string) (int, error) {
	value, err := requiredEnv(key)
	if err != nil {
		return 0, err
	}

	intVal, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("invalid integer value for %s: %w", key, err)
	}
	return intVal, nil
}

func durationEnv(key string) (time.Duration, error) {
	value, err := requiredEnv(key)
	if err != nil {
		return 0, err
	}

	duration, err := time.ParseDuration(value)
	if err != nil {
		return 0, fmt.Errorf("invalid duration value for %s: %w", key, err)
	}
	return duration, nil
}
