package routes

import (
	"log"
	"net/http"
	"time"

	"maceli-backend/internal/config"
	"maceli-backend/internal/handlers"
	"maceli-backend/internal/middleware"
	"maceli-backend/internal/storage"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func SetupRouter(db *gorm.DB, cfg config.Config) *gin.Engine {
	router := gin.Default()
	router.MaxMultipartMemory = 8 << 20

	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{cfg.FrontendURL},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-ADMIN-KEY"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	router.Static("/uploads", "./uploads")

	imageUploader := buildImageUploader(cfg)

	planHandler := handlers.NewPlanHandler(db, imageUploader)
	pedidoHandler := handlers.NewPedidoHandler(db)
	contactoHandler := handlers.NewContactoHandler(db)
	uploadHandler := handlers.NewUploadHandler(imageUploader)

	api := router.Group("/api")
	api.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"message": "MACELI API funcionando",
		})
	})

	api.GET("/planes", planHandler.ListPublic)
	api.GET("/planes/:id", planHandler.GetPublic)
	api.POST("/pedidos", pedidoHandler.Create)
	api.POST("/contacto", contactoHandler.Create)

	admin := api.Group("/admin")
	admin.Use(middleware.AdminAuth(cfg.AdminKey))
	{
		admin.GET("/planes", planHandler.ListAdmin)
		admin.POST("/planes", planHandler.Create)
		admin.PUT("/planes/:id", planHandler.Update)
		admin.PATCH("/planes/:id/desactivar", planHandler.Deactivate)

		admin.GET("/pedidos", pedidoHandler.ListAdmin)
		admin.PATCH("/pedidos/:id/estado", pedidoHandler.UpdateEstado)

		admin.GET("/contacto", contactoHandler.ListAdmin)

		admin.POST("/upload", uploadHandler.UploadImage)
	}

	return router
}

func buildImageUploader(cfg config.Config) storage.ImageUploader {
	if cfg.CloudinaryCloudName != "" && cfg.CloudinaryAPIKey != "" && cfg.CloudinaryAPISecret != "" {
		uploader, err := storage.NewCloudinaryUploader(
			cfg.CloudinaryCloudName,
			cfg.CloudinaryAPIKey,
			cfg.CloudinaryAPISecret,
			cfg.CloudinaryUploadPath,
		)
		if err == nil {
			log.Println("Subida de imagenes configurada con Cloudinary")
			return uploader
		}

		log.Printf("No se pudo configurar Cloudinary, se usara almacenamiento local: %v", err)
	}

	log.Println("Subida de imagenes configurada en almacenamiento local")
	return storage.NewLocalUploader("uploads", "/uploads")
}
