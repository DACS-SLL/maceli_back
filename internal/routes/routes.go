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
	if err := router.SetTrustedProxies(cfg.TrustedProxies); err != nil {
		panic("TRUSTED_PROXIES inválido")
	}
	router.Use(middleware.SecurityHeaders())
	router.MaxMultipartMemory = 8 << 20

	router.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.FrontendURLs,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length", "Retry-After"},
		AllowCredentials: false,
		MaxAge:           12 * time.Hour,
	}))
	router.Use(middleware.RateLimit(180, 1800, time.Minute), middleware.BodyLimit(), middleware.NumericID())

	router.Static("/uploads", "./uploads")

	imageUploader := buildImageUploader(cfg)

	planHandler := handlers.NewPlanHandler(db, imageUploader)
	pedidoHandler := handlers.NewPedidoHandler(db)
	contactoHandler := handlers.NewContactoHandler(db)
	siteHandler := handlers.NewSiteHandler(db, imageUploader)
	authHandler := handlers.NewAuthHandler(db)

	api := router.Group("/api")
	api.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"message": "MACELI API funcionando",
		})
	})

	api.GET("/planes", planHandler.ListPublic)
	api.GET("/site", siteHandler.Public)
	api.POST("/auth/login", middleware.RateLimit(5, 30, time.Minute), authHandler.Login)
	api.GET("/planes/:id", planHandler.GetPublic)
	api.POST("/pedidos", middleware.RateLimit(5, 60, time.Minute), pedidoHandler.Create)
	api.POST("/contacto", middleware.RateLimit(5, 60, time.Minute), contactoHandler.Create)

	admin := api.Group("/admin")
	admin.Use(authHandler.RequireSession())
	{
		admin.POST("/logout", authHandler.Logout)
		admin.GET("/users", authHandler.Users)
		admin.POST("/users", middleware.RateLimit(5, 10, time.Minute), authHandler.CreateUser)
		admin.DELETE("/users/:id", authHandler.DeactivateUser)
		admin.PUT("/password", middleware.RateLimit(5, 10, time.Minute), authHandler.ChangePassword)
		admin.GET("/site", siteHandler.Draft)
		admin.PUT("/site", siteHandler.Save)
		admin.POST("/site/publish", siteHandler.Publish)
		admin.GET("/planes", planHandler.ListAdmin)
		admin.POST("/planes", planHandler.Create)
		admin.PUT("/planes/:id", planHandler.Update)
		admin.PATCH("/planes/:id/desactivar", planHandler.Deactivate)

		admin.GET("/pedidos", pedidoHandler.ListAdmin)
		admin.PATCH("/pedidos/:id/estado", pedidoHandler.UpdateEstado)

		admin.GET("/contacto", contactoHandler.ListAdmin)

		admin.POST("/upload", middleware.RateLimit(12, 30, time.Minute), siteHandler.Upload)
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

		if cfg.Production {
			log.Fatal("No se pudo configurar Cloudinary en producción")
		}
		log.Printf("No se pudo configurar Cloudinary, se usara almacenamiento local: %v", err)
	}

	if cfg.Production {
		log.Fatal("Cloudinary es obligatorio en producción")
	}
	log.Println("Subida de imagenes configurada en almacenamiento local")
	return storage.NewLocalUploader("uploads", "/uploads")
}
