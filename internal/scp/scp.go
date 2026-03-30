package scp

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/saadh393/sshm/internal/config"
)

// Direction indicates which way the file transfer goes.
type Direction int

const (
	Upload   Direction = iota // local  → remote
	Download                  // remote → local
)

func (d Direction) String() string {
	if d == Upload {
		return "upload"
	}
	return "download"
}

// BuildArgs returns the scp argument list for the transfer.
// Upload:   scp [flags] localPath  user@host:remotePath
// Download: scp [flags] user@host:remotePath  localPath
func BuildArgs(c config.Connection, localPath, remotePath string, dir Direction) []string {
	args := []string{"scp"}

	if c.Port != 0 && c.Port != 22 {
		args = append(args, "-P", strconv.Itoa(c.Port))
	}

	if c.KeyPath != "" {
		args = append(args, "-i", expandTilde(c.KeyPath))
	}

	remote := fmt.Sprintf("%s@%s:%s", c.User, c.Host, remotePath)

	if dir == Upload {
		args = append(args, expandTilde(localPath), remote)
	} else {
		args = append(args, remote, expandTilde(localPath))
	}

	return args
}

// CommandString returns a human-readable scp command string.
func CommandString(c config.Connection, localPath, remotePath string, dir Direction) string {
	return strings.Join(BuildArgs(c, localPath, remotePath, dir), " ")
}

// Copy runs the scp transfer, streaming stdout/stderr directly to the terminal.
// Unlike ssh.Connect it uses exec.Command (not syscall.Exec) so control returns
// to sshm after the transfer completes.
func Copy(c config.Connection, localPath, remotePath string, dir Direction) error {
	scpPath, err := exec.LookPath("scp")
	if err != nil {
		return fmt.Errorf("scp not found in PATH: %w", err)
	}

	args := BuildArgs(c, localPath, remotePath, dir)
	cmd := exec.Command(scpPath, args[1:]...) // args[0] is "scp", skip it
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func expandTilde(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err == nil {
			return strings.Replace(path, "~", home, 1)
		}
	}
	return path
}
