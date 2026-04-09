package ssh

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"

	"github.com/saadh393/sshm/internal/config"
)

// BuildArgs returns the argument list (including the "ssh" argv[0]) for the
// given connection.
func BuildArgs(c config.Connection) []string {
	args := []string{"ssh"}

	if c.Port != 0 && c.Port != 22 {
		args = append(args, "-p", strconv.Itoa(c.Port))
	}

	if c.KeyPath != "" {
		keyPath := expandTilde(c.KeyPath)
		args = append(args, "-i", keyPath)
	}

	target := fmt.Sprintf("%s@%s", c.User, c.Host)
	args = append(args, target)
	return args
}

// CommandString returns a human-readable ssh command string.
func CommandString(c config.Connection) string {
	return strings.Join(BuildArgs(c), " ")
}

// BuildRemoteCommandArgs returns ssh argv with a remote command appended.
func BuildRemoteCommandArgs(c config.Connection, remoteCommand string) []string {
	args := BuildArgs(c)
	args = append(args, remoteCommand)
	return args
}

// RemoteCommandString returns a human-readable ssh command string for a remote command.
func RemoteCommandString(c config.Connection, remoteCommand string) string {
	return strings.Join(BuildRemoteCommandArgs(c, remoteCommand), " ")
}

// Connect replaces the current process with the ssh binary using syscall.Exec.
func Connect(c config.Connection) error {
	sshPath, err := exec.LookPath("ssh")
	if err != nil {
		return fmt.Errorf("ssh not found in PATH: %w", err)
	}

	args := BuildArgs(c)
	// argv[0] should be the binary path
	args[0] = sshPath

	return syscall.Exec(sshPath, args, os.Environ())
}

// ConnectRemoteCommand replaces the current process with ssh executing a remote command.
func ConnectRemoteCommand(c config.Connection, remoteCommand string) error {
	sshPath, err := exec.LookPath("ssh")
	if err != nil {
		return fmt.Errorf("ssh not found in PATH: %w", err)
	}

	args := BuildRemoteCommandArgs(c, remoteCommand)
	args[0] = sshPath

	return syscall.Exec(sshPath, args, os.Environ())
}

// expandTilde replaces a leading ~ with the user's home directory.
func expandTilde(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err == nil {
			return strings.Replace(path, "~", home, 1)
		}
	}
	return path
}

// ValidateKeyPath checks that the key file exists and is readable.
func ValidateKeyPath(path string) error {
	if path == "" {
		return nil
	}
	expanded := expandTilde(path)
	info, err := os.Stat(expanded)
	if err != nil {
		return fmt.Errorf("key file %q: %w", path, err)
	}
	if info.IsDir() {
		return fmt.Errorf("key path %q is a directory", path)
	}
	return nil
}
