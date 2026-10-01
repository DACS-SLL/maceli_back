package handlers

import (
	"context"
	"errors"
	_ "golang.org/x/image/webp"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"log"
	"net/http"
	"path/filepath"
	"strings"
	"time"

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
	if file.Size > 8<<20 || file.Size == 0 {
		return "", false, http.StatusRequestEntityTooLarge, "Cada imagen debe pesar como máximo 8 MB"
	}

	openedFile, err := file.Open()
	if err != nil {
		return "", false, http.StatusBadRequest, "No se pudo abrir la imagen enviada"
	}
	defer openedFile.Close()
	config, format, err := image.DecodeConfig(openedFile)
	if err != nil || config.Width < 1 || config.Height < 1 || int64(config.Width)*int64(config.Height) > 40000000 || config.Width > 16000 || config.Height > 16000 {
		return "", false, http.StatusBadRequest, "Imagen inválida o demasiado grande (máximo 40 megapíxeles)"
	}
	validFormat := (format == "jpeg" && (ext == ".jpg" || ext == ".jpeg")) || (format == "png" && ext == ".png") || (format == "webp" && ext == ".webp")
	if !validFormat {
		return "", false, http.StatusBadRequest, "El contenido del archivo no coincide con su extensión"
	}
	if _, err := openedFile.Seek(0, 0); err != nil {
		return "", false, http.StatusBadRequest, "No se pudo leer la imagen"
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	imagenURL, err := imageUploader.Upload(ctx, openedFile, file.Filename)
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
