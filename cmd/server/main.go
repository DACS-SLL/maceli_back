package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"maceli-backend/internal/config"
	"maceli-backend/internal/database"
	"maceli-backend/internal/handlers"
	"maceli-backend/internal/routes"
)

func main() {
	cfg := config.Load()
	if cfg.Production && (cfg.CloudinaryCloudName == "" || cfg.CloudinaryAPIKey == "" || cfg.CloudinaryAPISecret == "") {
		log.Fatal("Configura Cloudinary en producción; el disco local no es persistente")
	}

	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("No se pudo conectar a la base de datos: %v", err)
	}

	if err := database.Migrate(db); err != nil {
		log.Fatalf("No se pudieron ejecutar las migraciones: %v", err)
	}

	if err := handlers.InitSite(db); err != nil {
		log.Fatalf("No se pudo inicializar el sitio: %v", err)
	}
	if err := handlers.BootstrapAdmin(db, cfg.AdminEmail, cfg.AdminPassword); err != nil {
		log.Fatalf("No se pudo configurar administración: %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatal(err)
	}
	sqlDB.SetMaxOpenConns(10)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(30 * time.Minute)

	router := routes.SetupRouter(db, cfg)

	log.Printf("MACELI API escuchando en http://localhost:%s/api", cfg.Port)
	server := &http.Server{Addr: fmt.Sprintf(":%s", cfg.Port), Handler: router, ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 40 * time.Second, WriteTimeout: 45 * time.Second, IdleTimeout: 60 * time.Second, MaxHeaderBytes: 16 << 10}
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("No se pudo iniciar el servidor: %v", err)
	}
}
