package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	configDir  = ".config/sshm"
	configFile = "connections.json"
	dirPerm    = 0700
	filePerm   = 0600
)

// Connection represents a single SSH connection entry.
type Connection struct {
	Alias    string            `json:"alias"`
	Host     string            `json:"host"`
	User     string            `json:"user"`
	Port     int               `json:"port"`
	KeyPath  string            `json:"key_path,omitempty"`
	Group    string            `json:"group,omitempty"`
	Commands map[string]string `json:"commands,omitempty"`
}

// configPath returns the full path to connections.json.
func configPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("cannot determine home directory: %w", err)
	}
	return filepath.Join(home, configDir, configFile), nil
}

// Load reads and returns all connections from disk. Returns an empty slice if
// the file does not exist yet.
func Load() ([]Connection, error) {
	path, err := configPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return []Connection{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("reading config: %w", err)
	}

	var conns []Connection
	if err := json.Unmarshal(data, &conns); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}
	return conns, nil
}

// Save writes all connections to disk, creating the config directory if needed.
func Save(conns []Connection) error {
	path, err := configPath()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), dirPerm); err != nil {
		return fmt.Errorf("creating config dir: %w", err)
	}

	data, err := json.MarshalIndent(conns, "", "  ")
	if err != nil {
		return fmt.Errorf("serialising config: %w", err)
	}

	if err := os.WriteFile(path, data, filePerm); err != nil {
		return fmt.Errorf("writing config: %w", err)
	}
	return nil
}

// FindExact returns the connection whose alias matches exactly (case-insensitive).
func FindExact(conns []Connection, alias string) (Connection, bool) {
	lower := strings.ToLower(alias)
	for _, c := range conns {
		if strings.ToLower(c.Alias) == lower {
			return c, true
		}
	}
	return Connection{}, false
}

// FindExactWithIndex returns the matching connection and its index.
func FindExactWithIndex(conns []Connection, alias string) (Connection, int, bool) {
	lower := strings.ToLower(alias)
	for i, c := range conns {
		if strings.ToLower(c.Alias) == lower {
			return c, i, true
		}
	}
	return Connection{}, -1, false
}

// FindSubstring returns all connections whose alias, host, user or group
// contain the given query string (case-insensitive).
func FindSubstring(conns []Connection, query string) []Connection {
	lower := strings.ToLower(query)
	var results []Connection
	for _, c := range conns {
		if strings.Contains(strings.ToLower(c.Alias), lower) ||
			strings.Contains(strings.ToLower(c.Host), lower) ||
			strings.Contains(strings.ToLower(c.User), lower) ||
			strings.Contains(strings.ToLower(c.Group), lower) {
			results = append(results, c)
		}
	}
	return results
}

// Add appends a new connection after checking that the alias is unique.
func Add(conns []Connection, c Connection) ([]Connection, error) {
	if _, exists := FindExact(conns, c.Alias); exists {
		return conns, fmt.Errorf("alias %q already exists", c.Alias)
	}
	return append(conns, c), nil
}

// Remove deletes the connection with the given alias. Returns an error if the
// alias is not found.
func Remove(conns []Connection, alias string) ([]Connection, error) {
	lower := strings.ToLower(alias)
	updated := conns[:0:0] // nil-safe empty slice sharing backing array
	found := false
	for _, c := range conns {
		if strings.ToLower(c.Alias) == lower {
			found = true
			continue
		}
		updated = append(updated, c)
	}
	if !found {
		return conns, fmt.Errorf("alias %q not found", alias)
	}
	return updated, nil
}

// Update replaces the connection with the matching alias.
func Update(conns []Connection, updated Connection) ([]Connection, error) {
	lower := strings.ToLower(updated.Alias)
	for i, c := range conns {
		if strings.ToLower(c.Alias) == lower {
			conns[i] = updated
			return conns, nil
		}
	}
	return conns, fmt.Errorf("alias %q not found", updated.Alias)
}
