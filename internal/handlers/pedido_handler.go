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
	Lista     string  `json:"lista"`
}

type savePedidoRequest struct {
	MonthlyListID uint                `json:"monthly_list_id" binding:"required"`
	Items         []pedidoItemRequest `json:"items"`
	UserID        *uint               `json:"user_id"`
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

// GetAll returns all pedido items for a list (admin use).
func (h *PedidoHandler) GetAll(c *gin.Context) {
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

	type row struct {
		models.PedidoItem
		Usuario string `json:"usuario"`
		Nombre  string `json:"nombre"`
	}

	var items []models.PedidoItem
	if err := h.DB.Where("monthly_list_id = ?", listaID).Find(&items).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error interno"})
		return
	}

	// Collect unique user IDs
	userIDs := make([]uint, 0)
	seen := map[uint]bool{}
	for _, it := range items {
		if !seen[it.UserID] {
			userIDs = append(userIDs, it.UserID)
			seen[it.UserID] = true
		}
	}

	var users []models.User
	if len(userIDs) > 0 {
		h.DB.Where("id IN ?", userIDs).Find(&users)
	}
	userMap := map[uint]models.User{}
	for _, u := range users {
		userMap[u.ID] = u
	}

	result := make([]row, 0, len(items))
	for _, it := range items {
		u := userMap[it.UserID]
		result = append(result, row{PedidoItem: it, Usuario: u.Usuario, Nombre: u.Nombre})
	}

	c.JSON(http.StatusOK, result)
}

func (h *PedidoHandler) Save(c *gin.Context) {
	user := middleware.CurrentUser(c)

	var req savePedidoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "monthly_list_id requerido"})
		return
	}

	targetUserID := user.ID
	if req.UserID != nil && *req.UserID != user.ID {
		if user.Role != "admin" {
			c.JSON(http.StatusForbidden, gin.H{"error": "no autorizado para editar el pedido de otro usuario"})
			return
		}
		targetUserID = *req.UserID
	}

	// Verify list exists and is not cerrado
	var list models.MonthlyList
	if err := h.DB.First(&list, req.MonthlyListID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "lista no encontrada"})
		return
	}
	if list.Cerrado && user.Role != "admin" {
		c.JSON(http.StatusConflict, gin.H{"error": "lista cerrada, no se puede modificar"})
		return
	}

	err := h.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("monthly_list_id = ? AND user_id = ?", req.MonthlyListID, targetUserID).
			Delete(&models.PedidoItem{}).Error; err != nil {
			return err
		}
		for _, it := range req.Items {
			if it.Cantidad <= 0 {
				continue
			}
			lista := it.Lista
			if lista != "extra" {
				lista = "bono"
			}
			item := models.PedidoItem{
				MonthlyListID: req.MonthlyListID,
				UserID:        targetUserID,
				ProductID:     it.ProductID,
				Lista:         lista,
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
	h.DB.Where("monthly_list_id = ? AND user_id = ?", req.MonthlyListID, targetUserID).Find(&saved)
	c.JSON(http.StatusOK, saved)
}
