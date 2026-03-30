package cmd

import (
	"fmt"
	"os"

	"github.com/saadh393/sshm/internal/config"
	"github.com/saadh393/sshm/internal/tui"
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
		conns, err := config.Load()
		if err != nil {
			return err
		}
		if len(conns) == 0 {
			fmt.Fprintln(os.Stdout, "No connections saved. Run 'sshm add' to add one.")
			return nil
		}
		result := tui.Run(conns)
		if result.Quit || result.Conn == nil {
			return nil
		}
		return doConnect(*result.Conn, false)
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
