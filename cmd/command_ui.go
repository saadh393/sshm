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

type commandBrowserExit int

const (
	commandBrowserExitBack commandBrowserExit = iota
	commandBrowserExitConnect
)

func runCommandBrowserFlow(conn config.Connection) (commandBrowserExit, error) {
	status := ""
	for {
		browserResult := tui.RunCommandBrowser(conn, status)
		status = ""
		if browserResult.Quit {
			return commandBrowserExitBack, nil
		}

		if browserResult.AddNew {
			formResult := tui.RunCommandForm(conn)
			if !formResult.Saved {
				continue
			}
			updatedConn, err := addCommandToConnection(conn.Alias, formResult.Name, formResult.Command)
			if err != nil {
				status = fmt.Sprintf("Failed to add command %q: %s", formResult.Name, err.Error())
				continue
			}
			conn = updatedConn
			green := color.New(color.FgGreen, color.Bold)
			fmt.Fprintf(os.Stdout, "%s Added command %q to %q\n", green.Sprint("✓"), formResult.Name, conn.Alias)
			continue
		}

		if browserResult.Update {
			formResult := tui.RunUpdateCommandForm(conn, browserResult.Name, browserResult.Command)
			if !formResult.Saved {
				continue
			}
			updatedConn, err := updateCommandOnConnection(conn.Alias, browserResult.Name, formResult.Name, formResult.Command)
			if err != nil {
				status = fmt.Sprintf("Failed to update command %q: %s", browserResult.Name, err.Error())
				continue
			}
			conn = updatedConn
			green := color.New(color.FgGreen, color.Bold)
			fmt.Fprintf(os.Stdout, "%s Updated command %q on %q\n", green.Sprint("✓"), formResult.Name, conn.Alias)
			continue
		}

		if browserResult.Delete {
			confirm := tui.RunConfirm(
				"Delete Saved Command",
				fmt.Sprintf("Delete command %q from %q?\n\nCommand: %s", browserResult.Name, conn.Alias, browserResult.Command),
			)
			if !confirm.Confirmed {
				status = "Delete canceled."
				continue
			}
			updatedConn, err := deleteCommandFromConnection(conn.Alias, browserResult.Name)
			if err != nil {
				status = fmt.Sprintf("Failed to delete command %q: %s", browserResult.Name, err.Error())
				continue
			}
			conn = updatedConn
			red := color.New(color.FgRed, color.Bold)
			fmt.Fprintf(os.Stdout, "%s Deleted command %q from %q\n", red.Sprint("✗"), browserResult.Name, conn.Alias)
			continue
		}

		blue := color.New(color.FgCyan)
		cmdStr := sshpkg.RemoteCommandString(conn, browserResult.Command)
		fmt.Fprintf(os.Stderr, "%s Running: %s\n", blue.Sprint("→"), cmdStr)
		return commandBrowserExitConnect, sshpkg.ConnectRemoteCommand(conn, browserResult.Command)
	}
}

func runConnectionListFlow() error {
	for {
		conns, err := config.Load()
		if err != nil {
			return fmt.Errorf("failed to load saved connections: %w", err)
		}
		if len(conns) == 0 {
			fmt.Fprintln(os.Stdout, "No connections saved. Run 'sshm add' to add one.")
			return nil
		}
		result := tui.Run(conns)
		if result.Quit || result.Conn == nil {
			return nil
		}
		if result.OpenCommands {
			exit, err := runCommandBrowserFlow(*result.Conn)
			if err != nil {
				return err
			}
			if exit == commandBrowserExitBack {
				continue
			}
			return nil
		}
		return doConnect(*result.Conn, false)
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
	name = strings.TrimSpace(name)
	remoteCommand = strings.TrimSpace(remoteCommand)
	if name == "" {
		return config.Connection{}, fmt.Errorf("command name cannot be empty")
	}
	if remoteCommand == "" {
		return config.Connection{}, fmt.Errorf("remote command cannot be empty")
	}
	if _, exists := conn.Commands[name]; exists {
		return config.Connection{}, fmt.Errorf("command %q already exists for connection %q", name, conn.Alias)
	}

	conn.Commands[name] = remoteCommand
	conns[idx] = conn
	if err := config.Save(conns); err != nil {
		return config.Connection{}, err
	}
	return conn, nil
}

func updateCommandOnConnection(alias, oldName, newName, remoteCommand string) (config.Connection, error) {
	conns, idx, conn, err := loadConnectionAndIndex(alias)
	if err != nil {
		return config.Connection{}, err
	}
	if conn.Commands == nil {
		return config.Connection{}, fmt.Errorf("no saved commands for connection %q", conn.Alias)
	}
	oldName = strings.TrimSpace(oldName)
	newName = strings.TrimSpace(newName)
	remoteCommand = strings.TrimSpace(remoteCommand)
	if oldName == "" {
		return config.Connection{}, fmt.Errorf("existing command name cannot be empty")
	}
	if newName == "" {
		return config.Connection{}, fmt.Errorf("command name cannot be empty")
	}
	if remoteCommand == "" {
		return config.Connection{}, fmt.Errorf("remote command cannot be empty")
	}
	if _, exists := conn.Commands[oldName]; !exists {
		return config.Connection{}, fmt.Errorf("command %q not found for connection %q", oldName, conn.Alias)
	}
	if oldName != newName {
		if _, exists := conn.Commands[newName]; exists {
			return config.Connection{}, fmt.Errorf("command %q already exists for connection %q", newName, conn.Alias)
		}
		delete(conn.Commands, oldName)
	}
	conn.Commands[newName] = remoteCommand
	conns[idx] = conn
	if err := config.Save(conns); err != nil {
		return config.Connection{}, err
	}
	return conn, nil
}

func deleteCommandFromConnection(alias, name string) (config.Connection, error) {
	conns, idx, conn, err := loadConnectionAndIndex(alias)
	if err != nil {
		return config.Connection{}, err
	}
	if conn.Commands == nil {
		return config.Connection{}, fmt.Errorf("no saved commands for connection %q", conn.Alias)
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return config.Connection{}, fmt.Errorf("command name cannot be empty")
	}
	if _, exists := conn.Commands[name]; !exists {
		return config.Connection{}, fmt.Errorf("command %q not found for connection %q", name, conn.Alias)
	}
	delete(conn.Commands, name)
	if len(conn.Commands) == 0 {
		conn.Commands = nil
	}
	conns[idx] = conn
	if err := config.Save(conns); err != nil {
		return config.Connection{}, err
	}
	return conn, nil
}
