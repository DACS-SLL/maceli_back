package models

import "time"

type Plan struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Nombre      string    `gorm:"not null" json:"nombre"`
	Descripcion string    `gorm:"type:text" json:"descripcion"`
	Precio      float64   `gorm:"type:numeric(10,2);not null;default:0" json:"precio"`
	Categoria   string    `json:"categoria"`
	ImagenURL   string    `json:"imagen_url"`
	Activo      bool      `gorm:"not null;default:true" json:"activo"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (Plan) TableName() string {
	return "planes"
}
