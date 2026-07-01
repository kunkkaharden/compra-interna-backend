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

	if err := db.AutoMigrate(
		&models.User{},
		&models.Product{},
		&models.MonthlyList{},
		&models.MonthlyListItem{},
	); err != nil {
		return nil, err
	}

	return db, nil
}
