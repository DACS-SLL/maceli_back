package handlers

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net/http"
	"net/mail"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"maceli-backend/internal/models"
)

type AuthHandler struct {
	db    *gorm.DB
	dummy []byte
}

func NewAuthHandler(db *gorm.DB) *AuthHandler {
	dummy, _ := bcrypt.GenerateFromPassword([]byte("unused-comparison-password"), 12)
	return &AuthHandler{db: db, dummy: dummy}
}

func validCredentials(email, password string) bool {
	address, err := mail.ParseAddress(email)
	return err == nil && address.Address == email && len(email) <= 254 && len(password) >= 12 && len(password) <= 72
}

// Bootstrap only creates the first administrator, never resets an existing password.
func BootstrapAdmin(db *gorm.DB, email, password string) error {
	if email == "" && password == "" {
		return nil
	}
	email = strings.ToLower(strings.TrimSpace(email))
	if !validCredentials(email, password) {
		return errors.New("ADMIN_EMAIL y ADMIN_PASSWORD requieren correo válido y contraseña de 12 a 72 bytes")
	}
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec("SELECT pg_advisory_xact_lock(734621)").Error; err != nil {
			return err
		}
		var count int64
		if err := tx.Model(&models.AdminUser{}).Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			return nil
		}
		hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
		if err != nil {
			return err
		}
		return tx.Create(&models.AdminUser{Email: email, PasswordHash: string(hash), Active: true}).Error
	})
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if c.ShouldBindJSON(&req) != nil || len(req.Password) > 72 || len(req.Email) > 254 {
		c.JSON(400, gin.H{"error": "Credenciales inválidas"})
		return
	}
	var user models.AdminUser
	err := h.db.WithContext(c.Request.Context()).Where("email = ? AND active = ?", strings.ToLower(strings.TrimSpace(req.Email)), true).First(&user).Error
	hash := h.dummy
	if err == nil {
		hash = []byte(user.PasswordHash)
	}
	match := bcrypt.CompareHashAndPassword(hash, []byte(req.Password)) == nil
	if err != nil || !match {
		c.JSON(401, gin.H{"error": "Correo o contraseña incorrectos"})
		return
	}
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		c.JSON(500, gin.H{"error": "No se pudo iniciar sesión"})
		return
	}
	token := hex.EncodeToString(bytes)
	expires := time.Now().Add(8 * time.Hour)
	if err := h.db.WithContext(c.Request.Context()).Transaction(func(tx *gorm.DB) error {
		var current models.AdminUser
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&current, user.ID).Error; err != nil {
			return err
		}
		if !current.Active || current.PasswordHash != user.PasswordHash {
			return errors.New("credentials changed")
		}
		if err := tx.Where("expires_at < ? OR user_id = ?", time.Now(), user.ID).Delete(&models.AdminSession{}).Error; err != nil {
			return err
		}
		return tx.Create(&models.AdminSession{TokenHash: hashToken(token), UserID: user.ID, ExpiresAt: expires}).Error
	}); err != nil {
		c.JSON(500, gin.H{"error": "No se pudo iniciar sesión"})
		return
	}
	c.JSON(200, gin.H{"token": token, "expires_at": expires, "user": user})
}

func (h *AuthHandler) RequireSession() gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			c.AbortWithStatusJSON(401, gin.H{"error": "Inicia sesión para continuar"})
			return
		}
		token := strings.TrimPrefix(header, "Bearer ")
		var session models.AdminSession
		var user models.AdminUser
		if len(token) != 64 || h.db.WithContext(c.Request.Context()).Where("token_hash = ? AND expires_at > ?", hashToken(token), time.Now()).First(&session).Error != nil || h.db.WithContext(c.Request.Context()).Where("id = ? AND active = ?", session.UserID, true).First(&user).Error != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "La sesión venció. Vuelve a ingresar."})
			return
		}
		c.Set("adminID", user.ID)
		c.Set("sessionHash", session.TokenHash)
		c.Next()
	}
}

