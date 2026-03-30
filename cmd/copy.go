package cmd

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/saadh393/sshm/internal/config"
	"github.com/saadh393/sshm/internal/scp"
	"github.com/saadh393/sshm/internal/tui"
	"github.com/spf13/cobra"
)

var copyCmd = &cobra.Command{
	Use:   "copy [alias] [local] [remote]",
	Short: "Copy files to/from a saved connection via SCP",
	Long: `Transfer files using SCP with a saved connection.

All arguments are optional — omitting them opens interactive prompts:

  sshm copy                          # pick connection, then fill paths interactively
  sshm copy prod-api                 # pick paths interactively for prod-api
  sshm copy prod-api ./app /tmp/app  # direct upload (local → remote)
  sshm copy prod-api ./app /tmp/app --direction download  # explicit download

Direction defaults to upload (local → remote). Use --direction download to reverse.`,
	Aliases: []string{"cp", "scp"},
	Args:    cobra.RangeArgs(0, 3),
	RunE:    runCopy,
}

var (
	copyDirection string
	copyDry       bool
)

func init() {
	rootCmd.AddCommand(copyCmd)
	copyCmd.Flags().StringVarP(&copyDirection, "direction", "d", "", "Transfer direction: upload or download")
	copyCmd.Flags().BoolVar(&copyDry, "dry", false, "Print the scp command without executing")
}

func runCopy(cmd *cobra.Command, args []string) error {
	conns, err := config.Load()
	if err != nil {
		return err
	}

	if len(conns) == 0 {
		fmt.Fprintln(os.Stdout, "No connections saved. Run 'sshm add' to add one.")
		return nil
	}

	// --- Resolve connection ---
	var conn config.Connection

	if len(args) >= 1 {
		alias := args[0]
		c, ok := config.FindExact(conns, alias)
		if !ok {
			// Try substring
			matches := config.FindSubstring(conns, alias)
			switch len(matches) {
			case 0:
				return fmt.Errorf("no connection found matching %q", alias)
			case 1:
				yellow := color.New(color.FgYellow)
				fmt.Fprintf(os.Stderr, "%s No exact match; using %q (%s@%s)\n",
					yellow.Sprint("⚠"), matches[0].Alias, matches[0].User, matches[0].Host)
				c = matches[0]
			default:
				fmt.Fprintf(os.Stderr, "Multiple connections match %q — be more specific:\n\n", alias)
				for _, m := range matches {
					fmt.Fprintf(os.Stderr, "  %-20s %s@%s\n", m.Alias, m.User, m.Host)
				}
				return fmt.Errorf("ambiguous alias")
			}
		}
		conn = c
	} else {
		// No alias → TUI picker
		result := tui.RunPicker(conns, "Copy — select a connection  [enter] select  [/] filter  [q] quit")
		if result.Quit || result.Conn == nil {
			return nil
		}
		conn = *result.Conn
	}

	// --- Resolve paths and direction ---
	if len(args) == 3 {
		// Fully non-interactive: alias + local + remote provided
		localPath := args[1]
		remotePath := args[2]
		dir, err := resolveDirection(cmd)
		if err != nil {
			return err
		}
		return doCopy(conn, localPath, remotePath, dir, copyDry)
	}

	// Interactive copy form (direction toggle + path inputs)
	result := tui.RunCopyForm(conn)
	if !result.Confirmed {
		fmt.Fprintln(os.Stdout, "Copy cancelled.")
		return nil
	}

	return doCopy(conn, result.LocalPath, result.RemotePath, result.Direction, copyDry)
}

// resolveDirection maps the --direction flag to scp.Direction.
// Defaults to Upload when the flag is not set.
func resolveDirection(cmd *cobra.Command) (scp.Direction, error) {
	if !cmd.Flags().Changed("direction") {
		return scp.Upload, nil
	}
	switch copyDirection {
	case "upload", "up", "u":
		return scp.Upload, nil
	case "download", "down", "d":
		return scp.Download, nil
	default:
		return scp.Upload, fmt.Errorf("invalid direction %q — use upload or download", copyDirection)
	}
}

func doCopy(conn config.Connection, localPath, remotePath string, dir scp.Direction, dry bool) error {
	cmdStr := scp.CommandString(conn, localPath, remotePath, dir)

	if dry {
		fmt.Println(cmdStr)
		return nil
	}

	blue := color.New(color.FgCyan)
	arrow := "↑"
	if dir == scp.Download {
		arrow = "↓"
	}
	fmt.Fprintf(os.Stderr, "%s Copying (%s): %s\n", blue.Sprint(arrow), dir, cmdStr)

	if err := scp.Copy(conn, localPath, remotePath, dir); err != nil {
		return fmt.Errorf("scp failed: %w", err)
	}

	green := color.New(color.FgGreen, color.Bold)
	fmt.Fprintf(os.Stderr, "%s Transfer complete.\n", green.Sprint("✓"))
	return nil
}
