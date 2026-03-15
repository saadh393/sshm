package cmd

import (
	"fmt"
	"os"

	"github.com/saadh393/sshm/internal/config"
	"github.com/saadh393/sshm/internal/tui"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List connections in an interactive TUI",
	Args:    cobra.NoArgs,
	RunE:    runList,
}

func init() {
	rootCmd.AddCommand(listCmd)
}

func runList(_ *cobra.Command, _ []string) error {
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
}
