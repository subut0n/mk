package ui

import (
	"fmt"
	"os"
	"strings"
)

// CaptureKey enters raw mode and captures a single printable key press.
// Non-printable keys, '/' (reserved for filter), Enter, and Escape are rejected.
func CaptureKey() (byte, error) {
	oldState, err := makeRaw()
	if err != nil {
		return 0, err
	}
	defer restoreTerminal(oldState)

	for {
		b := make([]byte, 4)
		n, err := os.Stdin.Read(b)
		if err != nil {
			return 0, err
		}
		if n == 0 {
			continue
		}

		// Skip escape sequences (arrows, function keys, etc.)
		if b[0] == 27 {
			continue
		}

		// Skip Enter, Ctrl+C, Tab, and Backspace
		if b[0] == 13 || b[0] == 3 || b[0] == 9 || b[0] == 127 || b[0] == 8 {
			continue
		}

		// Accept printable ASCII (33-126), excluding '/' (reserved for filter)
		if b[0] >= 33 && b[0] <= 126 && b[0] != '/' {
			return b[0], nil
		}
	}
}

// KeyDisplayName returns a human-readable label for a key byte.
func KeyDisplayName(b byte) string {
	if b >= 33 && b <= 126 {
		return strings.ToUpper(string(b))
	}
	return fmt.Sprintf("0x%02x", b)
}
