package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/kada/compra-interna-backend/internal/models"
)

type ProductHandler struct {
	DB *gorm.DB
}

type productRequest struct {
	CodigoTkc string `json:"codigo_tkc" binding:"required"`
	Nombre    string `json:"nombre" binding:"required"`
}

func (h *ProductHandler) List(c *gin.Context) {
	var products []models.Product
	if err := h.DB.Where("archived = ?", false).Find(&products).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error interno"})
		return
	}
	c.JSON(http.StatusOK, products)
}

func (h *ProductHandler) Create(c *gin.Context) {
	var req productRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "codigo_tkc y nombre son requeridos"})
		return
	}

	product := models.Product{CodigoTkc: req.CodigoTkc, Nombre: req.Nombre}
	if err := h.DB.Create(&product).Error; err != nil {
		if isUniqueConstraintErr(err) {
			c.JSON(http.StatusConflict, gin.H{"error": "codigo_tkc ya existe"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error interno"})
		return
	}

	c.JSON(http.StatusCreated, product)
}

func (h *ProductHandler) Update(c *gin.Context) {
	var product models.Product
	if err := h.DB.First(&product, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "producto no encontrado"})
		return
	}

	var req productRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "codigo_tkc y nombre son requeridos"})
		return
	}

	product.CodigoTkc = req.CodigoTkc
	product.Nombre = req.Nombre

	if err := h.DB.Save(&product).Error; err != nil {
		if isUniqueConstraintErr(err) {
			c.JSON(http.StatusConflict, gin.H{"error": "codigo_tkc ya existe"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error interno"})
		return
	}

	c.JSON(http.StatusOK, product)
}

func (h *ProductHandler) Archive(c *gin.Context) {
	result := h.DB.Model(&models.Product{}).Where("id = ?", c.Param("id")).Update("archived", true)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error interno"})
		return
	}
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "producto no encontrado"})
		return
	}
	c.Status(http.StatusNoContent)
}
