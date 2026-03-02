package config

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	ServerPort  string
	Environment string

	JWT        JWTConfig
	Cookie     CookieConfig
	Database   DatabaseConfig
	Cloudflare CloudflareConfig
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
}

type JWTConfig struct {
	Secret           string
	ExpiresIn        time.Duration
	RefreshSecret    string
	RefreshExpiresIn time.Duration
}

type CookieConfig struct {
	Domain   string
	Secure   bool
	HTTPOnly bool
}

type CloudflareConfig struct {
	AccountID       string
	AccessKeyID     string
	SecretAccessKey string
	Bucket          string
	PublicURL       string
}

func LoadConfig() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("file .env tidak ditemukan")
	}

	jwtExpires, _ := time.ParseDuration(getEnv("JWT_EXPIRES_IN", "24h"))
	jwtRefreshExpires, _ := time.ParseDuration(getEnv("JWT_REFRESH_EXPIRES_IN", "168h"))

	return &Config{
		ServerPort:  getEnv("SERVER_PORT", ":8080"),
		Environment: getEnv("ENVIRONMENT", "development"),

		JWT: JWTConfig{
			Secret: getEnv("JWT_SECRET", "tikitiki@adminpanel!fromaslabtoaslab!basereal"),
			ExpiresIn: jwtExpires,
			RefreshSecret: getEnv("JWT_REFRESH_SECRET", "tikitiki@adminpanel!fromaslabtoaslab!refresh"),
			RefreshExpiresIn: jwtRefreshExpires,
		},
		Cookie: CookieConfig{
			Domain: getEnv("COOKIE_DOMAIN", "localhost"),
			Secure: getEnv("COOKIE_SECURE", "false") == "true",
			HTTPOnly: getEnv("COOKIE_HTTP_ONLY", "true") == "true",
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "3306"),
			User:     getEnv("DB_USER", "root"),
			Password: getEnv("DB_PASSWORD", ""),
			Name:     getEnv("DB_NAME", "tikitiki"),
		},
		Cloudflare: CloudflareConfig{
			AccountID:       getEnv("CF_ACCOUNT_ID", ""),
			AccessKeyID:     getEnv("CF_ACCESS_KEY_ID", ""),
			SecretAccessKey: getEnv("CF_SECRET_ACCESS_KEY", ""),
			Bucket:          getEnv("CF_BUCKET", ""),
			PublicURL:       getEnv("CF_PUBLIC_URL", ""),
		},
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func (c *DatabaseConfig) GetDSN() string {
	return fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		c.User,
		c.Password,
		c.Host,
		c.Port,
		c.Name,
	)
}
