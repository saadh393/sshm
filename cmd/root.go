package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// Version is set at build time via -ldflags.
var Version = "dev"

var rootCmd = &cobra.Command{
	Use:           "sshm",
	Short:         "sshm – a lightweight SSH connection manager",
	SilenceUsage:  true,
	SilenceErrors: true,
	Long: `sshm manages SSH connections stored in ~/.config/sshm/connections.json.

Run 'sshm -h' to see available commands.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runConnectionListFlow()
	},
}

// Execute is the entry-point called from main.go.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
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
