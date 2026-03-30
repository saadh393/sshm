package cmd

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/saadh393/sshm/internal/config"
	sshpkg "github.com/saadh393/sshm/internal/ssh"
	"github.com/saadh393/sshm/internal/tui"
	"github.com/spf13/cobra"
)

var editCmd = &cobra.Command{
	Use:   "edit [alias]",
	Short: "Edit an existing connection",
	Long: `Edit a saved connection interactively or via flags.

Without an alias, opens a TUI picker to choose which connection to edit.
Without any flags, opens an interactive edit form pre-filled with current values.
With flags, applies only the supplied values and saves immediately.`,
	Args: cobra.MaximumNArgs(1),
	RunE: runEdit,
}

var (
	editHost    string
	editUser    string
	editPort    int
	editKeyPath string
	editGroup   string
	editRename  string
)

func init() {
	rootCmd.AddCommand(editCmd)
	editCmd.Flags().StringVarP(&editHost, "host", "H", "", "New hostname or IP")
	editCmd.Flags().StringVarP(&editUser, "user", "u", "", "New username")
	editCmd.Flags().IntVarP(&editPort, "port", "p", 0, "New port")
	editCmd.Flags().StringVarP(&editKeyPath, "key", "k", "", "New key path")
	editCmd.Flags().StringVarP(&editGroup, "group", "g", "", "New group")
	editCmd.Flags().StringVar(&editRename, "rename", "", "New alias (rename)")
}

func runEdit(cmd *cobra.Command, args []string) error {
	conns, err := config.Load()
	if err != nil {
		return err
	}

	if len(conns) == 0 {
		fmt.Fprintln(os.Stdout, "No connections saved. Run 'sshm add' to add one.")
		return nil
	}

	// Resolve alias: from argument or TUI picker.
	var alias string
	if len(args) == 1 {
		alias = args[0]
	} else {
		result := tui.RunPicker(conns, "Edit — select a connection  [enter] select  [/] filter  [q] quit")
		if result.Quit || result.Conn == nil {
			return nil
		}
		alias = result.Conn.Alias
	}

	conn, ok := config.FindExact(conns, alias)
	if !ok {
		return fmt.Errorf("connection %q not found", alias)
	}

	flags := cmd.Flags()
	anyFlagSet := flags.Changed("host") || flags.Changed("user") || flags.Changed("port") ||
		flags.Changed("key") || flags.Changed("group") || flags.Changed("rename")

	if anyFlagSet {
		// Flag-based edit (non-interactive).
		conn, conns, err = applyEditFlags(cmd, conn, conns, alias)
		if err != nil {
			return err
		}
	} else {
		// Interactive edit form.
		result := tui.RunEditForm(conn)
		if !result.Saved || result.Conn == nil {
			fmt.Fprintln(os.Stdout, "Edit cancelled.")
			return nil
		}
		updated := *result.Conn

		// Handle rename: remove old alias, add new.
		if updated.Alias != conn.Alias {
			if _, exists := config.FindExact(conns, updated.Alias); exists {
				return fmt.Errorf("alias %q already exists", updated.Alias)
			}
			conns, err = config.Remove(conns, conn.Alias)
			if err != nil {
				return err
			}
			conns = append(conns, updated)
		} else {
			conns, err = config.Update(conns, updated)
			if err != nil {
				return err
			}
		}
		conn = updated
	}

	if err := config.Save(conns); err != nil {
		return err
	}

	green := color.New(color.FgGreen, color.Bold)
	fmt.Fprintf(os.Stdout, "%s Updated connection %q\n", green.Sprint("✓"), conn.Alias)
	return nil
}

func applyEditFlags(cmd *cobra.Command, conn config.Connection, conns []config.Connection, origAlias string) (config.Connection, []config.Connection, error) {
	flags := cmd.Flags()
	var err error

	if flags.Changed("host") {
		conn.Host = editHost
	}
	if flags.Changed("user") {
		conn.User = editUser
	}
	if flags.Changed("port") {
		conn.Port = editPort
	}
	if flags.Changed("key") {
		if err = sshpkg.ValidateKeyPath(editKeyPath); err != nil {
			return conn, conns, err
		}
		conn.KeyPath = editKeyPath
	}
	if flags.Changed("group") {
		conn.Group = editGroup
	}

	if flags.Changed("rename") {
		if _, exists := config.FindExact(conns, editRename); exists {
			return conn, conns, fmt.Errorf("alias %q already exists", editRename)
		}
		conns, err = config.Remove(conns, origAlias)
		if err != nil {
			return conn, conns, err
		}
		conn.Alias = editRename
		conns = append(conns, conn)
	} else {
		conns, err = config.Update(conns, conn)
		if err != nil {
			return conn, conns, err
		}
	}

	return conn, conns, nil
}
