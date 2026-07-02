package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/kada/compra-interna-backend/internal/middleware"
	"github.com/kada/compra-interna-backend/internal/models"
)

type PedidoHandler struct {
	DB *gorm.DB
}

type pedidoItemRequest struct {
	ProductID uint    `json:"product_id" binding:"required"`
	Cantidad  int     `json:"cantidad" binding:"required"`
	PrecioUsd float64 `json:"precio_usd"`
}

type savePedidoRequest struct {
	MonthlyListID uint                `json:"monthly_list_id" binding:"required"`
	Items         []pedidoItemRequest `json:"items"`
}

func (h *PedidoHandler) Get(c *gin.Context) {
	user := middleware.CurrentUser(c)

	listaIDRaw := c.Query("lista_id")
	if listaIDRaw == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "lista_id requerido"})
		return
	}
	listaID, err := strconv.Atoi(listaIDRaw)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "lista_id inválido"})
		return
	}

	var items []models.PedidoItem
	if err := h.DB.Where("monthly_list_id = ? AND user_id = ?", listaID, user.ID).Find(&items).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error interno"})
		return
	}
	c.JSON(http.StatusOK, items)
}

func (h *PedidoHandler) Save(c *gin.Context) {
	user := middleware.CurrentUser(c)

	var req savePedidoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "monthly_list_id requerido"})
		return
	}

	// Verify list exists and is not cerrado
	var list models.MonthlyList
	if err := h.DB.First(&list, req.MonthlyListID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "lista no encontrada"})
		return
	}
	if list.Cerrado {
		c.JSON(http.StatusConflict, gin.H{"error": "lista cerrada, no se puede modificar"})
		return
	}

	err := h.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("monthly_list_id = ? AND user_id = ?", req.MonthlyListID, user.ID).
			Delete(&models.PedidoItem{}).Error; err != nil {
			return err
		}
		for _, it := range req.Items {
			if it.Cantidad <= 0 {
				continue
			}
			item := models.PedidoItem{
				MonthlyListID: req.MonthlyListID,
				UserID:        user.ID,
				ProductID:     it.ProductID,
				Cantidad:      it.Cantidad,
				PrecioUsd:     it.PrecioUsd,
			}
			if err := tx.Create(&item).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error interno"})
		return
	}

	var saved []models.PedidoItem
	h.DB.Where("monthly_list_id = ? AND user_id = ?", req.MonthlyListID, user.ID).Find(&saved)
	c.JSON(http.StatusOK, saved)
}
