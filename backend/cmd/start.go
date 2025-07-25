package cmd

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"lan-relay/internal/config"
	"lan-relay/internal/database"
	"lan-relay/internal/handlers"
	"lan-relay/internal/logger"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
)

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the LAN Relay server",
	Long: `Start the LAN Relay server with embedded web dashboard.
The server will listen on the specified port and provide both
API endpoints and a web interface for managing proxy connections.`,
	Run: func(cmd *cobra.Command, args []string) {
		startServer()
	},
}

func init() {
	rootCmd.AddCommand(startCmd)
}

func startServer() {
	// Load configuration
	cfg := config.Load()

	// Override port if specified via CLI
	if port != "" {
		cfg.Port = port
	}

	// Initialize logger
	logger.Init(cfg.LogLevel)

	// Initialize database
	db, err := database.Init(cfg.DatabasePath)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to initialize database: %v", err))
		os.Exit(1)
	}
	defer db.Close()

	// Initialize Gin router
	if cfg.Environment == "production" || daemon {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.Default()

	// CORS middleware
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "HEAD", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", "Authorization"},
		AllowCredentials: true,
	}))

	// Initialize handlers
	h := handlers.New(db)

	// API routes
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

	// Serve embedded frontend
	setupStaticRoutes(r)

	// Create HTTP server
	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: r,
	}

	// Start server in goroutine
	go func() {
		logger.Info(fmt.Sprintf("🚀 LAN Relay Server starting on port %s", cfg.Port))
		logger.Info(fmt.Sprintf("📊 Dashboard: http://localhost:%s", cfg.Port))
		logger.Info(fmt.Sprintf("🔧 API: http://localhost:%s/api", cfg.Port))

		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error(fmt.Sprintf("Failed to start server: %v", err))
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("🛑 Shutting down server...")

	// Create a deadline for shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Attempt graceful shutdown
	if err := srv.Shutdown(ctx); err != nil {
		logger.Error(fmt.Sprintf("Server forced to shutdown: %v", err))
	}

	logger.Info("✅ Server stopped")
}
