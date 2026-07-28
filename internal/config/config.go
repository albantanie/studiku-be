package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DBHost         string
	DBPort         string
	DBUser         string
	DBPassword     string
	DBName         string
	DBSSLMode      string
	ServerPort     string
	AllowedOrigins string
	UploadDir      string
	// WebDir menunjuk ke hasil build frontend (react-fe/dist). Bila foldernya ada,
	// server ikut menyajikan berkas statis dan menangani deep link di satu origin.
	WebDir string
}

func Load() *Config {
	_ = godotenv.Load()

	return &Config{
		DBHost:         getEnv("DB_HOST", "localhost"),
		DBPort:         getEnv("DB_PORT", "5432"),
		DBUser:         getEnv("DB_USER", "postgres"),
		DBPassword:     getEnv("DB_PASSWORD", "postgres"),
		DBName:         getEnv("DB_NAME", "studi_ku"),
		DBSSLMode:      getEnv("DB_SSLMODE", "disable"),
		ServerPort:     getEnv("SERVER_PORT", "8080"),
		AllowedOrigins: getEnv("ALLOWED_ORIGINS", "http://localhost:5173,http://127.0.0.1:5173"),
		UploadDir:      getEnv("UPLOAD_DIR", "uploads"),
		WebDir:         getEnv("WEB_DIR", "../react-fe/dist"),
	}
}

func (c *Config) DSN() string {
	return "host=" + c.DBHost +
		" port=" + c.DBPort +
		" user=" + c.DBUser +
		" password=" + c.DBPassword +
		" dbname=" + c.DBName +
		" sslmode=" + c.DBSSLMode
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
