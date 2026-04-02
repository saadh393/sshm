package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/saadh393/sshm/internal/config"
	sshpkg "github.com/saadh393/sshm/internal/ssh"
	"github.com/saadh393/sshm/internal/tui"
)

func runCommandBrowserFlow(conn config.Connection) error {
	for {
		browserResult := tui.RunCommandBrowser(conn)
		if browserResult.Quit {
			return nil
		}

		if browserResult.AddNew {
			formResult := tui.RunCommandForm(conn)
			if !formResult.Saved {
				continue
			}
			updatedConn, err := addCommandToConnection(conn.Alias, formResult.Name, formResult.Command)
			if err != nil {
				return err
			}
			conn = updatedConn
			green := color.New(color.FgGreen, color.Bold)
			fmt.Fprintf(os.Stdout, "%s Added command %q to %q\n", green.Sprint("✓"), formResult.Name, conn.Alias)
			continue
		}

		blue := color.New(color.FgCyan)
		cmdStr := sshpkg.RemoteCommandString(conn, browserResult.Command)
		fmt.Fprintf(os.Stderr, "%s Running: %s\n", blue.Sprint("→"), cmdStr)
		return sshpkg.ConnectRemoteCommand(conn, browserResult.Command)
	}
}

func addCommandToConnection(alias, name, remoteCommand string) (config.Connection, error) {
	conns, idx, conn, err := loadConnectionAndIndex(alias)
	if err != nil {
		return config.Connection{}, err
	}
	if conn.Commands == nil {
		conn.Commands = make(map[string]string)
	}
	if _, exists := conn.Commands[name]; exists {
		return config.Connection{}, fmt.Errorf("command %q already exists for connection %q", name, conn.Alias)
	}
	name = strings.TrimSpace(name)
	remoteCommand = strings.TrimSpace(remoteCommand)
	if name == "" {
		return config.Connection{}, fmt.Errorf("command name cannot be empty")
	}
	if remoteCommand == "" {
		return config.Connection{}, fmt.Errorf("remote command cannot be empty")
	}

	conn.Commands[name] = remoteCommand
	conns[idx] = conn
	if err := config.Save(conns); err != nil {
		return config.Connection{}, err
	}
	return conn, nil
}
