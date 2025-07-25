package models

import "time"

// LogEntry represents a proxy request log entry
type LogEntry struct {
	ID         int       `json:"id" db:"id"`
	Timestamp  time.Time `json:"timestamp" db:"timestamp"`
	SourceIP   string    `json:"source_ip" db:"source_ip"`
	Method     string    `json:"method" db:"method"`
	TargetHost string    `json:"target_host" db:"target_host"`
	TargetPort string    `json:"target_port" db:"target_port"`
	Path       string    `json:"path" db:"path"`
	StatusCode int       `json:"status_code" db:"status_code"`
	Duration   int64     `json:"duration_ms" db:"duration_ms"`
	Error      string    `json:"error,omitempty" db:"error"`
}

// SystemStatus represents the current status of the relay system
type SystemStatus struct {
	Online        bool      `json:"online"`
	LastCheck     time.Time `json:"last_check"`
	TotalRequests int       `json:"total_requests"`
	Uptime        string    `json:"uptime"`
	NgrokStatus   string    `json:"ngrok_status"`
	NgrokURL      string    `json:"ngrok_url,omitempty"`
}

// HealthResponse represents health check response
type HealthResponse struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
	Version   string    `json:"version"`
}

// Settings represents application settings
type Settings struct {
	ID          int       `json:"id" db:"id"`
	NgrokToken  string    `json:"ngrok_token" db:"ngrok_token"`
	NgrokDomain string    `json:"ngrok_domain" db:"ngrok_domain"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// NgrokTunnelResponse represents ngrok tunnel start response
type NgrokTunnelResponse struct {
	URL     string `json:"url"`
	Message string `json:"message"`
}
