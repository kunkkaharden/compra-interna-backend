package db

import (
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"

	"github.com/kada/compra-interna-backend/internal/models"
)

func Open(path string) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(path), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// idx_pedido used to cover (monthly_list_id, user_id, product_id) only; it now
	// also includes "lista" so the same product can have separate bono/extra lines.
	// Drop the old narrower index so AutoMigrate recreates it with the new column.
	if err := db.Exec("DROP INDEX IF EXISTS idx_pedido").Error; err != nil {
		return nil, err
	}

	if err := db.AutoMigrate(
		&models.User{},
		&models.Product{},
		&models.MonthlyList{},
		&models.MonthlyListItem{},
		&models.PedidoItem{},
	); err != nil {
		return nil, err
	}

	return db, nil
}
