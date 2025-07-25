package cmd

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// setupStaticRoutes configures routes for serving the embedded frontend
func setupStaticRoutes(r *gin.Engine) {
	// Get embedded filesystem
	staticFS, err := GetFrontendFS()
	if err != nil {
		panic("Failed to get embedded frontend filesystem: " + err.Error())
	}

	// Serve static files from embedded filesystem
	r.StaticFS("/static", http.FS(staticFS))

	// Handle SPA routing - serve index.html for all non-API routes
	r.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path

		// Skip API routes and proxy routes
		if path == "/" || (!isAPIRoute(path) && !isProxyRoute(path)) {
			// Try to serve the file from embedded FS first
			if file, err := staticFS.Open(path[1:]); err == nil {
				file.Close()
				c.FileFromFS(path[1:], http.FS(staticFS))
				return
			}

			// Fall back to index.html for SPA routing
			c.FileFromFS("index.html", http.FS(staticFS))
			return
		}

		// For unmatched API routes, return 404
		c.JSON(404, gin.H{"error": "Not found"})
	})
}

// isAPIRoute checks if the path is an API route
func isAPIRoute(path string) bool {
	return len(path) >= 4 && path[:4] == "/api"
}

// isProxyRoute checks if the path is a proxy route
func isProxyRoute(path string) bool {
	return len(path) >= 6 && path[:6] == "/proxy"
}
