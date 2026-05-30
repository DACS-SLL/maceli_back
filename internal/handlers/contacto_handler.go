package handlers

import (
	"net/http"
	"strings"

	"maceli-backend/internal/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ContactoHandler struct {
	db *gorm.DB
}

func NewContactoHandler(db *gorm.DB) *ContactoHandler {
	return &ContactoHandler{db: db}
}

type contactoRequest struct {
	Nombre   string `json:"nombre"`
	Telefono string `json:"telefono"`
	Correo   string `json:"correo"`
	Mensaje  string `json:"mensaje"`
}

func (h *ContactoHandler) Create(c *gin.Context) {
	var req contactoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "El cuerpo de la solicitud no es valido"})
		return
	}

	if strings.TrimSpace(req.Mensaje) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "El mensaje de contacto no debe estar vacio"})
		return
	}

	contacto := models.Contacto{
		Nombre:   strings.TrimSpace(req.Nombre),
		Telefono: strings.TrimSpace(req.Telefono),
		Correo:   strings.TrimSpace(req.Correo),
		Mensaje:  strings.TrimSpace(req.Mensaje),
	}

	if err := h.db.Create(&contacto).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "No se pudo registrar el mensaje de contacto"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Mensaje de contacto registrado correctamente",
		"data":    contacto,
	})
}

func (h *ContactoHandler) ListAdmin(c *gin.Context) {
	var contactos []models.Contacto
	if err := h.db.Order("created_at DESC").Find(&contactos).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "No se pudieron listar los mensajes de contacto"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": contactos})
}
