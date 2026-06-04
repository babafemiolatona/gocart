package config

import (
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
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

	AllowedOrigins []string
	TrustedProxies []string
}

var CFG *Config

func LoadConfig() {
	loadEnv()
	confLogger()
}

func loadEnv() {
	if os.Getenv("GO_MODE") != "release" {
		if err := godotenv.Load(); err != nil {
			log.Printf("Warning: .env file not found")
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
	}

	tokenDurationMinutes, err := strconv.Atoi(getEnvOptional("TOKEN_DURATION_MINUTES", "60"))
	if err != nil {
		log.Fatalf("Error parsing TOKEN_DURATION_MINUTES: %v", err)
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

func confLogger() {
	zerolog.TimeFieldFormat = time.RFC3339
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
	log.Fatalf("Environment variable %s is not set.", key)
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
		log.Fatalf("Invalid integer value for %s: %v", key, err)
	}
	return intVal
}

func parseDuration(value string) time.Duration {
	duration, err := time.ParseDuration(value)
	if err != nil {
		log.Fatalf("Invalid duration format for %s: %v", value, err)
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
