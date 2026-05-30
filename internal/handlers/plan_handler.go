package handlers

import (
	"errors"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"

	"maceli-backend/internal/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type PlanHandler struct {
	db *gorm.DB
}

func NewPlanHandler(db *gorm.DB) *PlanHandler {
	return &PlanHandler{db: db}
}

type planRequest struct {
	Nombre      string  `json:"nombre"`
	Descripcion string  `json:"descripcion"`
	Precio      float64 `json:"precio"`
	Categoria   string  `json:"categoria"`
	ImagenURL   string  `json:"imagen_url"`
	Activo      *bool   `json:"activo"`
}

type updatePlanRequest struct {
	Nombre      *string  `json:"nombre"`
	Descripcion *string  `json:"descripcion"`
	Precio      *float64 `json:"precio"`
	Categoria   *string  `json:"categoria"`
	ImagenURL   *string  `json:"imagen_url"`
	Activo      *bool    `json:"activo"`
}

func (h *PlanHandler) ListPublic(c *gin.Context) {
	var planes []models.Plan
	if err := h.db.Where("activo = ?", true).Order("id ASC").Find(&planes).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "No se pudieron listar los planes"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": planes})
}

func (h *PlanHandler) GetPublic(c *gin.Context) {
	var plan models.Plan
	if err := h.db.Where("id = ? AND activo = ?", c.Param("id"), true).First(&plan).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "No se encontro el plan solicitado"})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": "No se pudo obtener el plan"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": plan})
}

func (h *PlanHandler) ListAdmin(c *gin.Context) {
	var planes []models.Plan
	if err := h.db.Order("id ASC").Find(&planes).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "No se pudieron listar los planes"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": planes})
}

func (h *PlanHandler) Create(c *gin.Context) {
	if c.ContentType() == "multipart/form-data" {
		h.createWithMultipart(c)
		return
	}

	var req planRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "El cuerpo de la solicitud no es valido"})
		return
	}

	if strings.TrimSpace(req.Nombre) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "El nombre del plan es obligatorio"})
		return
	}

	if req.Precio < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "El precio del plan debe ser mayor o igual a 0"})
		return
	}

	activo := true
	if req.Activo != nil {
		activo = *req.Activo
	}

	plan := models.Plan{
		Nombre:      strings.TrimSpace(req.Nombre),
		Descripcion: strings.TrimSpace(req.Descripcion),
		Precio:      req.Precio,
		Categoria:   strings.TrimSpace(req.Categoria),
		ImagenURL:   strings.TrimSpace(req.ImagenURL),
		Activo:      activo,
	}

	if err := h.db.Create(&plan).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "No se pudo crear el plan"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Plan creado correctamente",
		"data":    plan,
	})
}

func (h *PlanHandler) Update(c *gin.Context) {
	var plan models.Plan
	if err := h.db.First(&plan, c.Param("id")).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "No se encontro el plan solicitado"})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": "No se pudo obtener el plan"})
		return
	}

	if c.ContentType() == "multipart/form-data" {
		if !h.applyMultipartPlanUpdate(c, &plan) {
			return
		}
	} else {
		var req updatePlanRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "El cuerpo de la solicitud no es valido"})
			return
		}

		if req.Nombre != nil {
			if strings.TrimSpace(*req.Nombre) == "" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "El nombre del plan es obligatorio"})
				return
			}
			plan.Nombre = strings.TrimSpace(*req.Nombre)
		}

		if req.Precio != nil {
			if *req.Precio < 0 {
				c.JSON(http.StatusBadRequest, gin.H{"error": "El precio del plan debe ser mayor o igual a 0"})
				return
			}
			plan.Precio = *req.Precio
		}

		if req.Descripcion != nil {
			plan.Descripcion = strings.TrimSpace(*req.Descripcion)
		}
		if req.Categoria != nil {
			plan.Categoria = strings.TrimSpace(*req.Categoria)
		}
		if req.ImagenURL != nil {
			plan.ImagenURL = strings.TrimSpace(*req.ImagenURL)
		}
		if req.Activo != nil {
			plan.Activo = *req.Activo
		}
	}

	if err := h.db.Save(&plan).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "No se pudo actualizar el plan"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Plan actualizado correctamente",
		"data":    plan,
	})
}

