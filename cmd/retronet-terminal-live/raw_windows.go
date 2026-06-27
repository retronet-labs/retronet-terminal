//go:build windows

package main

import (
	"os"
	"syscall"
)

const (
	enableProcessedInput            = 0x0001
	enableLineInput                 = 0x0002
	enableEchoInput                 = 0x0004
	enableVirtualTerminalInput      = 0x0200
	enableProcessedOutput           = 0x0001
	enableVirtualTerminalProcessing = 0x0004
)

var procSetConsoleMode = syscall.NewLazyDLL("kernel32.dll").NewProc("SetConsoleMode")

func enterConsoleRaw(input *os.File, output *os.File) (func() error, error) {
	inputHandle := syscall.Handle(input.Fd())
	var inputMode uint32
	if err := syscall.GetConsoleMode(inputHandle, &inputMode); err != nil {
		return nil, err
	}
	rawInputMode := inputMode
	rawInputMode &^= enableProcessedInput | enableLineInput | enableEchoInput
	rawInputMode |= enableVirtualTerminalInput
	if err := setConsoleMode(inputHandle, rawInputMode); err != nil {
		return nil, err
	}

	outputHandle := syscall.Handle(output.Fd())
	var outputMode uint32
	outputModeOK := syscall.GetConsoleMode(outputHandle, &outputMode) == nil
	if outputModeOK {
		_ = setConsoleMode(outputHandle, outputMode|enableProcessedOutput|enableVirtualTerminalProcessing)
	}

	return func() error {
		err := setConsoleMode(inputHandle, inputMode)
		if outputModeOK {
			if outputErr := setConsoleMode(outputHandle, outputMode); err == nil {
				err = outputErr
			}
		}
		return err
	}, nil
}

func setConsoleMode(handle syscall.Handle, mode uint32) error {
	result, _, err := procSetConsoleMode.Call(uintptr(handle), uintptr(mode))
	if result != 0 {
		return nil
	}
	if err != syscall.Errno(0) {
		return err
	}
	return syscall.EINVAL
}
