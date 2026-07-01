package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/kada/compra-interna-backend/internal/auth"
	"github.com/kada/compra-interna-backend/internal/models"
)

const ContextUserKey = "currentUser"

func RequireAuth(secret string, db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "falta token de autorización"})
			return
		}

		claims, err := auth.ParseToken(secret, parts[1])
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "token inválido o expirado"})
			return
		}

		var user models.User
		if err := db.First(&user, claims.UserID).Error; err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "usuario no encontrado"})
			return
		}
		if !user.IsActive {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "usuario inactivo"})
			return
		}

		c.Set(ContextUserKey, user)
		c.Next()
	}
}

func CurrentUser(c *gin.Context) models.User {
	return c.MustGet(ContextUserKey).(models.User)
}
