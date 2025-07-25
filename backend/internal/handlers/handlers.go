package handlers

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"lan-relay/internal/database"
	"lan-relay/internal/logger"
	"lan-relay/internal/models"
	"lan-relay/internal/ngrok"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	db           *database.DB
	startTime    time.Time
	ngrokManager *ngrok.NgrokManager
	ngrokMutex   sync.Mutex
}

func New(db *database.DB) *Handler {
	return &Handler{
		db:        db,
		startTime: time.Now(),
	}
}

// ProxyRequest handles proxying HTTP requests to internal network targets
func (h *Handler) ProxyRequest(c *gin.Context) {
	start := time.Now()

	// Extract target from path: /proxy/HOST:PORT/path
	fullPath := c.Param("path")
	if fullPath == "" || !strings.HasPrefix(fullPath, "/") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid proxy path format. Use: /proxy/HOST:PORT/path"})
		return
	}

	// Remove leading slash and split
	cleanPath := strings.TrimPrefix(fullPath, "/")
	parts := strings.SplitN(cleanPath, "/", 2)
	if len(parts) < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid proxy path format. Use: /proxy/HOST:PORT/path"})
		return
	}

	hostPort := parts[0]
	targetPath := "/"
	if len(parts) > 1 {
		targetPath = "/" + parts[1]
	}

	// Parse host and port
	host, portStr, err := net.SplitHostPort(hostPort)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid host:port format"})
		return
	}

	port, err := strconv.Atoi(portStr)
	if err != nil || port < 1 || port > 65535 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid port number"})
		return
	}

	// Validate that the target is a private IP
	if !isPrivateIP(host) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only private IP addresses are allowed"})
		return
	}

	// Create target URL
	target := fmt.Sprintf("http://%s:%d%s", host, port, targetPath)

	// Create reverse proxy with custom response modifier for HTML rewriting
	proxy := &httputil.ReverseProxy{
		Director: func(req *http.Request) {
			targetURL, _ := url.Parse(target)
			req.URL = targetURL
			req.Host = targetURL.Host
			req.Header.Set("X-Forwarded-For", c.ClientIP())
			req.Header.Set("X-Forwarded-Proto", "http")
		},
		ModifyResponse: func(resp *http.Response) error {
			// Only modify HTML responses
			if isHTMLResponse(resp) {
				// Read the response body
				body, err := io.ReadAll(resp.Body)
				if err != nil {
					return err
				}
				resp.Body.Close()

				// Create HTML rewriter
				proxyPrefix := fmt.Sprintf("/proxy/%s:%d", host, port)
				rewriter := NewHTMLRewriter(proxyPrefix, host)

				// Rewrite HTML content
				modifiedBody := rewriter.RewriteHTML(body)

				// Create new response body
				resp.Body = io.NopCloser(bytes.NewReader(modifiedBody))
				resp.ContentLength = int64(len(modifiedBody))
				resp.Header.Set("Content-Length", strconv.Itoa(len(modifiedBody)))

				// Add custom header to indicate the response was modified
				resp.Header.Set("X-Proxy-Modified", "true")

				logger.Debug(fmt.Sprintf("Rewrote HTML response for %s (original: %d bytes, modified: %d bytes)",
					target, len(body), len(modifiedBody)))
			}
			return nil
		},
	}

	// Perform the proxy request
	proxy.ServeHTTP(c.Writer, c.Request)

	// Log the request
	status := c.Writer.Status()
	if status == 0 {
		status = 200 // Default status if not set
	}
	h.logRequest(c, host, portStr, targetPath, status, time.Since(start), "")
}

// HealthCheck returns the health status of the service
func (h *Handler) HealthCheck(c *gin.Context) {
	response := models.HealthResponse{
		Status:    "ok",
		Timestamp: time.Now(),
		Version:   "1.0.0",
	}
	c.JSON(http.StatusOK, response)
}

