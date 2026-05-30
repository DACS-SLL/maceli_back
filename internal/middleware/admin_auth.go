package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func AdminAuth(adminKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := c.GetHeader("X-ADMIN-KEY")
		if key == "" || key != adminKey {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "No autorizado. Verifica el header X-ADMIN-KEY",
			})
			return
		}

		c.Next()
	}
}
