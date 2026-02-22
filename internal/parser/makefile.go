package parser

import (
	"bufio"
	"os"
	"strings"
)

// Target represents a Makefile target with its description.
type Target struct {
	Name        string
	Description string
}

// ParseMakefile reads a Makefile and extracts targets with their descriptions.
// Two documentation conventions are supported:
//
//	## Description
//	my-target:
//	    command
//
// Or inline:
//
//	my-target: ## Description
func ParseMakefile(path string) ([]Target, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var targets []Target
	var pendingDescription string

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()

		// A ## comment becomes the description for the next target
		if strings.HasPrefix(strings.TrimSpace(line), "##") {
			desc := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(line), "##"))
			pendingDescription = desc
			continue
		}

		// Target line: starts with a word followed by ":"
		// and is not a variable assignment
		if isTarget(line) {
			name := extractTargetName(line)
			if name == "" || strings.HasPrefix(name, ".") {
				pendingDescription = ""
				continue
			}

			desc := pendingDescription

			// Support inline description: target: ## description
			if idx := strings.Index(line, "##"); idx != -1 {
				desc = strings.TrimSpace(line[idx+2:])
			}

			targets = append(targets, Target{
				Name:        name,
				Description: desc,
			})
			pendingDescription = ""
			continue
		}

		// Any non-comment line resets the pending description
		if !strings.HasPrefix(strings.TrimSpace(line), "#") {
			pendingDescription = ""
		}
	}

	return targets, scanner.Err()
}

func isTarget(line string) bool {
	// A Make target starts without whitespace and contains ":"
	if len(line) == 0 || line[0] == ' ' || line[0] == '\t' {
		return false
	}
	colonIdx := strings.Index(line, ":")
	if colonIdx <= 0 {
		return false
	}
	// := or ::= assignment operator
	rest := line[colonIdx:]
	if strings.HasPrefix(rest, ":=") || strings.HasPrefix(rest, "::=") {
		return false
	}
	// = before : means it's a variable (=, ?=, +=, !=)
	before := line[:colonIdx]
	if strings.Contains(before, "=") {
		return false
	}
	// Target names cannot contain spaces
	name := strings.TrimSpace(before)
	if strings.Contains(name, " ") || strings.Contains(name, "\t") {
		return false
	}
	return true
}

func extractTargetName(line string) string {
	colonIdx := strings.Index(line, ":")
	if colonIdx <= 0 {
		return ""
	}
	return strings.TrimSpace(line[:colonIdx])
}
