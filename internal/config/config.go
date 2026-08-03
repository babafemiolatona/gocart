package config

import (
	"os"
	"strconv"
	"strings"
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

	RedisURL string

	JWTSecret            string
	JWTExpiry            time.Duration
	TokenDurationMinutes int

	SMTPHost     string
	SMTPPort     int
	SMTPUser     string
	SMTPPassword string

	UploadDir     string
	MaxUploadSize int64

	// MinIO
	MinioEndpoint  string
	MinioAccessKey string
	MinioSecretKey string
	MinioBucket    string
	MinioUseSSL    bool

	AllowedOrigins []string
	TrustedProxies []string
}

var CFG *Config

func LoadConfig() {
	loadEnv()
}

func loadEnv() {
	if os.Getenv("GO_MODE") != "release" {
		if err := godotenv.Load(); err != nil {
			logger.Log.Warn().Err(err).Msg("warning: .env file not found")
		}
	}

	CFG = &Config{
		ServerPort:       ":" + getEnv("SERVER_PORT"),
		Env:              getEnv("ENV"),
		DatabaseHost:     getEnv("DB_HOST"),
		DatabasePort:     getEnv("DB_PORT"),
		DatabaseUser:     getEnv("DB_USER"),
		DatabasePassword: getEnv("DB_PASSWORD"),
		DatabaseName:     getEnv("DB_NAME"),
		DatabaseSSLMode:  getEnv("DB_SSL_MODE"),
		RedisURL:         getEnv("REDIS_URL"),
		JWTSecret:        getEnv("JWT_SECRET"),
		JWTExpiry:        parseDuration(getEnv("JWT_EXPIRY")),
		AllowedOrigins:   parseCommaSeparated(getEnv("ALLOWED_ORIGINS")),
		UploadDir:        getEnvOptional("UPLOAD_DIR", "./uploads"),
		MaxUploadSize:    int64(getEnvInt("MAX_UPLOAD_SIZE")),

		// MinIO
		MinioEndpoint:  getEnv("MINIO_ENDPOINT"),
		MinioAccessKey: getEnv("MINIO_ACCESS_KEY"),
		MinioSecretKey: getEnv("MINIO_SECRET_KEY"),
		MinioBucket:    getEnv("MINIO_BUCKET"),
		MinioUseSSL:    getEnvOptional("MINIO_USE_SSL", "false") == "true",
	}

	tokenDurationMinutes, err := strconv.Atoi(getEnvOptional("TOKEN_DURATION_MINUTES", "60"))
	if err != nil {
		logger.Log.Fatal().Err(err).Msg("error parsing TOKEN_DURATION_MINUTES")
	}
	CFG.TokenDurationMinutes = tokenDurationMinutes

	CFG.TrustedProxies = []string{}
	trustedProxies := getEnvOptional("TRUSTED_PROXY_IPS", "")
	if trustedProxies != "" {
		for _, proxy := range strings.Split(trustedProxies, ",") {
			if trimmed := strings.TrimSpace(proxy); trimmed != "" {
				CFG.TrustedProxies = append(CFG.TrustedProxies, trimmed)
			}
		}
	}
}

func (c *Config) GetDSN() string {
	return "postgres://" + c.DatabaseUser + ":" + c.DatabasePassword +
		"@" + c.DatabaseHost + ":" + c.DatabasePort +
		"/" + c.DatabaseName + "?sslmode=" + c.DatabaseSSLMode
}

func getEnv(key string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	logger.Log.Fatal().Str("variable", key).Msg("required environment variable is not set")
	return ""
}

func getEnvOptional(key string, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string) int {
	value := getEnv(key)
	intVal, err := strconv.Atoi(value)
	if err != nil {
		logger.Log.Fatal().Err(err).Str("variable", key).Msg("invalid integer value for environment variable")
	}
	return intVal
}

func parseDuration(value string) time.Duration {
	duration, err := time.ParseDuration(value)
	if err != nil {
		logger.Log.Fatal().Err(err).Str("value", value).Msg("invalid duration format")
	}
	return duration
}

func parseCommaSeparated(value string) []string {
	var result []string
	for _, item := range strings.Split(value, ",") {
		if trimmed := strings.TrimSpace(item); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}
