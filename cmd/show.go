package cmd

import (
	"fmt"
	"sort"

	"github.com/fatih/color"
	"github.com/saadh393/sshm/internal/config"
	sshpkg "github.com/saadh393/sshm/internal/ssh"
	"github.com/spf13/cobra"
)

var showCmd = &cobra.Command{
	Use:   "show <alias>",
	Short: "Show full details of a connection",
	Args:  cobra.ExactArgs(1),
	RunE:  runShow,
}

func init() {
	rootCmd.AddCommand(showCmd)
}

func runShow(cmd *cobra.Command, args []string) error {
	alias := args[0]

	conns, err := config.Load()
	if err != nil {
		return err
	}

	conn, ok := config.FindExact(conns, alias)
	if !ok {
		return fmt.Errorf("connection %q not found", alias)
	}

	bold := color.New(color.Bold)
	cyan := color.New(color.FgCyan)

	label := func(s string) string { return bold.Sprintf("  %-12s", s) }
	value := func(s string) string { return cyan.Sprint(s) }

	fmt.Printf("\n%s\n", bold.Sprint("Connection Details"))
	fmt.Printf("%s %s\n", label("Alias:"), value(conn.Alias))
	fmt.Printf("%s %s\n", label("Host:"), value(conn.Host))
	fmt.Printf("%s %s\n", label("User:"), value(conn.User))
	fmt.Printf("%s %s\n", label("Port:"), value(fmt.Sprintf("%d", conn.Port)))

	if conn.KeyPath != "" {
		fmt.Printf("%s %s\n", label("Key:"), value(conn.KeyPath))
	} else {
		fmt.Printf("%s %s\n", label("Key:"), "(none)")
	}

	if conn.Group != "" {
		fmt.Printf("%s %s\n", label("Group:"), value(conn.Group))
	}

	if len(conn.Commands) > 0 {
		names := make([]string, 0, len(conn.Commands))
		for name := range conn.Commands {
			names = append(names, name)
		}
		sort.Strings(names)

		fmt.Printf("%s\n", label("Commands:"))
		for _, name := range names {
			fmt.Printf("  %-12s %s\n", name+":", value(conn.Commands[name]))
		}
	}

	fmt.Printf("\n%s %s\n\n", bold.Sprint("SSH Command:"), sshpkg.CommandString(conn))
	return nil
}
