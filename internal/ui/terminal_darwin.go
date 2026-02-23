//go:build darwin

package ui

import (
	"os"
	"syscall"
	"unsafe"
)

type termios struct {
	Iflag  uint64
	Oflag  uint64
	Cflag  uint64
	Lflag  uint64
	Cc     [20]uint8
	Ispeed uint64
	Ospeed uint64
}

type termState struct {
	t termios
}

const (
	ioctlGetTermios = syscall.TIOCGETA
	ioctlSetTermios = syscall.TIOCSETA
)

func makeRaw() (*termState, error) {
	fd := os.Stdin.Fd()
	var t termios
	if _, _, errno := syscall.Syscall(syscall.SYS_IOCTL, fd, ioctlGetTermios, uintptr(unsafe.Pointer(&t))); errno != 0 {
		return nil, errno
	}
	old := &termState{t}

	t.Lflag &^= syscall.ECHO | syscall.ICANON | syscall.ISIG
	t.Iflag &^= syscall.IXON | syscall.ICRNL
	t.Cc[syscall.VMIN] = 1
	t.Cc[syscall.VTIME] = 0

	if _, _, errno := syscall.Syscall(syscall.SYS_IOCTL, fd, ioctlSetTermios, uintptr(unsafe.Pointer(&t))); errno != 0 {
		return nil, errno
	}
	return old, nil
}

func restoreTerminal(state *termState) {
	if state == nil {
		return
	}
	fd := os.Stdin.Fd()
	syscall.Syscall(syscall.SYS_IOCTL, fd, ioctlSetTermios, uintptr(unsafe.Pointer(&state.t)))
}
