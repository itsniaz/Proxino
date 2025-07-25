package cmd

import (
	"fmt"
	"net/http"
	"time"

	"github.com/spf13/cobra"
)

// statusCmd represents the status command
var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check if LAN Relay is running",
	Long:  `Check the status of the LAN Relay service and show if it's accessible.`,
	Run: func(cmd *cobra.Command, args []string) {
		checkStatus()
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}

func checkStatus() {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	url := fmt.Sprintf("http://localhost:%s/api/health", port)
	resp, err := client.Get(url)
	if err != nil {
		fmt.Printf("❌ LAN Relay is not running on port %s\n", port)
		fmt.Printf("   Error: %v\n", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		fmt.Printf("✅ LAN Relay is running on port %s\n", port)
		fmt.Printf("📊 Dashboard: http://localhost:%s\n", port)
		fmt.Printf("🔧 API: http://localhost:%s/api\n", port)
	} else {
		fmt.Printf("⚠️  LAN Relay responded with status code %d\n", resp.StatusCode)
	}
}
