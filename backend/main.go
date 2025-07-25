package main

import (
	"log"

	"lan-relay/internal/config"
	"lan-relay/internal/database"
	"lan-relay/internal/handlers"
	"lan-relay/internal/logger"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize logger
	logger.Init(cfg.LogLevel)

	// Initialize database
	db, err := database.Init(cfg.DatabasePath)
	if err != nil {
		log.Fatal("Failed to initialize database:", err)
	}
	defer db.Close()

	// Initialize Gin router
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.Default()

	// CORS middleware for React frontend
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000", "http://127.0.0.1:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "HEAD", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", "Authorization"},
		AllowCredentials: true,
	}))

	// Initialize handlers
	h := handlers.New(db)

	// Routes
	api := r.Group("/api")
	{
		api.GET("/health", h.HealthCheck)
		api.GET("/status", h.GetStatus)
		api.GET("/logs", h.GetLogs)
		api.POST("/logs/clear", h.ClearLogs)

		// Settings routes
		api.GET("/settings", h.GetSettings)
		api.POST("/settings", h.UpdateSettings)

		// Ngrok routes
		api.POST("/ngrok/start", h.StartNgrokTunnel)
		api.POST("/ngrok/stop", h.StopNgrokTunnel)
		api.POST("/ngrok/test", h.TestNgrokToken)
	}

	// Proxy routes - catch-all for proxy requests
	r.Any("/proxy/*path", h.ProxyRequest)

	// Start server
	port := cfg.Port
	if port == "" {
		port = "8080"
	}

	logger.Info("Starting LAN Relay Server on port " + port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
