package middleware

import (
	"context"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type bucket struct {
	count int
	until time.Time
}

// Keep route identifiers out of GORM's string-expression overloads.
func NumericID() gin.HandlerFunc {
	return func(c *gin.Context) {
		if value := c.Param("id"); value != "" {
			id, err := strconv.ParseUint(value, 10, 32)
			if err != nil || id == 0 {
				c.AbortWithStatusJSON(400, gin.H{"error": "Identificador inválido"})
				return
			}
		}
		c.Next()
	}
}

// Bounded memory and both per-client and global limits, including when proxy IPs are shared.
func RateLimit(limit, globalLimit int, window time.Duration) gin.HandlerFunc {
	var mu sync.Mutex
	clients := make(map[string]bucket)
	global := bucket{}
	nextSweep := time.Now().Add(window)
	return func(c *gin.Context) {
		now := time.Now()
		key := c.ClientIP()
		mu.Lock()
		if now.After(nextSweep) {
			for k, v := range clients {
				if now.After(v.until) {
					delete(clients, k)
				}
			}
			nextSweep = now.Add(window)
		}
		if now.After(global.until) {
			global = bucket{until: now.Add(window)}
		}
		value, known := clients[key]
		if now.After(value.until) {
			value = bucket{until: now.Add(window)}
		}
		allowed := global.count < globalLimit && value.count < limit && (known || len(clients) < 10000)
		if allowed {
			global.count++
			value.count++
			clients[key] = value
		}
		mu.Unlock()
		if !allowed {
			c.Header("Retry-After", strconv.Itoa(int(window.Seconds())))
			c.AbortWithStatusJSON(429, gin.H{"error": "Demasiadas solicitudes. Espera un momento e intenta de nuevo."})
			return
		}
		c.Next()
	}
}

func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Header("Cache-Control", "no-store")
		c.Next()
	}
}

func BodyLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 35*time.Second)
		defer cancel()
		c.Request = c.Request.WithContext(ctx)
		limit := int64(128 << 10)
		if c.ContentType() == "multipart/form-data" {
			limit = 9 << 20
		}
		if c.Request.ContentLength > limit {
			c.AbortWithStatusJSON(413, gin.H{"error": "El archivo o solicitud supera el tamaño permitido"})
			return
		}
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, limit)
		c.Next()
		if c.Request.MultipartForm != nil {
			_ = c.Request.MultipartForm.RemoveAll()
		}
	}
}
