package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"

	"gorm.io/gorm"

	"github.com/kada/compra-interna-backend/internal/auth"
	"github.com/kada/compra-interna-backend/internal/db"
	"github.com/kada/compra-interna-backend/internal/models"
)

func main() {
	usuario := flag.String("usuario", "", "nombre de usuario a crear")
	contrasenna := flag.String("contrasenna", "", "contraseña en texto plano")
	dbPath := flag.String("db", "", "ruta al archivo sqlite (default: DB_PATH env var o compra_interna.db)")
	flag.Parse()

	if *usuario == "" || *contrasenna == "" {
		fmt.Fprintln(os.Stderr, "uso: seed -usuario=<usuario> -contrasenna=<contrasenna>")
		os.Exit(1)
	}

	path := *dbPath
	if path == "" {
		path = os.Getenv("DB_PATH")
	}
	if path == "" {
		path = "compra_interna.db"
	}

	gormDB, err := db.Open(path)
	if err != nil {
		log.Fatalf("db error: %v", err)
	}

	hash, err := auth.HashPassword(*contrasenna)
	if err != nil {
		log.Fatalf("hash error: %v", err)
	}

	var existing models.User
	lookupErr := gormDB.Where("usuario = ?", *usuario).First(&existing).Error
	switch {
	case errors.Is(lookupErr, gorm.ErrRecordNotFound):
		user := models.User{Usuario: *usuario, Contrasenna: hash, IsActive: true}
		if err := gormDB.Create(&user).Error; err != nil {
			log.Fatalf("create error: %v", err)
		}
		fmt.Printf("usuario %q creado\n", *usuario)
	case lookupErr != nil:
		log.Fatalf("lookup error: %v", lookupErr)
	default:
		existing.Contrasenna = hash
		existing.IsActive = true
		if err := gormDB.Save(&existing).Error; err != nil {
			log.Fatalf("update error: %v", err)
		}
		fmt.Printf("usuario %q actualizado\n", *usuario)
	}
}
