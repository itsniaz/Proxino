package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	version = "1.0.0"
	port    string
	daemon  bool
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "lan-relay",
	Short: "LAN Relay - Forward local network traffic through secure tunnels",
	Long: `LAN Relay is a powerful tool that allows you to expose and proxy 
local network services through secure tunnels. It provides a web dashboard
for easy management and supports ngrok integration for external access.`,
	Version: version,
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func init() {
	// Global flags
	rootCmd.PersistentFlags().StringVarP(&port, "port", "p", "8080", "Port to run the server on")
	rootCmd.PersistentFlags().BoolVarP(&daemon, "daemon", "d", false, "Run in daemon mode")
}
