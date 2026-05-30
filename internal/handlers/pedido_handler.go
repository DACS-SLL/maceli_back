package handlers

import (
	"errors"
	"net/http"
	"strings"

	"maceli-backend/internal/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type PedidoHandler struct {
	db *gorm.DB
}

func NewPedidoHandler(db *gorm.DB) *PedidoHandler {
	return &PedidoHandler{db: db}
}

type pedidoRequest struct {
	NombreCliente string `json:"nombre_cliente"`
	Telefono      string `json:"telefono"`
	PlanID        uint   `json:"plan_id"`
	Mensaje       string `json:"mensaje"`
	DireccionZona string `json:"direccion_zona"`
}

type updatePedidoEstadoRequest struct {
	Estado string `json:"estado"`
}

func (h *PedidoHandler) Create(c *gin.Context) {
	var req pedidoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "El cuerpo de la solicitud no es valido"})
		return
	}

	if strings.TrimSpace(req.NombreCliente) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "El nombre del cliente es obligatorio"})
		return
	}

	if strings.TrimSpace(req.Telefono) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "El telefono es obligatorio"})
		return
	}

	if req.PlanID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "El plan es obligatorio"})
		return
	}

	var plan models.Plan
	if err := h.db.Where("id = ? AND activo = ?", req.PlanID, true).First(&plan).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "No se encontro el plan solicitado"})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": "No se pudo validar el plan"})
		return
	}

	pedido := models.Pedido{
		NombreCliente: strings.TrimSpace(req.NombreCliente),
		Telefono:      strings.TrimSpace(req.Telefono),
		PlanID:        plan.ID,
		Mensaje:       strings.TrimSpace(req.Mensaje),
		DireccionZona: strings.TrimSpace(req.DireccionZona),
		Estado:        models.EstadoPendiente,
	}

	if err := h.db.Create(&pedido).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "No se pudo registrar el pedido"})
		return
	}

	pedido.Plan = plan
	c.JSON(http.StatusCreated, gin.H{
		"message": "Pedido registrado correctamente",
		"data":    pedido,
	})
}

func (h *PedidoHandler) ListAdmin(c *gin.Context) {
	var pedidos []models.Pedido
	if err := h.db.Preload("Plan").Order("created_at DESC").Find(&pedidos).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "No se pudieron listar los pedidos"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": pedidos})
}

func (h *PedidoHandler) UpdateEstado(c *gin.Context) {
	var req updatePedidoEstadoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "El cuerpo de la solicitud no es valido"})
		return
	}

	req.Estado = strings.TrimSpace(strings.ToLower(req.Estado))
	if !models.IsValidPedidoEstado(req.Estado) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "El estado del pedido no es valido"})
		return
	}

	var pedido models.Pedido
	if err := h.db.Preload("Plan").First(&pedido, c.Param("id")).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "No se encontro el pedido solicitado"})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": "No se pudo obtener el pedido"})
		return
	}

	pedido.Estado = req.Estado
	if err := h.db.Save(&pedido).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "No se pudo actualizar el estado del pedido"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Estado del pedido actualizado correctamente",
		"data":    pedido,
	})
}
