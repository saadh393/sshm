package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/saadh393/sshm/internal/config"
	"github.com/spf13/cobra"
)

var removeCmd = &cobra.Command{
	Use:     "remove <alias>",
	Aliases: []string{"rm"},
	Short:   "Remove a saved connection",
	Args:    cobra.ExactArgs(1),
	RunE:    runRemove,
}

var removeYes bool

func init() {
	rootCmd.AddCommand(removeCmd)
	removeCmd.Flags().BoolVarP(&removeYes, "yes", "y", false, "Skip confirmation prompt")
}

func runRemove(cmd *cobra.Command, args []string) error {
	alias := args[0]

	conns, err := config.Load()
	if err != nil {
		return err
	}

	conn, ok := config.FindExact(conns, alias)
	if !ok {
		return fmt.Errorf("connection %q not found", alias)
	}

	if !removeYes {
		yellow := color.New(color.FgYellow)
		fmt.Fprintf(os.Stderr, "%s Remove connection %q (%s@%s)? [y/N] ",
			yellow.Sprint("?"), conn.Alias, conn.User, conn.Host)

		reader := bufio.NewReader(os.Stdin)
		line, _ := reader.ReadString('\n')
		line = strings.TrimSpace(strings.ToLower(line))
		if line != "y" && line != "yes" {
			fmt.Fprintln(os.Stdout, "Aborted.")
			return nil
		}
	}

	conns, err = config.Remove(conns, alias)
	if err != nil {
		return err
	}

	if err := config.Save(conns); err != nil {
		return err
	}

	red := color.New(color.FgRed, color.Bold)
	fmt.Fprintf(os.Stdout, "%s Removed connection %q\n", red.Sprint("✗"), alias)
	return nil
}
