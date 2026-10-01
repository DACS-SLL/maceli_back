package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"maceli-backend/internal/models"
	"maceli-backend/internal/storage"
)

type SiteHandler struct {
	db       *gorm.DB
	uploader storage.ImageUploader
}

func NewSiteHandler(db *gorm.DB, uploader storage.ImageUploader) *SiteHandler {
	return &SiteHandler{db, uploader}
}

func InitSite(db *gorm.DB) error {
	raw, _ := json.Marshal(models.SiteContent{})
	return db.Clauses(clause.OnConflict{DoNothing: true}).Create(&models.SiteState{ID: 1, Version: 1, Draft: string(raw), Published: string(raw)}).Error
}

func menuStatus(menu models.WeeklyMenu, now time.Time) string {
	if len(menu.Pages) == 0 {
		return "empty"
	}
	today := now.In(time.FixedZone("America/Lima", -5*3600)).Format("2006-01-02")
	if today < menu.Start {
		return "upcoming"
	}
	if today > menu.End {
		return "expired"
	}
	return "current"
}

func validateContent(content models.SiteContent) error {
	groups := [][]models.SiteImage{content.Hero, content.Dishes, content.Plans, content.About, content.Menu.Pages}
	limits := []int{2, 12, 8, 8, 8}
	for i, group := range groups {
		if len(group) > limits[i] {
			return errors.New("Se excedió la cantidad de imágenes permitida")
		}
		for _, img := range group {
			if img.URL == "" || len(img.URL) > 2048 || strings.TrimSpace(img.Alt) == "" || len(img.Alt) > 500 {
				return errors.New("Cada imagen necesita una descripción de hasta 500 caracteres y un archivo válido")
			}
		}
	}
	if len(content.Menu.Title) > 160 || len(content.Menu.Description) > 4000 {
		return errors.New("El título o la descripción de la carta es demasiado largo")
	}
	if len(content.Menu.Pages) > 0 {
		start, err := time.Parse("2006-01-02", content.Menu.Start)
		end, endErr := time.Parse("2006-01-02", content.Menu.End)
		if err != nil || endErr != nil || end.Sub(start) != 6*24*time.Hour {
			return errors.New("La carta debe cubrir siete días: el inicio y los seis días siguientes")
		}
		if strings.TrimSpace(content.Menu.Title) == "" {
			return errors.New("Escribe un título para la carta semanal")
		}
	}
	return nil
}

func (h *SiteHandler) validateAssets(c *gin.Context, content models.SiteContent) error {
	for _, group := range [][]models.SiteImage{content.Hero, content.Dishes, content.Plans, content.About, content.Menu.Pages} {
		for _, img := range group {
			var count int64
			if h.db.WithContext(c.Request.Context()).Model(&models.MediaAsset{}).Where("url = ?", img.URL).Count(&count).Error != nil || count != 1 {
				return errors.New("Utiliza imágenes subidas desde este panel")
			}
		}
	}
	return nil
}

func (h *SiteHandler) Public(c *gin.Context) {
	var state models.SiteState
	if h.db.WithContext(c.Request.Context()).First(&state, 1).Error != nil {
		c.JSON(503, gin.H{"error": "No se pudo cargar el contenido"})
		return
	}
	var content models.SiteContent
	if json.Unmarshal([]byte(state.Published), &content) != nil {
		c.JSON(500, gin.H{"error": "No se pudo leer el contenido"})
		return
	}
	status := menuStatus(content.Menu, time.Now())
	if status != "current" {
		content.Menu.Pages = nil
		content.Menu.Title = "Carta semanal"
		content.Menu.Description = ""
	}
	c.Header("Cache-Control", "public, max-age=30, s-maxage=30")
	c.JSON(200, gin.H{"content": content, "menu_status": status, "published_at": state.PublishedAt})
}

