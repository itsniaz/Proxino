package ngrok

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

type TunnelInfo struct {
	PublicURL string `json:"public_url"`
	Proto     string `json:"proto"`
	Config    struct {
		Addr string `json:"addr"`
	} `json:"config"`
}

type TunnelsResponse struct {
	Tunnels []TunnelInfo `json:"tunnels"`
}

type NgrokManager struct {
	cmd     *exec.Cmd
	token   string
	domain  string
	port    int
	apiPort int
}

func NewManager(token, domain string, port int) *NgrokManager {
	return &NgrokManager{
		token:   token,
		domain:  domain,
		port:    port,
		apiPort: 4040, // Default ngrok API port
	}
}

// StartTunnel starts an ngrok tunnel programmatically
func (n *NgrokManager) StartTunnel() (string, error) {
	// Validate token first
	if n.token == "" {
		return "", fmt.Errorf("ngrok token is required")
	}

	// Kill any existing ngrok processes to avoid conflicts
	exec.Command("pkill", "-f", "ngrok").Run()
	time.Sleep(1 * time.Second)

	// Build ngrok command
	args := []string{"http", strconv.Itoa(n.port), "--authtoken", n.token}

	if n.domain != "" {
		args = append(args, "--hostname", n.domain)
	}

	// Add log configuration for better debugging
	args = append(args, "--log", "stdout", "--log-level", "info")

	n.cmd = exec.Command("ngrok", args...)

	// Start the process
	if err := n.cmd.Start(); err != nil {
		return "", fmt.Errorf("failed to start ngrok process: %v", err)
	}

	// Wait longer for ngrok to fully initialize
	fmt.Println("Waiting for ngrok to initialize...")

	// Try to connect to ngrok API with retries
	maxRetries := 12 // 12 seconds total
	for i := 0; i < maxRetries; i++ {
		time.Sleep(1 * time.Second)

		// Check if the process is still running
		if n.cmd.Process == nil {
			return "", fmt.Errorf("ngrok process died unexpectedly")
		}

		// Try to get the URL
		url, err := n.getPublicURL()
		if err == nil {
			fmt.Printf("Ngrok tunnel established: %s\n", url)
			return url, nil
		}

		// On last retry, return the error
		if i == maxRetries-1 {
			return "", fmt.Errorf("failed to establish ngrok tunnel after %d attempts: %v", maxRetries, err)
		}

		fmt.Printf("Attempt %d/%d: Waiting for ngrok API... (%v)\n", i+1, maxRetries, err)
	}

	return "", fmt.Errorf("timeout waiting for ngrok to start")
}

// StopTunnel stops the ngrok tunnel
func (n *NgrokManager) StopTunnel() error {
	if n.cmd != nil && n.cmd.Process != nil {
		err := n.cmd.Process.Kill()
		if err != nil {
			// Try pkill as backup
			exec.Command("pkill", "-f", "ngrok").Run()
		}
		n.cmd = nil
		return err
	}
	// Also try pkill to ensure cleanup
	exec.Command("pkill", "-f", "ngrok").Run()
	return nil
}

// getPublicURL retrieves the public URL from ngrok's local API
func (n *NgrokManager) getPublicURL() (string, error) {
	url := fmt.Sprintf("http://localhost:%d/api/tunnels", n.apiPort)

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 3 * time.Second,
	}

	resp, err := client.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to connect to ngrok API at %s: %v", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("ngrok API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read ngrok API response: %v", err)
	}

	var tunnels TunnelsResponse
	if err := json.Unmarshal(body, &tunnels); err != nil {
		return "", fmt.Errorf("failed to parse ngrok API response: %v", err)
	}

	// Find the HTTPS tunnel first, then HTTP as fallback
	for _, tunnel := range tunnels.Tunnels {
		if tunnel.Proto == "https" &&
			strings.Contains(tunnel.Config.Addr, strconv.Itoa(n.port)) {
			return tunnel.PublicURL, nil
		}
	}

	// Fallback to HTTP tunnel
	for _, tunnel := range tunnels.Tunnels {
		if tunnel.Proto == "http" &&
			strings.Contains(tunnel.Config.Addr, strconv.Itoa(n.port)) {
			return tunnel.PublicURL, nil
		}
	}

	return "", fmt.Errorf("no tunnel found for port %d in %d tunnels", n.port, len(tunnels.Tunnels))
}

// GetStatus checks if ngrok is running and returns tunnel info
func (n *NgrokManager) GetStatus() (string, bool) {
	url, err := n.getPublicURL()
	if err != nil {
		return "", false
	}
	return url, true
}

// TestConnection tests if ngrok can connect with the given token
func (n *NgrokManager) TestConnection() error {
	if n.token == "" {
		return fmt.Errorf("no token provided")
	}

	// Try a quick test command
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "ngrok", "authtoken", n.token)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("invalid ngrok token: %v", err)
	}

	return nil
}
