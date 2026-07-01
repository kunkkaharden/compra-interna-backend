package handlers

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/kada/compra-interna-backend/internal/auth"
	"github.com/kada/compra-interna-backend/internal/models"
)

type UserHandler struct {
	DB *gorm.DB
}

type createUserRequest struct {
	Usuario     string `json:"usuario" binding:"required"`
	Nombre      string `json:"nombre"`
	Contrasenna string `json:"contrasenna" binding:"required"`
	IsActive    *bool  `json:"isactive"`
}

func (h *UserHandler) List(c *gin.Context) {
	var users []models.User
	if err := h.DB.Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error interno"})
		return
	}
	c.JSON(http.StatusOK, users)
}

func (h *UserHandler) Create(c *gin.Context) {
	var req createUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "usuario y contrasenna son requeridos"})
		return
	}

	hash, err := auth.HashPassword(req.Contrasenna)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error interno"})
		return
	}

	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	user := models.User{
		Usuario:     req.Usuario,
		Nombre:      req.Nombre,
		Contrasenna: hash,
		IsActive:    isActive,
	}

	if err := h.DB.Create(&user).Error; err != nil {
		if isUniqueConstraintErr(err) {
			c.JSON(http.StatusConflict, gin.H{"error": "usuario ya existe"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error interno"})
		return
	}

	c.JSON(http.StatusCreated, user)
}

func (h *UserHandler) Get(c *gin.Context) {
	var user models.User
	if err := h.DB.First(&user, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "usuario no encontrado"})
		return
	}
	c.JSON(http.StatusOK, user)
}

type updateUserRequest struct {
	Usuario     *string `json:"usuario"`
	Nombre      *string `json:"nombre"`
	Contrasenna *string `json:"contrasenna"`
	IsActive    *bool   `json:"isactive"`
}

func (h *UserHandler) Update(c *gin.Context) {
	var user models.User
	if err := h.DB.First(&user, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "usuario no encontrado"})
		return
	}

	var req updateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "body inválido"})
		return
	}

	if req.Usuario != nil {
		user.Usuario = *req.Usuario
	}
	if req.Nombre != nil {
		user.Nombre = *req.Nombre
	}
	if req.IsActive != nil {
		user.IsActive = *req.IsActive
	}
	if req.Contrasenna != nil {
		hash, err := auth.HashPassword(*req.Contrasenna)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error interno"})
			return
		}
		user.Contrasenna = hash
	}

	if err := h.DB.Save(&user).Error; err != nil {
		if isUniqueConstraintErr(err) {
			c.JSON(http.StatusConflict, gin.H{"error": "usuario ya existe"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error interno"})
		return
	}

	c.JSON(http.StatusOK, user)
}

func (h *UserHandler) Delete(c *gin.Context) {
	result := h.DB.Delete(&models.User{}, c.Param("id"))
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error interno"})
		return
	}
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "usuario no encontrado"})
		return
	}
	c.Status(http.StatusNoContent)
}

func isUniqueConstraintErr(err error) bool {
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return true
	}
	return strings.Contains(strings.ToLower(err.Error()), "unique")
}
