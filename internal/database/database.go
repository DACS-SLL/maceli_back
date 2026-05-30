package database

import (
	"maceli-backend/internal/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func Connect(databaseURL string) (*gorm.DB, error) {
	return gorm.Open(postgres.Open(databaseURL), &gorm.Config{})
}

func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&models.Plan{},
		&models.Pedido{},
		&models.Contacto{},
	)
}

func SeedPlans(db *gorm.DB) error {
	var total int64
	if err := db.Model(&models.Plan{}).Count(&total).Error; err != nil {
		return err
	}

	if total > 0 {
		return nil
	}

	planes := []models.Plan{
		{
			Nombre:      "Plan semanal",
			Precio:      91,
			Categoria:   "Plan semanal",
			Descripcion: "Plan saludable de 7 dias para personas que desean organizar mejor su alimentacion.",
			Activo:      true,
		},
		{
			Nombre:      "Plan diario",
			Precio:      15,
			Categoria:   "Almuerzo diario",
			Descripcion: "Almuerzo saludable del dia, fresco y practico.",
			Activo:      true,
		},
		{
			Nombre:      "Plan vegano",
			Precio:      18,
			Categoria:   "Vegano",
			Descripcion: "Opcion saludable sin ingredientes de origen animal.",
			Activo:      true,
		},
		{
			Nombre:      "Plan cero grasas",
			Precio:      18,
			Categoria:   "Cero grasas",
			Descripcion: "Alternativa ligera para quienes desean cuidar su alimentacion.",
			Activo:      true,
		},
		{
			Nombre:      "Plan MACELI 20 almuerzos",
			Precio:      300,
			Categoria:   "Plan mensual",
			Descripcion: "Plan de 20 almuerzos saludables para mantener una rutina constante.",
			Activo:      true,
		},
	}

	return db.Create(&planes).Error
}
