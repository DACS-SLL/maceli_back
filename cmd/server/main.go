package main

import (
	"fmt"
	"log"

	"maceli-backend/internal/config"
	"maceli-backend/internal/database"
	"maceli-backend/internal/routes"
)

func main() {
	cfg := config.Load()

	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("No se pudo conectar a la base de datos: %v", err)
	}

	if err := database.Migrate(db); err != nil {
		log.Fatalf("No se pudieron ejecutar las migraciones: %v", err)
	}

	if err := database.SeedPlans(db); err != nil {
		log.Fatalf("No se pudieron insertar los datos iniciales: %v", err)
	}

	router := routes.SetupRouter(db, cfg)

	log.Printf("MACELI API escuchando en http://localhost:%s/api", cfg.Port)
	if err := router.Run(fmt.Sprintf(":%s", cfg.Port)); err != nil {
		log.Fatalf("No se pudo iniciar el servidor: %v", err)
	}
}
