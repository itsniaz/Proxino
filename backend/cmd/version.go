package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version information",
	Long:  `Display the current version of LAN Relay along with build information.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("LAN Relay v%s\n", version)
		fmt.Println("A secure local network proxy with web dashboard")
		fmt.Println("GitHub: https://github.com/yourusername/local_router")
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
