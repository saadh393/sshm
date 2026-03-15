package cmd

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/saadh393/sshm/internal/config"
	sshpkg "github.com/saadh393/sshm/internal/ssh"
	"github.com/spf13/cobra"
)

var editCmd = &cobra.Command{
	Use:   "edit <alias>",
	Short: "Edit an existing connection (only supplied flags are updated)",
	Args:  cobra.ExactArgs(1),
	RunE:  runEdit,
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
	alias := args[0]

	conns, err := config.Load()
	if err != nil {
		return err
	}

	conn, ok := config.FindExact(conns, alias)
	if !ok {
		return fmt.Errorf("connection %q not found", alias)
	}

	// Apply only the flags the user explicitly passed.
	flags := cmd.Flags()

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
		if err := sshpkg.ValidateKeyPath(editKeyPath); err != nil {
			return err
		}
		conn.KeyPath = editKeyPath
	}
	if flags.Changed("group") {
		conn.Group = editGroup
	}

	if flags.Changed("rename") {
		// Make sure the new alias is not already taken.
		if _, exists := config.FindExact(conns, editRename); exists {
			return fmt.Errorf("alias %q already exists", editRename)
		}
		// Remove old entry, set new alias, add fresh.
		conns, err = config.Remove(conns, alias)
		if err != nil {
			return err
		}
		conn.Alias = editRename
		conns = append(conns, conn)
	} else {
		conns, err = config.Update(conns, conn)
		if err != nil {
			return err
		}
	}

	if err := config.Save(conns); err != nil {
		return err
	}

	green := color.New(color.FgGreen, color.Bold)
	fmt.Fprintf(os.Stdout, "%s Updated connection %q\n", green.Sprint("✓"), conn.Alias)
	return nil
}
