package handlers

import (
	"net/http"

	"maceli-backend/internal/storage"

	"github.com/gin-gonic/gin"
)

type UploadHandler struct {
	imageUploader storage.ImageUploader
}

func NewUploadHandler(imageUploader storage.ImageUploader) *UploadHandler {
	return &UploadHandler{imageUploader: imageUploader}
}

func (h *UploadHandler) UploadImage(c *gin.Context) {
	imagenURL, saved, status, message := saveUploadedImage(c, h.imageUploader)
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
