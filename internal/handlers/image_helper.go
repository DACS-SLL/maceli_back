package handlers

import (
	"errors"
	"log"
	"net/http"
	"path/filepath"
	"strings"

	"maceli-backend/internal/storage"

	"github.com/gin-gonic/gin"
)

func saveUploadedImage(c *gin.Context, imageUploader storage.ImageUploader) (string, bool, int, string) {
	file, err := c.FormFile("imagen")
	if err != nil {
		if errors.Is(err, http.ErrMissingFile) {
			return "", false, http.StatusOK, ""
		}

		return "", false, http.StatusBadRequest, "No se pudo leer la imagen enviada"
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	if !isAllowedImageExtension(ext) {
		return "", false, http.StatusBadRequest, "La imagen debe ser JPG, PNG o WEBP"
	}

	openedFile, err := file.Open()
	if err != nil {
		return "", false, http.StatusBadRequest, "No se pudo abrir la imagen enviada"
	}
	defer openedFile.Close()

	imagenURL, err := imageUploader.Upload(c.Request.Context(), openedFile, file.Filename)
	if err != nil {
		log.Printf("error subiendo imagen: %v", err)
		return "", false, http.StatusInternalServerError, "No se pudo guardar la imagen"
	}

	return imagenURL, true, http.StatusCreated, ""
}

func isAllowedImageExtension(ext string) bool {
	switch ext {
	case ".jpg", ".jpeg", ".png", ".webp":
		return true
	default:
		return false
	}
}
