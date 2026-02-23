//go:build linux

package ui

import (
	"os"
	"syscall"
	"unsafe"
)

type termios struct {
	Iflag  uint32
	Oflag  uint32
	Cflag  uint32
	Lflag  uint32
	Cc     [20]uint8
	Ispeed uint32
	Ospeed uint32
}

type termState struct {
	t termios
}

func makeRaw() (*termState, error) {
	fd := os.Stdin.Fd()
	var t termios
	if _, _, errno := syscall.Syscall(syscall.SYS_IOCTL, fd, syscall.TCGETS, uintptr(unsafe.Pointer(&t))); errno != 0 {
		return nil, errno
	}
	old := &termState{t}

	t.Lflag &^= syscall.ECHO | syscall.ICANON | syscall.ISIG
	t.Iflag &^= syscall.IXON | syscall.ICRNL
	t.Cc[syscall.VMIN] = 1
	t.Cc[syscall.VTIME] = 0

	if _, _, errno := syscall.Syscall(syscall.SYS_IOCTL, fd, syscall.TCSETS, uintptr(unsafe.Pointer(&t))); errno != 0 {
		return nil, errno
	}
	return old, nil
}

func restoreTerminal(state *termState) {
	if state == nil {
		return
	}
	fd := os.Stdin.Fd()
	syscall.Syscall(syscall.SYS_IOCTL, fd, syscall.TCSETS, uintptr(unsafe.Pointer(&state.t)))
}
