package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/kada/compra-interna-backend/internal/auth"
	"github.com/kada/compra-interna-backend/internal/handlers"
	"github.com/kada/compra-interna-backend/internal/models"
)

func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		t.Fatal("DATABASE_URL env var is required for tests")
	}

	db, err := gorm.Open(postgres.Open(databaseURL), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := db.AutoMigrate(&models.User{}, &models.Product{}, &models.MonthlyList{}, &models.MonthlyListItem{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return db
}

func newLoginRouter(t *testing.T, db *gorm.DB) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	router := gin.New()
	h := &handlers.AuthHandler{DB: db, JWTSecret: "test-secret", JWTExpiryHours: 1}
	router.POST("/api/auth/login", h.Login)
	return router
}

func doLogin(router *gin.Engine, usuario, contrasenna string) *httptest.ResponseRecorder {
	body, _ := json.Marshal(map[string]string{"usuario": usuario, "contrasenna": contrasenna})
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}

func TestLogin_ValidCredentials(t *testing.T) {
	db := setupTestDB(t)
	hash, err := auth.HashPassword("secret123")
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	db.Create(&models.User{Usuario: "admin", Contrasenna: hash, IsActive: true})

	router := newLoginRouter(t, db)
	rec := doLogin(router, "admin", "secret123")

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var resp map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if resp["token"] == "" {
		t.Fatal("expected non-empty token")
	}
}

func TestLogin_InvalidPassword(t *testing.T) {
	db := setupTestDB(t)
	hash, _ := auth.HashPassword("secret123")
	db.Create(&models.User{Usuario: "admin", Contrasenna: hash, IsActive: true})

	router := newLoginRouter(t, db)
	rec := doLogin(router, "admin", "wrong-password")

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestLogin_UnknownUser(t *testing.T) {
	db := setupTestDB(t)

	router := newLoginRouter(t, db)
	rec := doLogin(router, "ghost", "whatever")

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestLogin_InactiveUser(t *testing.T) {
	db := setupTestDB(t)
	hash, _ := auth.HashPassword("secret123")
	db.Create(&models.User{Usuario: "admin", Contrasenna: hash, IsActive: false})

	router := newLoginRouter(t, db)
	rec := doLogin(router, "admin", "secret123")

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d: %s", rec.Code, rec.Body.String())
	}
}