func (h *SiteHandler) Draft(c *gin.Context) {
	var state models.SiteState
	if h.db.WithContext(c.Request.Context()).First(&state, 1).Error != nil {
		c.JSON(500, gin.H{"error": "No se pudo cargar el borrador"})
		return
	}
	c.JSON(200, gin.H{"content": json.RawMessage(state.Draft), "version": state.Version, "published_at": state.PublishedAt})
}

func (h *SiteHandler) Save(c *gin.Context) {
	var req struct {
		Content models.SiteContent `json:"content"`
		Version uint               `json:"version"`
	}
	if c.ShouldBindJSON(&req) != nil {
		c.JSON(400, gin.H{"error": "Borrador inválido"})
		return
	}
	if err := validateContent(req.Content); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	if err := h.validateAssets(c, req.Content); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	raw, _ := json.Marshal(req.Content)
	result := h.db.WithContext(c.Request.Context()).Model(&models.SiteState{}).Where("id = 1 AND version = ?", req.Version).Updates(map[string]interface{}{"draft": string(raw), "version": gorm.Expr("version + 1"), "updated_by": c.GetUint("adminID")})
	if result.Error != nil {
		c.JSON(500, gin.H{"error": "No se pudo guardar; el sitio publicado no cambió"})
		return
	}
	if result.RowsAffected == 0 {
		c.JSON(409, gin.H{"error": "Otro administrador guardó cambios. Recarga el borrador antes de continuar."})
		return
	}
	c.JSON(200, gin.H{"version": req.Version + 1})
}

func (h *SiteHandler) Publish(c *gin.Context) {
	var req struct {
		Version uint `json:"version"`
	}
	if c.ShouldBindJSON(&req) != nil {
		c.JSON(400, gin.H{"error": "Solicitud inválida"})
		return
	}
	var state models.SiteState
	if h.db.WithContext(c.Request.Context()).First(&state, 1).Error != nil {
		c.JSON(500, gin.H{"error": "No se pudo cargar el borrador"})
		return
	}
	if state.Version != req.Version {
		c.JSON(409, gin.H{"error": "El borrador cambió. Recárgalo antes de publicar."})
		return
	}
	var content models.SiteContent
	if json.Unmarshal([]byte(state.Draft), &content) != nil {
		c.JSON(500, gin.H{"error": "Borrador inválido"})
		return
	}
	if err := validateContent(content); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	// A future week stays a draft until its start date; never hides the current menu early.
	if len(content.Menu.Pages) > 0 && menuStatus(content.Menu, time.Now()) != "current" {
		c.JSON(400, gin.H{"error": "Para publicar, la carta debe estar vigente hoy. Ajusta las fechas o retírala del borrador."})
		return
	}
	result := h.db.WithContext(c.Request.Context()).Model(&models.SiteState{}).Where("id = 1 AND version = ?", req.Version).Updates(map[string]interface{}{"published": state.Draft, "published_at": time.Now(), "version": gorm.Expr("version + 1"), "updated_by": c.GetUint("adminID")})
	if result.Error != nil {
		c.JSON(500, gin.H{"error": "No se pudo publicar; la versión anterior sigue disponible"})
		return
	}
	if result.RowsAffected == 0 {
		c.JSON(409, gin.H{"error": "El borrador cambió. Recárgalo y revisa antes de publicar."})
		return
	}
	c.JSON(200, gin.H{"version": req.Version + 1, "message": "Sitio publicado"})
}

func (h *SiteHandler) Upload(c *gin.Context) {
	url, saved, status, message := saveUploadedImage(c, h.uploader)
	if !saved {
		if message == "" {
			message = "Selecciona una imagen"
			status = 400
		}
		c.JSON(status, gin.H{"error": message})
		return
	}
	asset := models.MediaAsset{URL: url, UploadedBy: c.GetUint("adminID")}
	if err := h.db.WithContext(c.Request.Context()).Create(&asset).Error; err != nil {
		c.JSON(500, gin.H{"error": "No se pudo registrar la imagen. El sitio no cambió."})
		return
	}
	c.JSON(201, gin.H{"url": url, "id": fmt.Sprint(asset.ID)})
}
