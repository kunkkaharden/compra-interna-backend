package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/kada/compra-interna-backend/internal/auth"
	"github.com/kada/compra-interna-backend/internal/middleware"
	"github.com/kada/compra-interna-backend/internal/models"
)

type AuthHandler struct {
	DB             *gorm.DB
	JWTSecret      string
	JWTExpiryHours int
}

type loginRequest struct {
	Usuario     string `json:"usuario" binding:"required"`
	Contrasenna string `json:"contrasenna" binding:"required"`
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "usuario y contrasenna son requeridos"})
		return
	}

	var user models.User
	if err := h.DB.Where("usuario = ?", req.Usuario).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "credenciales inválidas"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error interno"})
		return
	}

	if !user.IsActive {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "usuario inactivo"})
		return
	}

	if !auth.CheckPassword(user.Contrasenna, req.Contrasenna) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "credenciales inválidas"})
		return
	}

	token, err := auth.GenerateToken(h.JWTSecret, user.ID, h.JWTExpiryHours)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error generando token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}

func (h *AuthHandler) Me(c *gin.Context) {
	user := middleware.CurrentUser(c)
	c.JSON(http.StatusOK, user)
}
