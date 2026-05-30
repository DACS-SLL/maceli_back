package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type UploadHandler struct {
	db *gorm.DB
}

func NewUploadHandler(db *gorm.DB) *UploadHandler {
	return &UploadHandler{db: db}
}

func (h *UploadHandler) UploadImage(c *gin.Context) {
	imagenURL, saved, status, message := saveUploadedImage(c)
	if !saved {
		if message == "" {
			message = "Debes enviar una imagen en el campo imagen"
			status = http.StatusBadRequest
		}
		c.JSON(status, gin.H{"error": message})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":    "Imagen subida correctamente",
		"imagen_url": imagenURL,
	})
}
