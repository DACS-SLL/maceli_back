package models

import "time"

const (
	EstadoPendiente  = "pendiente"
	EstadoContactado = "contactado"
	EstadoConfirmado = "confirmado"
	EstadoCancelado  = "cancelado"
)

type Pedido struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	NombreCliente string    `gorm:"not null" json:"nombre_cliente"`
	Telefono      string    `gorm:"not null" json:"telefono"`
	PlanID        uint      `gorm:"not null" json:"plan_id"`
	Plan          Plan      `json:"plan"`
	Mensaje       string    `gorm:"type:text" json:"mensaje"`
	DireccionZona string    `json:"direccion_zona"`
	Estado        string    `gorm:"not null;default:pendiente" json:"estado"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

func (Pedido) TableName() string {
	return "pedidos"
}

func IsValidPedidoEstado(estado string) bool {
	switch estado {
	case EstadoPendiente, EstadoContactado, EstadoConfirmado, EstadoCancelado:
		return true
	default:
		return false
	}
}
