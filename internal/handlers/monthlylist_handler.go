package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/kada/compra-interna-backend/internal/models"
)

type MonthlyListHandler struct {
	DB *gorm.DB
}

type monthlyListItemRequest struct {
	ProductID  uint    `json:"product_id" binding:"required"`
	CodigoTkc  string  `json:"codigo_tkc"`
	Nombre     string  `json:"nombre"`
	PrecioUsd  float64 `json:"precio_usd"`
	Stock      int     `json:"stock"`
	LimitAdmin int     `json:"limit_admin"`
	Extra      bool    `json:"extra"`
}

type monthlyListRequest struct {
	Mes     int                      `json:"mes" binding:"required"`
	Year    int                      `json:"year" binding:"required"`
	BudgetA float64                  `json:"budget_a"`
	Items   []monthlyListItemRequest `json:"items"`
}

func (h *MonthlyListHandler) List(c *gin.Context) {
	query := h.DB.Preload("Items")

	if mesRaw := c.Query("mes"); mesRaw != "" {
		mes, err := strconv.Atoi(mesRaw)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "mes inválido"})
			return
		}
		query = query.Where("mes = ?", mes)
	}
	if yearRaw := c.Query("year"); yearRaw != "" {
		year, err := strconv.Atoi(yearRaw)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "year inválido"})
			return
		}
		query = query.Where("year = ?", year)
	}

	var lists []models.MonthlyList
	if err := query.Find(&lists).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error interno"})
		return
	}
	c.JSON(http.StatusOK, lists)
}

func (h *MonthlyListHandler) Get(c *gin.Context) {
	var list models.MonthlyList
	if err := h.DB.Preload("Items").First(&list, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "lista no encontrada"})
		return
	}
	c.JSON(http.StatusOK, list)
}

// Create hace upsert por (mes, year): si ya existe una lista para ese
// mes/año, reemplaza sus items; si no, crea una nueva.
func (h *MonthlyListHandler) Create(c *gin.Context) {
	var req monthlyListRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "mes y year son requeridos"})
		return
	}

	items := toItemModels(req.Items)

	var list models.MonthlyList
	err := h.DB.Transaction(func(tx *gorm.DB) error {
		lookupErr := tx.Where("mes = ? AND year = ?", req.Mes, req.Year).First(&list).Error
		switch {
		case errors.Is(lookupErr, gorm.ErrRecordNotFound):
			list = models.MonthlyList{Mes: req.Mes, Year: req.Year, BudgetA: req.BudgetA, Items: items}
			return tx.Create(&list).Error
		case lookupErr != nil:
			return lookupErr
		}

		if list.Cerrado {
			return gorm.ErrInvalidData
		}
		if err := tx.Where("monthly_list_id = ?", list.ID).Delete(&models.MonthlyListItem{}).Error; err != nil {
			return err
		}
		list.BudgetA = req.BudgetA
		list.Items = items
		return tx.Save(&list).Error
	})

	if errors.Is(err, gorm.ErrInvalidData) {
		c.JSON(http.StatusConflict, gin.H{"error": "lista cerrada, no se puede modificar"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error interno"})
		return
	}

	h.DB.Preload("Items").First(&list, list.ID)
	c.JSON(http.StatusCreated, list)
}

func (h *MonthlyListHandler) Close(c *gin.Context) {
	result := h.DB.Model(&models.MonthlyList{}).Where("id = ?", c.Param("id")).Update("cerrado", true)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error interno"})
		return
	}
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "lista no encontrada"})
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *MonthlyListHandler) Update(c *gin.Context) {
	var list models.MonthlyList
	if err := h.DB.First(&list, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "lista no encontrada"})
		return
	}

	var req monthlyListRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "mes y year son requeridos"})
		return
	}

	err := h.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("monthly_list_id = ?", list.ID).Delete(&models.MonthlyListItem{}).Error; err != nil {
			return err
		}
		list.Mes = req.Mes
		list.Year = req.Year
		list.BudgetA = req.BudgetA
		list.Items = toItemModels(req.Items)
		return tx.Save(&list).Error
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error interno"})
		return
	}

	h.DB.Preload("Items").First(&list, list.ID)
	c.JSON(http.StatusOK, list)
}

func (h *MonthlyListHandler) Delete(c *gin.Context) {
	result := h.DB.Delete(&models.MonthlyList{}, c.Param("id"))
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error interno"})
		return
	}
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "lista no encontrada"})
		return
	}
	c.Status(http.StatusNoContent)
}

func toItemModels(reqItems []monthlyListItemRequest) []models.MonthlyListItem {
	items := make([]models.MonthlyListItem, 0, len(reqItems))
	for _, it := range reqItems {
		items = append(items, models.MonthlyListItem{
			ProductID:  it.ProductID,
			CodigoTkc:  it.CodigoTkc,
			Nombre:     it.Nombre,
			PrecioUsd:  it.PrecioUsd,
			Stock:      it.Stock,
			LimitAdmin: it.LimitAdmin,
			Extra:      it.Extra,
		})
	}
	return items
}
