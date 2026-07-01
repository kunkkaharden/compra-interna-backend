package models

import "time"

type User struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Usuario     string    `gorm:"uniqueIndex;not null" json:"usuario"`
	Contrasenna string    `gorm:"not null" json:"-"`
	IsActive    bool      `gorm:"not null" json:"isactive"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Product struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	CodigoTkc string    `gorm:"uniqueIndex;not null" json:"codigo_tkc"`
	Nombre    string    `gorm:"not null" json:"nombre"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type MonthlyList struct {
	ID        uint              `gorm:"primaryKey" json:"id"`
	Mes       int               `gorm:"not null;uniqueIndex:idx_mes_year" json:"mes"`
	Year      int               `gorm:"not null;uniqueIndex:idx_mes_year" json:"year"`
	BudgetA   float64           `json:"budget_a"`
	Items     []MonthlyListItem `json:"items"`
	CreatedAt time.Time         `json:"created_at"`
	UpdatedAt time.Time         `json:"updated_at"`
}

type MonthlyListItem struct {
	ID            uint    `gorm:"primaryKey" json:"id"`
	MonthlyListID uint    `gorm:"not null;index" json:"monthly_list_id"`
	ProductID     uint    `gorm:"not null;index" json:"product_id"`
	CodigoTkc     string  `json:"codigo_tkc"`
	Nombre        string  `json:"nombre"`
	PrecioUsd     float64 `json:"precio_usd"`
	Stock         int     `json:"stock"`
	LimitAdmin    int     `json:"limit_admin"`
	Extra         bool    `json:"extra"`
}