func (h *AuthHandler) Logout(c *gin.Context) {
	if h.db.WithContext(c.Request.Context()).Where("token_hash = ?", c.GetString("sessionHash")).Delete(&models.AdminSession{}).Error != nil {
		c.JSON(500, gin.H{"error": "No se pudo cerrar la sesión"})
		return
	}
	c.Status(204)
}

func (h *AuthHandler) Users(c *gin.Context) {
	var users []models.AdminUser
	if h.db.WithContext(c.Request.Context()).Order("id").Find(&users).Error != nil {
		c.JSON(500, gin.H{"error": "No se pudieron cargar las cuentas"})
		return
	}
	c.JSON(200, gin.H{"data": users})
}

func (h *AuthHandler) CreateUser(c *gin.Context) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if c.ShouldBindJSON(&req) != nil {
		c.JSON(400, gin.H{"error": "Solicitud inválida"})
		return
	}
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))
	if !validCredentials(req.Email, req.Password) {
		c.JSON(400, gin.H{"error": "Usa un correo válido y una contraseña de 12 a 72 bytes"})
		return
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), 12)
	if err != nil {
		c.JSON(500, gin.H{"error": "No se pudo crear la cuenta"})
		return
	}
	user := models.AdminUser{Email: req.Email, PasswordHash: string(hash), Active: true}
	if h.db.WithContext(c.Request.Context()).Create(&user).Error != nil {
		c.JSON(409, gin.H{"error": "No se pudo crear la cuenta. Comprueba si el correo ya existe."})
		return
	}
	c.JSON(201, gin.H{"data": user})
}

func (h *AuthHandler) DeactivateUser(c *gin.Context) {
	var user models.AdminUser
	if h.db.WithContext(c.Request.Context()).First(&user, "id = ?", c.Param("id")).Error != nil {
		c.JSON(404, gin.H{"error": "Cuenta no encontrada"})
		return
	}
	if user.ID == c.GetUint("adminID") {
		c.JSON(400, gin.H{"error": "No puedes desactivar tu propia cuenta"})
		return
	}
	err := h.db.WithContext(c.Request.Context()).Transaction(func(tx *gorm.DB) error {
		// Serialize changes to keep at least one active administrator under concurrent requests.
		var users []models.AdminUser
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Order("id").Find(&users).Error; err != nil {
			return err
		}
		active := 0
		for _, u := range users {
			if u.Active && u.ID != user.ID {
				active++
			}
		}
		if active == 0 {
			return errors.New("last administrator")
		}
		if err := tx.Model(&user).Update("active", false).Error; err != nil {
			return err
		}
		return tx.Where("user_id = ?", user.ID).Delete(&models.AdminSession{}).Error
	})
	if err != nil {
		c.JSON(409, gin.H{"error": "No se pudo desactivar la cuenta; debe quedar un administrador activo"})
		return
	}
	c.Status(204)
}

func (h *AuthHandler) ChangePassword(c *gin.Context) {
	var req struct {
		Current  string `json:"current"`
		Password string `json:"password"`
	}
	var user models.AdminUser
	if c.ShouldBindJSON(&req) != nil || len(req.Password) < 12 || len(req.Password) > 72 || len(req.Current) > 72 {
		c.JSON(400, gin.H{"error": "La contraseña debe tener entre 12 y 72 bytes"})
		return
	}
	if h.db.WithContext(c.Request.Context()).First(&user, c.GetUint("adminID")).Error != nil || bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Current)) != nil {
		c.JSON(400, gin.H{"error": "La contraseña actual no coincide"})
		return
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), 12)
	if err != nil {
		c.JSON(500, gin.H{"error": "No se pudo cambiar la contraseña"})
		return
	}
	err = h.db.WithContext(c.Request.Context()).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&user).Update("password_hash", string(hash)).Error; err != nil {
			return err
		}
		return tx.Where("user_id = ?", user.ID).Delete(&models.AdminSession{}).Error
	})
	if err != nil {
		c.JSON(500, gin.H{"error": "No se pudo cambiar la contraseña"})
		return
	}
	c.Status(204)
}
