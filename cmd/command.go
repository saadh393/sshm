package cmd

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/fatih/color"
	"github.com/saadh393/sshm/internal/config"
	sshpkg "github.com/saadh393/sshm/internal/ssh"
	"github.com/spf13/cobra"
)

var commandCmd = &cobra.Command{
	Use:     "command",
	Aliases: []string{"cmd"},
	Short:   "Manage saved remote commands for a connection",
}

var commandAddCmd = &cobra.Command{
	Use:   "add <alias> <name> <remote-command>",
	Short: "Add a saved remote command to a connection",
	Args:  cobra.ExactArgs(3),
	RunE:  runCommandAdd,
}

var commandUpdateCmd = &cobra.Command{
	Use:   "update <alias> <name> <remote-command>",
	Short: "Update a saved remote command on a connection",
	Args:  cobra.ExactArgs(3),
	RunE:  runCommandUpdate,
}

var commandDeleteCmd = &cobra.Command{
	Use:     "delete <alias> <name>",
	Aliases: []string{"remove", "rm"},
	Short:   "Delete a saved remote command from a connection",
	Args:    cobra.ExactArgs(2),
	RunE:    runCommandDelete,
}

var commandListCmd = &cobra.Command{
	Use:   "list <alias>",
	Short: "List saved remote commands for a connection",
	Args:  cobra.ExactArgs(1),
	RunE:  runCommandList,
}

var commandRunCmd = &cobra.Command{
	Use:   "run <alias> <name>",
	Short: "Run a saved remote command on a connection",
	Args:  cobra.ExactArgs(2),
	RunE:  runCommandRun,
}

var commandRunDry bool

func init() {
	rootCmd.AddCommand(commandCmd)

	commandCmd.AddCommand(commandAddCmd)
	commandCmd.AddCommand(commandUpdateCmd)
	commandCmd.AddCommand(commandDeleteCmd)
	commandCmd.AddCommand(commandListCmd)
	commandCmd.AddCommand(commandRunCmd)

	commandRunCmd.Flags().BoolVar(&commandRunDry, "dry", false, "Print the ssh command without executing")
}

func runCommandAdd(_ *cobra.Command, args []string) error {
	alias, name, remoteCommand := args[0], args[1], args[2]

	conns, idx, conn, err := loadConnectionForMutation(alias)
	if err != nil {
		return err
	}

	if conn.Commands == nil {
		conn.Commands = make(map[string]string)
	}
	if _, exists := conn.Commands[name]; exists {
		return fmt.Errorf("command %q already exists for connection %q", name, conn.Alias)
	}
	if strings.TrimSpace(remoteCommand) == "" {
		return fmt.Errorf("remote command cannot be empty")
	}

	conn.Commands[name] = remoteCommand
	conns[idx] = conn
	if err := config.Save(conns); err != nil {
		return err
	}

	green := color.New(color.FgGreen, color.Bold)
	fmt.Fprintf(os.Stdout, "%s Added command %q to %q\n", green.Sprint("✓"), name, conn.Alias)
	return nil
}

func runCommandUpdate(_ *cobra.Command, args []string) error {
	alias, name, remoteCommand := args[0], args[1], args[2]

	conns, idx, conn, err := loadConnectionForMutation(alias)
	if err != nil {
		return err
	}
	if conn.Commands == nil {
		return fmt.Errorf("no saved commands for connection %q", conn.Alias)
	}
	if _, exists := conn.Commands[name]; !exists {
		return fmt.Errorf("command %q not found for connection %q", name, conn.Alias)
	}
	if strings.TrimSpace(remoteCommand) == "" {
		return fmt.Errorf("remote command cannot be empty")
	}

	conn.Commands[name] = remoteCommand
	conns[idx] = conn
	if err := config.Save(conns); err != nil {
		return err
	}

	green := color.New(color.FgGreen, color.Bold)
	fmt.Fprintf(os.Stdout, "%s Updated command %q on %q\n", green.Sprint("✓"), name, conn.Alias)
	return nil
}

func runCommandDelete(_ *cobra.Command, args []string) error {
	alias, name := args[0], args[1]

	conns, idx, conn, err := loadConnectionForMutation(alias)
	if err != nil {
		return err
	}
	if conn.Commands == nil {
		return fmt.Errorf("no saved commands for connection %q", conn.Alias)
	}
	if _, exists := conn.Commands[name]; !exists {
		return fmt.Errorf("command %q not found for connection %q", name, conn.Alias)
	}

	delete(conn.Commands, name)
	if len(conn.Commands) == 0 {
		conn.Commands = nil
	}
	conns[idx] = conn
	if err := config.Save(conns); err != nil {
		return err
	}

	red := color.New(color.FgRed, color.Bold)
	fmt.Fprintf(os.Stdout, "%s Deleted command %q from %q\n", red.Sprint("✗"), name, conn.Alias)
	return nil
}

func runCommandList(_ *cobra.Command, args []string) error {
	alias := args[0]
	conns, err := config.Load()
	if err != nil {
		return err
	}

	conn, ok := config.FindExact(conns, alias)
	if !ok {
		return fmt.Errorf("connection %q not found", alias)
	}

	if len(conn.Commands) == 0 {
		fmt.Fprintf(os.Stdout, "No saved commands for %q.\n", conn.Alias)
		return nil
	}

	names := make([]string, 0, len(conn.Commands))
	for n := range conn.Commands {
		names = append(names, n)
	}
	sort.Strings(names)

	bold := color.New(color.Bold)
	cyan := color.New(color.FgCyan)
	fmt.Fprintf(os.Stdout, "%s\n", bold.Sprintf("Saved commands for %s:", conn.Alias))
	for _, n := range names {
		fmt.Fprintf(os.Stdout, "  %-20s %s\n", bold.Sprint(n), cyan.Sprint(conn.Commands[n]))
	}
	return nil
}

func runCommandRun(_ *cobra.Command, args []string) error {
	alias, name := args[0], args[1]

	conns, err := config.Load()
	if err != nil {
		return err
	}
	conn, ok := config.FindExact(conns, alias)
	if !ok {
		return fmt.Errorf("connection %q not found", alias)
	}
	if len(conn.Commands) == 0 {
		return fmt.Errorf("no saved commands for connection %q", conn.Alias)
	}
	remoteCommand, exists := conn.Commands[name]
	if !exists {
		return fmt.Errorf("command %q not found for connection %q", name, conn.Alias)
	}

	cmdStr := sshpkg.RemoteCommandString(conn, remoteCommand)
	if commandRunDry {
		fmt.Fprintln(os.Stdout, cmdStr)
		return nil
	}

	blue := color.New(color.FgCyan)
	fmt.Fprintf(os.Stderr, "%s Running: %s\n", blue.Sprint("→"), cmdStr)

	return sshpkg.ConnectRemoteCommand(conn, remoteCommand)
}

func loadConnectionForMutation(alias string) ([]config.Connection, int, config.Connection, error) {
	conns, err := config.Load()
	if err != nil {
		return nil, -1, config.Connection{}, err
	}

	conn, ok := config.FindExact(conns, alias)
	if !ok {
		return nil, -1, config.Connection{}, fmt.Errorf("connection %q not found", alias)
	}

	for i, c := range conns {
		if strings.EqualFold(c.Alias, conn.Alias) {
			return conns, i, conn, nil
		}
	}

	return nil, -1, config.Connection{}, fmt.Errorf("connection %q not found", conn.Alias)
}
