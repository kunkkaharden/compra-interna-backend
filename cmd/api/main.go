package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/kada/compra-interna-backend/internal/config"
	"github.com/kada/compra-interna-backend/internal/db"
	"github.com/kada/compra-interna-backend/internal/handlers"
	"github.com/kada/compra-interna-backend/internal/middleware"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config error: %v", err)
	}

	gormDB, err := db.Open(cfg.DBPath)
	if err != nil {
		log.Fatalf("db error: %v", err)
	}

	router := gin.Default()
	router.Use(corsMiddleware())

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	authHandler := &handlers.AuthHandler{DB: gormDB, JWTSecret: cfg.JWTSecret, JWTExpiryHours: cfg.JWTExpiryHours}
	userHandler := &handlers.UserHandler{DB: gormDB}
	productHandler := &handlers.ProductHandler{DB: gormDB}
	monthlyListHandler := &handlers.MonthlyListHandler{DB: gormDB}
	pedidoHandler := &handlers.PedidoHandler{DB: gormDB}

	api := router.Group("/api")
	api.POST("/auth/login", authHandler.Login)

	protected := api.Group("")
	protected.Use(middleware.RequireAuth(cfg.JWTSecret, gormDB))
	{
		protected.GET("/auth/me", authHandler.Me)

		protected.GET("/products", productHandler.List)
		protected.GET("/monthly-lists", monthlyListHandler.List)
		protected.GET("/monthly-lists/:id", monthlyListHandler.Get)

		protected.GET("/pedidos", pedidoHandler.Get)
		protected.POST("/pedidos", pedidoHandler.Save)
		protected.GET("/users/count", userHandler.Count)
	}

	admin := api.Group("")
	admin.Use(middleware.RequireAuth(cfg.JWTSecret, gormDB), middleware.RequireAdmin())
	{
		admin.GET("/users", userHandler.List)
		admin.POST("/users", userHandler.Create)
		admin.GET("/users/:id", userHandler.Get)
		admin.PUT("/users/:id", userHandler.Update)
		admin.DELETE("/users/:id", userHandler.Delete)

		admin.POST("/products", productHandler.Create)
		admin.PUT("/products/:id", productHandler.Update)
		admin.PATCH("/products/:id/archive", productHandler.Archive)
		admin.PATCH("/products/:id/unarchive", productHandler.Unarchive)

		admin.POST("/monthly-lists", monthlyListHandler.Create)
		admin.PUT("/monthly-lists/:id", monthlyListHandler.Update)
		admin.DELETE("/monthly-lists/:id", monthlyListHandler.Delete)
		admin.PATCH("/monthly-lists/:id/close", monthlyListHandler.Close)
		admin.PATCH("/monthly-lists/:id/reopen", monthlyListHandler.Reopen)

		admin.GET("/pedidos/all", pedidoHandler.GetAll)
	}

	log.Printf("listening on :%s", cfg.Port)
	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Authorization, Content-Type")
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}