// GetStatus returns system status information
func (h *Handler) GetStatus(c *gin.Context) {
	totalRequests, _ := h.db.GetLogCount()
	uptime := time.Since(h.startTime)

	// Check ngrok status
	ngrokStatus := "disconnected"
	ngrokURL := ""

	h.ngrokMutex.Lock()
	if h.ngrokManager != nil {
		if url, isActive := h.ngrokManager.GetStatus(); isActive {
			ngrokStatus = "connected"
			ngrokURL = url
		}
	}
	h.ngrokMutex.Unlock()

	status := models.SystemStatus{
		Online:        true,
		LastCheck:     time.Now(),
		TotalRequests: totalRequests,
		Uptime:        uptime.String(),
		NgrokStatus:   ngrokStatus,
		NgrokURL:      ngrokURL,
	}

	c.JSON(http.StatusOK, status)
}

// GetLogs returns recent proxy logs
func (h *Handler) GetLogs(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "50")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 50
	}
	if limit > 1000 {
		limit = 1000
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	logs, err := h.db.GetLogs(limit, offset)
	if err != nil {
		logger.Error("Error fetching logs:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch logs"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"logs":   logs,
		"limit":  limit,
		"offset": offset,
	})
}

// ClearLogs clears all proxy logs
func (h *Handler) ClearLogs(c *gin.Context) {
	if err := h.db.ClearLogs(); err != nil {
		logger.Error("Error clearing logs:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to clear logs"})
		return
	}

	logger.Info("Logs cleared by user")
	c.JSON(http.StatusOK, gin.H{"message": "Logs cleared successfully"})
}

// GetSettings returns current application settings
func (h *Handler) GetSettings(c *gin.Context) {
	settings, err := h.db.GetSettings()
	if err != nil {
		logger.Error("Error fetching settings:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch settings"})
		return
	}

	// Don't expose the full token in the response for security
	tokenMasked := ""
	if len(settings.NgrokToken) > 8 {
		tokenMasked = settings.NgrokToken[:4] + "****" + settings.NgrokToken[len(settings.NgrokToken)-4:]
	} else if len(settings.NgrokToken) > 0 {
		tokenMasked = "****"
	}

	c.JSON(http.StatusOK, gin.H{
		"ngrok_token":  tokenMasked,
		"ngrok_domain": settings.NgrokDomain,
	})
}

