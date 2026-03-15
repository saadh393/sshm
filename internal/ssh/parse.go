package ssh

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// ParsedCommand holds fields extracted from a raw ssh command string.
type ParsedCommand struct {
	User    string
	Host    string
	Port    int    // 0 means not specified
	KeyPath string
}

// ParseCommand parses a raw ssh command string such as:
//
//	ssh -i ~/.ssh/key user@host -p 2222
//
// It extracts user, host, port, and identity key.
func ParseCommand(input string) (ParsedCommand, error) {
	input = strings.TrimSpace(input)
	// Strip leading "ssh" binary name
	if strings.HasPrefix(input, "ssh ") {
		input = strings.TrimPrefix(input, "ssh ")
	} else if input == "ssh" {
		return ParsedCommand{}, fmt.Errorf("no target specified")
	}

	var result ParsedCommand
	tokens := tokenize(input)

	// Flags that consume the next token as a value
	valueFlags := regexp.MustCompile(`^-[bcDEeFIiJlmnoOpQRSWw]$`)

	for i := 0; i < len(tokens); i++ {
		tok := tokens[i]
		switch tok {
		case "-i":
			i++
			if i >= len(tokens) {
				return result, fmt.Errorf("missing value after -i")
			}
			result.KeyPath = tokens[i]
		case "-p":
			i++
			if i >= len(tokens) {
				return result, fmt.Errorf("missing value after -p")
			}
			p, err := strconv.Atoi(tokens[i])
			if err != nil || p < 1 || p > 65535 {
				return result, fmt.Errorf("invalid port %q", tokens[i])
			}
			result.Port = p
		default:
			if strings.HasPrefix(tok, "-") {
				if valueFlags.MatchString(tok) {
					i++ // consume the flag's value
				}
				continue
			}
			// First non-flag token is the destination: [user@]host
			if result.Host == "" {
				if strings.Contains(tok, "@") {
					parts := strings.SplitN(tok, "@", 2)
					result.User = parts[0]
					result.Host = parts[1]
				} else {
					result.Host = tok
				}
			}
		}
	}

	if result.Host == "" {
		return result, fmt.Errorf("could not find host in SSH command")
	}
	return result, nil
}

// tokenize splits a string by spaces, respecting single and double quoted segments.
func tokenize(s string) []string {
	var tokens []string
	var cur strings.Builder
	inQuote := false
	quoteChar := rune(0)

	for _, r := range s {
		switch {
		case inQuote && r == quoteChar:
			inQuote = false
		case !inQuote && (r == '"' || r == '\''):
			inQuote = true
			quoteChar = r
		case !inQuote && r == ' ':
			if cur.Len() > 0 {
				tokens = append(tokens, cur.String())
				cur.Reset()
			}
		default:
			cur.WriteRune(r)
		}
	}
	if cur.Len() > 0 {
		tokens = append(tokens, cur.String())
	}
	return tokens
}
