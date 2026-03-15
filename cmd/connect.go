package cmd

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/saadh393/sshm/internal/config"
	sshpkg "github.com/saadh393/sshm/internal/ssh"
	"github.com/spf13/cobra"
)

var connectCmd = &cobra.Command{
	Use:     "connect <alias>",
	Aliases: []string{"c"},
	Short:   "Connect to a saved SSH connection",
	Args:    cobra.ExactArgs(1),
	RunE:    runConnect,
}

var connectDry bool

func init() {
	rootCmd.AddCommand(connectCmd)
	connectCmd.Flags().BoolVar(&connectDry, "dry", false, "Print the SSH command without executing")
}

func runConnect(cmd *cobra.Command, args []string) error {
	query := args[0]

	conns, err := config.Load()
	if err != nil {
		return err
	}

	// 1. Exact match
	if conn, ok := config.FindExact(conns, query); ok {
		return doConnect(conn, connectDry)
	}

	// 2. Substring match
	matches := config.FindSubstring(conns, query)

	switch len(matches) {
	case 0:
		return fmt.Errorf("no connection found matching %q", query)

	case 1:
		yellow := color.New(color.FgYellow)
		fmt.Fprintf(os.Stderr, "%s No exact match; connecting to %q (%s@%s)\n",
			yellow.Sprint("⚠"), matches[0].Alias, matches[0].User, matches[0].Host)
		return doConnect(matches[0], connectDry)

	default:
		fmt.Fprintf(os.Stderr, "Multiple connections match %q — be more specific:\n\n", query)
		for _, c := range matches {
			fmt.Fprintf(os.Stderr, "  %-20s %s@%s\n", c.Alias, c.User, c.Host)
		}
		return fmt.Errorf("ambiguous alias")
	}
}

func doConnect(conn config.Connection, dry bool) error {
	cmdStr := sshpkg.CommandString(conn)
	if dry {
		fmt.Println(cmdStr)
		return nil
	}

	blue := color.New(color.FgCyan)
	fmt.Fprintf(os.Stderr, "%s Connecting: %s\n", blue.Sprint("→"), cmdStr)

	return sshpkg.Connect(conn)
}