// UpdateSettings updates application settings
func (h *Handler) UpdateSettings(c *gin.Context) {
	var request struct {
		NgrokToken  string `json:"ngrok_token"`
		NgrokDomain string `json:"ngrok_domain"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	settings := &models.Settings{
		NgrokToken:  request.NgrokToken,
		NgrokDomain: request.NgrokDomain,
	}

	if err := h.db.UpdateSettings(settings); err != nil {
		logger.Error("Error updating settings:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update settings"})
		return
	}

	logger.Info("Settings updated by user")
	c.JSON(http.StatusOK, gin.H{"message": "Settings updated successfully"})
}

// StartNgrokTunnel starts an ngrok tunnel
func (h *Handler) StartNgrokTunnel(c *gin.Context) {
	h.ngrokMutex.Lock()
	defer h.ngrokMutex.Unlock()

	settings, err := h.db.GetSettings()
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to get settings: %v", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get settings"})
		return
	}

	if settings.NgrokToken == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Ngrok token not configured. Please set your token in settings first.",
			"hint":  "Get your token from: https://dashboard.ngrok.com/get-started/your-authtoken",
		})
		return
	}

	// Stop any existing tunnel
	if h.ngrokManager != nil {
		h.ngrokManager.StopTunnel()
	}

	h.ngrokManager = ngrok.NewManager(settings.NgrokToken, settings.NgrokDomain, 8080)

	// Test connection first
	logger.Info("Testing ngrok authentication...")
	if err := h.ngrokManager.TestConnection(); err != nil {
		logger.Error(fmt.Sprintf("Ngrok authentication failed: %v", err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid ngrok token. Please check your token in settings.",
			"hint":    "Get your token from: https://dashboard.ngrok.com/get-started/your-authtoken",
			"details": err.Error(),
		})
		return
	}

	logger.Info("Starting ngrok tunnel...")
	url, err := h.ngrokManager.StartTunnel()
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to start ngrok tunnel: %v", err))

		// Provide specific error messages based on the error
		errorMsg := "Failed to start tunnel"
		hint := "Please check your internet connection and try again."

		if strings.Contains(err.Error(), "token") {
			errorMsg = "Authentication failed"
			hint = "Please verify your ngrok token in settings."
		} else if strings.Contains(err.Error(), "timeout") {
			errorMsg = "Tunnel startup timeout"
			hint = "Ngrok is taking longer than expected. Please try again."
		} else if strings.Contains(err.Error(), "connect: connection refused") {
			errorMsg = "Ngrok API connection failed"
			hint = "Ngrok process may not be starting properly. Check if port 4040 is available."
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   errorMsg,
			"hint":    hint,
			"details": err.Error(),
		})
		return
	}

	logger.Info(fmt.Sprintf("Ngrok tunnel started successfully: %s", url))
	c.JSON(http.StatusOK, gin.H{
		"url":     url,
		"message": "🎉 Tunnel started successfully! You can now access your relay from anywhere using this URL.",
		"usage":   fmt.Sprintf("Try: curl %s/api/health", url),
	})
}

// StopNgrokTunnel stops the ngrok tunnel
func (h *Handler) StopNgrokTunnel(c *gin.Context) {
	h.ngrokMutex.Lock()
	defer h.ngrokMutex.Unlock()

	if h.ngrokManager != nil {
		if err := h.ngrokManager.StopTunnel(); err != nil {
			logger.Error(fmt.Sprintf("Failed to stop ngrok tunnel: %v", err))
		} else {
			logger.Info("Ngrok tunnel stopped")
		}
		h.ngrokManager = nil
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Tunnel stopped successfully",
	})
}

// TestNgrokToken tests if the provided ngrok token is valid
func (h *Handler) TestNgrokToken(c *gin.Context) {
	var request struct {
		Token string `json:"token" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Token is required"})
		return
	}

	manager := ngrok.NewManager(request.Token, "", 8080)
	if err := manager.TestConnection(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"valid":   false,
			"error":   "Invalid token",
			"details": err.Error(),
			"hint":    "Get your token from: https://dashboard.ngrok.com/get-started/your-authtoken",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"valid":   true,
		"message": "✅ Token is valid!",
	})
}

// HTMLRewriter handles rewriting HTML content for proxied responses
type HTMLRewriter struct {
	proxyPrefix string
	targetHost  string
}

// NewHTMLRewriter creates a new HTML rewriter
func NewHTMLRewriter(proxyPrefix, targetHost string) *HTMLRewriter {
	return &HTMLRewriter{
		proxyPrefix: proxyPrefix,
		targetHost:  targetHost,
	}
}

// RewriteHTML rewrites relative URLs in HTML content
func (hr *HTMLRewriter) RewriteHTML(content []byte) []byte {
	html := string(content)

	// Define patterns for different HTML elements with relative URLs
	patterns := map[string]*regexp.Regexp{
		"href":    regexp.MustCompile(`href=['"](\s*[^'"]*?\s*)['"]`),
		"src":     regexp.MustCompile(`src=['"](\s*[^'"]*?\s*)['"]`),
		"action":  regexp.MustCompile(`action=['"](\s*[^'"]*?\s*)['"]`),
		"content": regexp.MustCompile(`content=['"](\s*[^'"]*?\s*)['"]`), // for meta redirects
		"url":     regexp.MustCompile(`url\(['"](\s*[^'"]*?\s*)['"]\)`),  // CSS url()
	}

	// Rewrite each pattern
	for attr, pattern := range patterns {
		html = pattern.ReplaceAllStringFunc(html, func(match string) string {
			return hr.rewriteAttribute(match, attr, pattern)
		})
	}

	// Inject base tag if not present (fallback mechanism)
	if !strings.Contains(strings.ToLower(html), "<base") {
		html = hr.injectBaseTag(html)
	}

	return []byte(html)
}

