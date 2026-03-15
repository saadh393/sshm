package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

// Version is set at build time via -ldflags.
var Version = "dev"

var rootCmd = &cobra.Command{
	Use:   "sshm",
	Short: "sshm – a lightweight SSH connection manager",
	Long: `sshm manages SSH connections stored in ~/.config/sshm/connections.json.

Use 'sshm add', 'sshm list', 'sshm connect' and friends to manage your
connections.`,
}

// Execute is the entry-point called from main.go.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Printf("sshm %s\n", Version)
	},
}
