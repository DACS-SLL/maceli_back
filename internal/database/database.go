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
			Nombre:      "Plan Fit Maceli",
			Precio:      360,
			Categoria:   "Plan mensual",
			Descripcion: "Almuerzos saludables altos en proteínas, carbohidratos balanceados, grasas saludables y micronutrientes. Incluye 24 almuerzos, 24 bebidas y 24 postres.",
			Activo:      true,
		},
		{
			Nombre:      "Plan Familiar",
			Precio:      1125,
			Categoria:   "Plan familiar",
			Descripcion: "Alimentación equilibrada, nutritiva y variada para el hogar. Incluye 75 almuerzos, 75 bebidas y 75 postres.",
			Activo:      true,
		},
		{
			Nombre:      "Plan Doble Pack",
			Precio:      860,
			Categoria:   "Mensual; semanal S/. 220",
			Descripcion: "Almuerzo y cena saludables de la carta semanal, personalizados según preferencias y necesidades.",
			Activo:      true,
		},
		{
			Nombre:      "Plan Personalizado",
			Precio:      430,
			Categoria:   "Plan mensual",
			Descripcion: "Almuerzos personalizados según requerimientos específicos. Incluye 24 almuerzos, 24 bebidas y 24 postres personalizados.",
			Activo:      true,
		},
		{
			Nombre:      "Plan Triple Pack",
			Precio:      1160,
			Categoria:   "Mensual; semanal S/. 300",
			Descripcion: "Planificación completa con 24 desayunos personalizados, 24 almuerzos personalizados y 24 cenas personalizadas.",
			Activo:      true,
		},
		{
			Nombre:      "Menú Saludable del Día",
			Precio:      16,
			Categoria:   "Plato individual",
			Descripcion: "Opción flexible de comida saludable del día. Incluye bebida y postre fit; delivery según punto de entrega.",
			Activo:      true,
		},
	}

	return db.Create(&planes).Error
}