// rewriteAttribute rewrites a single attribute match
func (hr *HTMLRewriter) rewriteAttribute(match, attr string, pattern *regexp.Regexp) string {
	submatches := pattern.FindStringSubmatch(match)
	if len(submatches) < 2 {
		return match
	}

	originalURL := strings.TrimSpace(submatches[1])

	// Skip if already absolute, data URLs, javascript, mailto, etc.
	if hr.shouldSkipURL(originalURL) {
		return match
	}

	// Rewrite relative URLs
	newURL := hr.rewriteURL(originalURL)
	return strings.Replace(match, originalURL, newURL, 1)
}

// shouldSkipURL determines if a URL should be skipped from rewriting
func (hr *HTMLRewriter) shouldSkipURL(url string) bool {
	url = strings.ToLower(strings.TrimSpace(url))

	// Skip absolute URLs, special protocols, and anchors
	skipPrefixes := []string{
		"http://", "https://", "//", "ftp://",
		"data:", "javascript:", "mailto:", "tel:",
		"#", "about:", "chrome://", "file://",
	}

	for _, prefix := range skipPrefixes {
		if strings.HasPrefix(url, prefix) {
			return true
		}
	}

	// Skip empty URLs
	return url == "" || url == "/" || url == "#"
}

// rewriteURL converts a relative URL to use the proxy prefix
func (hr *HTMLRewriter) rewriteURL(originalURL string) string {
	// Handle root-relative URLs (starting with /)
	if strings.HasPrefix(originalURL, "/") {
		return hr.proxyPrefix + originalURL
	}

	// Handle relative URLs (no leading slash)
	// These are trickier - we assume they're relative to current path
	return hr.proxyPrefix + "/" + originalURL
}

// injectBaseTag injects a base tag into the HTML head
func (hr *HTMLRewriter) injectBaseTag(html string) string {
	baseTag := fmt.Sprintf(`<base href="%s/">`, hr.proxyPrefix)

	// Try to inject after <head> tag
	headPattern := regexp.MustCompile(`(?i)(<head[^>]*>)`)
	if headPattern.MatchString(html) {
		return headPattern.ReplaceAllString(html, `$1`+"\n    "+baseTag)
	}

	// Fallback: inject after <html> tag
	htmlPattern := regexp.MustCompile(`(?i)(<html[^>]*>)`)
	if htmlPattern.MatchString(html) {
		return htmlPattern.ReplaceAllString(html, `$1`+"\n  <head>\n    "+baseTag+"\n  </head>")
	}

	// Last resort: prepend to content
	return baseTag + "\n" + html
}

// isHTMLResponse checks if the response is HTML content
func isHTMLResponse(resp *http.Response) bool {
	contentType := resp.Header.Get("Content-Type")
	return strings.Contains(strings.ToLower(contentType), "text/html")
}

// Helper functions

func (h *Handler) logRequest(c *gin.Context, host, port, path string, statusCode int, duration time.Duration, errorMsg string) {
	entry := &models.LogEntry{
		Timestamp:  time.Now(),
		SourceIP:   c.ClientIP(),
		Method:     c.Request.Method,
		TargetHost: host,
		TargetPort: port,
		Path:       path,
		StatusCode: statusCode,
		Duration:   duration.Milliseconds(),
		Error:      errorMsg,
	}

	if err := h.db.InsertLogEntry(entry); err != nil {
		logger.Error("Failed to log request:", err)
	}
}

func isPrivateIP(ip string) bool {
	// Parse the IP
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return false
	}

	// Check for private IP ranges
	private := []string{
		"127.0.0.0/8",    // loopback
		"10.0.0.0/8",     // private
		"172.16.0.0/12",  // private
		"192.168.0.0/16", // private
		"169.254.0.0/16", // link-local
	}

	for _, cidr := range private {
		_, subnet, _ := net.ParseCIDR(cidr)
		if subnet.Contains(parsedIP) {
			return true
		}
	}

	return false
}

func isHopByHopHeader(name string) bool {
	hopByHop := []string{
		"Connection",
		"Keep-Alive",
		"Proxy-Authenticate",
		"Proxy-Authorization",
		"Te",
		"Trailers",
		"Transfer-Encoding",
		"Upgrade",
	}

	name = strings.ToLower(name)
	for _, header := range hopByHop {
		if strings.ToLower(header) == name {
			return true
		}
	}
	return false
}
