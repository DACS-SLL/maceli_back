package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func saveUploadedImage(c *gin.Context) (string, bool, int, string) {
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

	if err := os.MkdirAll("uploads", os.ModePerm); err != nil {
		return "", false, http.StatusInternalServerError, "No se pudo preparar la carpeta de imagenes"
	}

	filename := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)
	destination := filepath.Join("uploads", filename)

	if err := c.SaveUploadedFile(file, destination); err != nil {
		return "", false, http.StatusInternalServerError, "No se pudo guardar la imagen"
	}

	return "/uploads/" + filename, true, http.StatusCreated, ""
}

func isAllowedImageExtension(ext string) bool {
	switch ext {
	case ".jpg", ".jpeg", ".png", ".webp":
		return true
	default:
		return false
	}
}