func (h *PlanHandler) createWithMultipart(c *gin.Context) {
	form, err := c.MultipartForm()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "El formulario enviado no es valido"})
		return
	}

	nombre := strings.TrimSpace(firstMultipartValue(form, "nombre"))
	if nombre == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "El nombre del plan es obligatorio"})
		return
	}

	precio, ok := parseOptionalPrice(firstMultipartValue(form, "precio"))
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "El precio del plan debe ser mayor o igual a 0"})
		return
	}

	activo := true
	if multipartFieldExists(form, "activo") {
		parsedActivo, ok := parseBoolValue(firstMultipartValue(form, "activo"))
		if !ok {
			c.JSON(http.StatusBadRequest, gin.H{"error": "El valor de activo debe ser true o false"})
			return
		}
		activo = parsedActivo
	}

	imagenURL := strings.TrimSpace(firstMultipartValue(form, "imagen_url"))
	if uploadedURL, saved, status, message := saveUploadedImage(c); saved {
		imagenURL = uploadedURL
	} else if message != "" {
		c.JSON(status, gin.H{"error": message})
		return
	}

	plan := models.Plan{
		Nombre:      nombre,
		Descripcion: strings.TrimSpace(firstMultipartValue(form, "descripcion")),
		Precio:      precio,
		Categoria:   strings.TrimSpace(firstMultipartValue(form, "categoria")),
		ImagenURL:   imagenURL,
		Activo:      activo,
	}

	if err := h.db.Create(&plan).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "No se pudo crear el plan"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Plan creado correctamente",
		"data":    plan,
	})
}

func (h *PlanHandler) applyMultipartPlanUpdate(c *gin.Context, plan *models.Plan) bool {
	form, err := c.MultipartForm()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "El formulario enviado no es valido"})
		return false
	}

	if multipartFieldExists(form, "nombre") {
		nombre := strings.TrimSpace(firstMultipartValue(form, "nombre"))
		if nombre == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "El nombre del plan es obligatorio"})
			return false
		}
		plan.Nombre = nombre
	}

	if multipartFieldExists(form, "precio") {
		precio, ok := parseOptionalPrice(firstMultipartValue(form, "precio"))
		if !ok {
			c.JSON(http.StatusBadRequest, gin.H{"error": "El precio del plan debe ser mayor o igual a 0"})
			return false
		}
		plan.Precio = precio
	}

	if multipartFieldExists(form, "descripcion") {
		plan.Descripcion = strings.TrimSpace(firstMultipartValue(form, "descripcion"))
	}
	if multipartFieldExists(form, "categoria") {
		plan.Categoria = strings.TrimSpace(firstMultipartValue(form, "categoria"))
	}
	if multipartFieldExists(form, "imagen_url") {
		plan.ImagenURL = strings.TrimSpace(firstMultipartValue(form, "imagen_url"))
	}
	if multipartFieldExists(form, "activo") {
		activo, ok := parseBoolValue(firstMultipartValue(form, "activo"))
		if !ok {
			c.JSON(http.StatusBadRequest, gin.H{"error": "El valor de activo debe ser true o false"})
			return false
		}
		plan.Activo = activo
	}

	if uploadedURL, saved, status, message := saveUploadedImage(c); saved {
		plan.ImagenURL = uploadedURL
	} else if message != "" {
		c.JSON(status, gin.H{"error": message})
		return false
	}

	return true
}

func (h *PlanHandler) Deactivate(c *gin.Context) {
	var plan models.Plan
	if err := h.db.First(&plan, c.Param("id")).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "No se encontro el plan solicitado"})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": "No se pudo obtener el plan"})
		return
	}

	plan.Activo = false
	if err := h.db.Save(&plan).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "No se pudo desactivar el plan"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Plan desactivado correctamente",
		"data":    plan,
	})
}

func firstMultipartValue(form *multipart.Form, key string) string {
	values, ok := form.Value[key]
	if !ok || len(values) == 0 {
		return ""
	}

	return values[0]
}

func multipartFieldExists(form *multipart.Form, key string) bool {
	_, ok := form.Value[key]
	return ok
}

func parseOptionalPrice(value string) (float64, bool) {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0, true
	}

	price, err := strconv.ParseFloat(value, 64)
	if err != nil || price < 0 {
		return 0, false
	}

	return price, true
}

func parseBoolValue(value string) (bool, bool) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "true", "1", "si":
		return true, true
	case "false", "0", "no":
		return false, true
	default:
		return false, false
	}
}
