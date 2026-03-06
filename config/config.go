package config

import (
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	// Server
	Port string

	// Database
	DBHost            string
	DBUser            string
	DBPassword        string
	DBName            string
	DBPort            string
	DBSSLMode         string
	DBMaxOpenConns    int
	DBMaxIdleConns    int
	DBConnMaxLifetime time.Duration
	DBConnMaxIdleTime time.Duration

	// Firebase Admin (inline credentials)
	FirebaseProjectID   string
	FirebasePrivateKey  string
	FirebaseClientEmail string

	// FP Tenant API
	FPBaseURL      string
	FPClientID     string
	FPClientSecret string
	FPTenantID     string

	// FP POA API
	FPPoaBaseURL      string
	FPPoaAuthURL      string
	FPPoaClientID     string
	FPPoaClientSecret string

	// MSG91
	MSG91AuthKey    string
	MSG91TemplateID string

	// Payment
	PaymentPostbackURL  string
	MandatePostbackURL  string
}

var AppConfig *Config

func Init() {
	_ = godotenv.Load()

	AppConfig = &Config{
		Port: getEnv("PORT", "9113"),

		DBHost:            getEnv("DB_HOST", ""),
		DBUser:            getEnv("DB_USER", ""),
		DBPassword:        getEnv("DB_PASSWORD", ""),
		DBName:            getEnv("DB_NAME", "sure-app"),
		DBPort:            getEnv("DB_PORT", "5432"),
		DBSSLMode:         getEnv("DB_SSL_MODE", "require"),
		DBMaxOpenConns:    getIntEnv("DB_MAX_OPEN_CONNS", 100),
		DBMaxIdleConns:    getIntEnv("DB_MAX_IDLE_CONNS", 10),
		DBConnMaxLifetime: getDurationEnv("DB_CONN_MAX_LIFETIME", 1*time.Hour),
		DBConnMaxIdleTime: getDurationEnv("DB_CONN_MAX_IDLE_TIME", 10*time.Minute),

		FirebaseProjectID:   getEnv("FIREBASE_PROJECT_ID", ""),
		FirebasePrivateKey:  strings.ReplaceAll(getEnv("FIREBASE_PRIVATE_KEY", ""), `\n`, "\n"),
		FirebaseClientEmail: getEnv("FIREBASE_CLIENT_EMAIL", ""),

		FPBaseURL:      getEnv("FP_BASE_URL", ""),
		FPClientID:     getEnv("FP_CLIENT_ID", ""),
		FPClientSecret: getEnv("FP_CLIENT_SECRET", ""),
		FPTenantID:     getEnv("FP_TENANT_ID", ""),

		FPPoaBaseURL:      getEnv("FP_POA_BASE_URL", ""),
		FPPoaAuthURL:      getEnv("FP_POA_AUTH_URL", ""),
		FPPoaClientID:     getEnv("FP_POA_CLIENT_ID", ""),
		FPPoaClientSecret: getEnv("FP_POA_CLIENT_SECRET", ""),

		MSG91AuthKey:    getEnv("MSG91_AUTH_KEY", ""),
		MSG91TemplateID: getEnv("MSG91_TEMPLATE_ID", ""),

		PaymentPostbackURL:  getEnv("PAYMENT_POSTBACK_URL", "http://localhost:9113/sure-mf/callbacks/payment"),
		MandatePostbackURL:  getEnv("MANDATE_POSTBACK_URL", "http://localhost:9113/sure-mf/callbacks/mandate"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return strings.TrimSpace(value)
	}
	return defaultValue
}

func getIntEnv(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getDurationEnv(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}
