package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/fatih/color"
	"github.com/saadh393/sshm/internal/config"
	sshpkg "github.com/saadh393/sshm/internal/ssh"
	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a new SSH connection (interactive)",
	Args:  cobra.NoArgs,
	RunE:  runAdd,
}

func init() {
	rootCmd.AddCommand(addCmd)
}

func prompt(reader *bufio.Reader, question string) string {
	fmt.Print(question)
	line, _ := reader.ReadString('\n')
	return strings.TrimSpace(line)
}

func promptDefault(reader *bufio.Reader, question, defaultVal string) string {
	ans := prompt(reader, question)
	if ans == "" {
		return defaultVal
	}
	return ans
}

func runAdd(_ *cobra.Command, _ []string) error {
	reader := bufio.NewReader(os.Stdin)
	bold := color.New(color.Bold)
	cyan := color.New(color.FgCyan, color.Bold)
	green := color.New(color.FgGreen, color.Bold)
	yellow := color.New(color.FgYellow)
	dim := color.New(color.Faint)

	cyan.Println("\n  Add New SSH Connection")
	dim.Println("  ─────────────────────────────────────")
	fmt.Println()

	// 1. Alias
	alias := ""
	for alias == "" {
		alias = prompt(reader, bold.Sprint("  Alias (name for this connection): "))
		if alias == "" {
			yellow.Println("  Alias cannot be empty.")
		}
	}

	// 2. SSH command
	fmt.Println()
	dim.Println("  Paste your SSH command (e.g. ssh -i ~/.ssh/id_ed25519 ubuntu@1.2.3.4)")
	rawCmd := ""
	for {
		rawCmd = prompt(reader, bold.Sprint("  SSH command: "))
		if rawCmd != "" {
			break
		}
		yellow.Println("  Please enter an SSH command.")
	}

	parsed, err := sshpkg.ParseCommand(rawCmd)
	if err != nil {
		return fmt.Errorf("could not parse SSH command: %w", err)
	}

	fmt.Println()
	dim.Println("  Parsed:")
	fmt.Printf("    Host : %s\n", parsed.Host)
	if parsed.User != "" {
		fmt.Printf("    User : %s\n", parsed.User)
	}
	if parsed.KeyPath != "" {
		fmt.Printf("    Key  : %s\n", parsed.KeyPath)
	}
	if parsed.Port != 0 {
		fmt.Printf("    Port : %d\n", parsed.Port)
	}
	fmt.Println()

	// 3. User — if not found in the command
	if parsed.User == "" {
		parsed.User = prompt(reader, bold.Sprint("  Username (e.g. ubuntu): "))
		if parsed.User == "" {
			return fmt.Errorf("username is required")
		}
	}

	// 4. Port — if not found in the command
	finalPort := 22
	if parsed.Port != 0 {
		finalPort = parsed.Port
	} else {
		portStr := promptDefault(reader, bold.Sprint("  Port [22]: "), "22")
		p, err := strconv.Atoi(portStr)
		if err != nil || p < 1 || p > 65535 {
			yellow.Println("  Invalid port, using 22.")
			p = 22
		}
		finalPort = p
	}

	// 5. Key — if not found in the command
	finalKey := parsed.KeyPath
	if finalKey == "" {
		defaultKey := "~/.ssh/id_ed25519"
		home, _ := os.UserHomeDir()
		defaultExpanded := strings.Replace(defaultKey, "~", home, 1)
		if _, err := os.Stat(defaultExpanded); os.IsNotExist(err) {
			defaultKey = ""
		}
		hint := fmt.Sprintf(" [%s]", defaultKey)
		if defaultKey == "" {
			hint = " [none — password auth]"
		}
		finalKey = promptDefault(reader, bold.Sprintf("  Private key path%s: ", hint), defaultKey)
	}

	if err := sshpkg.ValidateKeyPath(finalKey); err != nil {
		return err
	}

	// 6. Group (optional)
	group := promptDefault(reader, bold.Sprint("  Group / tag (optional): "), "")

	// 7. Build connection + preview
	conn := config.Connection{
		Alias:   alias,
		Host:    parsed.Host,
		User:    parsed.User,
		Port:    finalPort,
		KeyPath: finalKey,
		Group:   group,
	}

	fmt.Println()
	dim.Println("  ─────────────────────────────────────")
	bold.Println("  Preview")
	dim.Println("  ─────────────────────────────────────")
	fmt.Printf("    Alias   : %s\n", conn.Alias)
	fmt.Printf("    Host    : %s\n", conn.Host)
	fmt.Printf("    User    : %s\n", conn.User)
	fmt.Printf("    Port    : %d\n", conn.Port)
	if conn.KeyPath != "" {
		fmt.Printf("    Key     : %s\n", conn.KeyPath)
	} else {
		fmt.Printf("    Key     : (none — password auth)\n")
	}
	if conn.Group != "" {
		fmt.Printf("    Group   : %s\n", conn.Group)
	}
	fmt.Println()
	fmt.Printf("    Command : %s\n", sshpkg.CommandString(conn))
	dim.Println("  ─────────────────────────────────────")
	fmt.Println()

	// 8. Confirm
	confirm := promptDefault(reader, bold.Sprint("  Save this connection? [Y/n]: "), "y")
	if strings.ToLower(confirm) != "y" && strings.ToLower(confirm) != "yes" {
		yellow.Println("\n  Aborted.")
		return nil
	}

	// 9. Save
	conns, err := config.Load()
	if err != nil {
		return err
	}
	conns, err = config.Add(conns, conn)
	if err != nil {
		return err
	}
	if err := config.Save(conns); err != nil {
		return err
	}

	fmt.Println()
	green.Printf("  ✓ Connection %q saved.\n\n", alias)
	return nil
}
