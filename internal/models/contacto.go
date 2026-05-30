package models

import "time"

type Contacto struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Nombre    string    `json:"nombre"`
	Telefono  string    `json:"telefono"`
	Correo    string    `json:"correo"`
	Mensaje   string    `gorm:"type:text;not null" json:"mensaje"`
	CreatedAt time.Time `json:"created_at"`
}

func (Contacto) TableName() string {
	return "contactos"
}
