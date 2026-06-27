//go:build linux

package main

import (
	"os"
	"syscall"
	"unsafe"
)

func enterConsoleRaw(input *os.File, output *os.File) (func() error, error) {
	fd := uintptr(input.Fd())
	var original syscall.Termios
	if err := ioctl(fd, syscall.TCGETS, uintptr(unsafe.Pointer(&original))); err != nil {
		return nil, err
	}
	raw := original
	raw.Iflag &^= syscall.IGNBRK | syscall.BRKINT | syscall.PARMRK | syscall.ISTRIP | syscall.INLCR | syscall.IGNCR | syscall.ICRNL | syscall.IXON
	raw.Oflag &^= syscall.OPOST
	raw.Lflag &^= syscall.ECHO | syscall.ECHONL | syscall.ICANON | syscall.ISIG | syscall.IEXTEN
	raw.Cflag &^= syscall.CSIZE | syscall.PARENB
	raw.Cflag |= syscall.CS8
	raw.Cc[syscall.VMIN] = 1
	raw.Cc[syscall.VTIME] = 0
	if err := ioctl(fd, syscall.TCSETS, uintptr(unsafe.Pointer(&raw))); err != nil {
		return nil, err
	}
	return func() error {
		return ioctl(fd, syscall.TCSETS, uintptr(unsafe.Pointer(&original)))
	}, nil
}

func ioctl(fd uintptr, request uintptr, arg uintptr) error {
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, fd, request, arg)
	if errno != 0 {
		return errno
	}
	return nil
}
