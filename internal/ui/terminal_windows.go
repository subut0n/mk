//go:build windows

package ui

import "fmt"

type termState struct{}

func makeRaw() (*termState, error) {
	return nil, fmt.Errorf("raw mode not supported on Windows")
}

func restoreTerminal(state *termState) {}
