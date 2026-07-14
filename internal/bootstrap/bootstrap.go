package bootstrap
package bootstrap

import (
    "errors"
    "log"
    "os"

    "github.com/kada/compra-interna-backend/internal/auth"
    "github.com/kada/compra-interna-backend/internal/models"
    "gorm.io/gorm"
)

func EnsureDefaultAdmin(gormDB *gorm.DB) error {
    adminUser := os.Getenv("ADMIN_USER")
    if adminUser == "" {
        adminUser = "admin"
    }
    adminPass := os.Getenv("ADMIN_PASS")
    if adminPass == "" {
        adminPass = "admin"
    }

    var existing models.User
    lookupErr := gormDB.Where("role = ?", "admin").First(&existing).Error
    switch {
    case errors.Is(lookupErr, gorm.ErrRecordNotFound):
        hash, err := auth.HashPassword(adminPass)
        if err != nil {
            return err
        }
        user := models.User{Usuario: adminUser, Contrasenna: hash, Role: "admin", IsActive: true}
        if err := gormDB.Create(&user).Error; err != nil {
            return err
        }
        log.Printf("default admin created: %s", adminUser)
    case lookupErr != nil:
        return lookupErr
    default:
        // Admin exists, do nothing
    }
    return nil
}
