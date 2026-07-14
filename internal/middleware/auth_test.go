package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/kada/compra-interna-backend/internal/auth"
	"github.com/kada/compra-interna-backend/internal/middleware"
	"github.com/kada/compra-interna-backend/internal/models"
)

const testSecret = "test-secret"

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
	if err := db.AutoMigrate(&models.User{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return db
}

func newProtectedRouter(db *gorm.DB) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/protected", middleware.RequireAuth(testSecret, db), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})
	return router
}

func doRequest(router *gin.Engine, token string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}

func TestRequireAuth_ValidToken(t *testing.T) {
	db := setupTestDB(t)
	user := models.User{Usuario: "admin", Contrasenna: "hash", IsActive: true}
	db.Create(&user)

	token, err := auth.GenerateToken(testSecret, user.ID, 1)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}

	rec := doRequest(newProtectedRouter(db), token)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestRequireAuth_MissingToken(t *testing.T) {
	db := setupTestDB(t)
	rec := doRequest(newProtectedRouter(db), "")
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestRequireAuth_InvalidToken(t *testing.T) {
	db := setupTestDB(t)
	rec := doRequest(newProtectedRouter(db), "not-a-real-token")
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestRequireAuth_ExpiredToken(t *testing.T) {
	db := setupTestDB(t)
	user := models.User{Usuario: "admin", Contrasenna: "hash", IsActive: true}
	db.Create(&user)

	token, err := auth.GenerateToken(testSecret, user.ID, -1)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}
	time.Sleep(10 * time.Millisecond)

	rec := doRequest(newProtectedRouter(db), token)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestRequireAuth_InactiveUser(t *testing.T) {
	db := setupTestDB(t)
	user := models.User{Usuario: "admin", Contrasenna: "hash", IsActive: false}
	db.Create(&user)

	token, err := auth.GenerateToken(testSecret, user.ID, 1)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}

	rec := doRequest(newProtectedRouter(db), token)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}
