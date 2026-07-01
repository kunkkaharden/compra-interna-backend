package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/kada/compra-interna-backend/internal/handlers"
)

func newProductRouter(db *gorm.DB) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	h := &handlers.ProductHandler{DB: db}
	router.GET("/api/products", h.List)
	router.POST("/api/products", h.Create)
	router.PUT("/api/products/:id", h.Update)
	router.DELETE("/api/products/:id", h.Delete)
	return router
}

func jsonRequest(method, path string, body any) *http.Request {
	var buf bytes.Buffer
	if body != nil {
		json.NewEncoder(&buf).Encode(body)
	}
	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	return req
}

func TestProduct_CreateAndList(t *testing.T) {
	db := setupTestDB(t)
	router := newProductRouter(db)

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, jsonRequest(http.MethodPost, "/api/products", map[string]string{
		"codigo_tkc": "TK001",
		"nombre":     "Arroz",
	}))
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}

	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/products", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var products []map[string]any
	json.Unmarshal(rec.Body.Bytes(), &products)
	if len(products) != 1 {
		t.Fatalf("expected 1 product, got %d", len(products))
	}
}

func TestProduct_DuplicateCodigoTkc(t *testing.T) {
	db := setupTestDB(t)
	router := newProductRouter(db)

	router.ServeHTTP(httptest.NewRecorder(), jsonRequest(http.MethodPost, "/api/products", map[string]string{
		"codigo_tkc": "TK001",
		"nombre":     "Arroz",
	}))

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, jsonRequest(http.MethodPost, "/api/products", map[string]string{
		"codigo_tkc": "TK001",
		"nombre":     "Otro",
	}))
	if rec.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestProduct_UpdateNotFound(t *testing.T) {
	db := setupTestDB(t)
	router := newProductRouter(db)

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, jsonRequest(http.MethodPut, "/api/products/999", map[string]string{
		"codigo_tkc": "TK999",
		"nombre":     "Fantasma",
	}))
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestProduct_DeleteNotFound(t *testing.T) {
	db := setupTestDB(t)
	router := newProductRouter(db)

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodDelete, "/api/products/999", nil))
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}
