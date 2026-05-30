package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port                 string
	DatabaseURL          string
	FrontendURL          string
	AdminKey             string
	CloudinaryCloudName  string
	CloudinaryAPIKey     string
	CloudinaryAPISecret  string
	CloudinaryUploadPath string
}

func Load() Config {
	if err := godotenv.Load(); err != nil {
		log.Println("No se encontro archivo .env, se usaran variables del entorno")
	}

	return Config{
		Port:                 getEnv("PORT", "8080"),
		DatabaseURL:          getEnv("DATABASE_URL", "postgres://usuario:password@localhost:5432/maceli_db?sslmode=disable"),
		FrontendURL:          getEnv("FRONTEND_URL", "http://localhost:5173"),
		AdminKey:             getEnv("ADMIN_KEY", "maceli_admin_123"),
		CloudinaryCloudName:  getEnv("CLOUDINARY_CLOUD_NAME", ""),
		CloudinaryAPIKey:     getEnv("CLOUDINARY_API_KEY", ""),
		CloudinaryAPISecret:  getEnv("CLOUDINARY_API_SECRET", ""),
		CloudinaryUploadPath: getEnv("CLOUDINARY_UPLOAD_PATH", "maceli/planes"),
	}
}

func getEnv(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	return value
}
