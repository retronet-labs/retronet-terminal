//go:build windows

package live

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

func EnterConsoleRaw(input *os.File, output *os.File) (func() error, error) {
	inputHandle := syscall.Handle(input.Fd())
	var inputMode uint32
	if err := syscall.GetConsoleMode(inputHandle, &inputMode); err != nil {
		return nil, err
	}
	rawInputMode := inputMode
	rawInputMode &^= enableProcessedInput | enableLineInput | enableEchoInput
	rawInputMode |= enableVirtualTerminalInput
	if err := setConsoleMode(inputHandle, rawInputMode); err != nil {
		rawInputMode &^= enableVirtualTerminalInput
		if fallbackErr := setConsoleMode(inputHandle, rawInputMode); fallbackErr != nil {
			return nil, err
		}
	}

	outputHandle := syscall.Handle(output.Fd())
	var outputMode uint32
	if err := syscall.GetConsoleMode(outputHandle, &outputMode); err != nil {
		_ = setConsoleMode(inputHandle, inputMode)
		return nil, err
	}
	if err := setConsoleMode(outputHandle, outputMode|enableProcessedOutput|enableVirtualTerminalProcessing); err != nil {
		_ = setConsoleMode(inputHandle, inputMode)
		return nil, err
	}

	return func() error {
		err := setConsoleMode(inputHandle, inputMode)
		if outputErr := setConsoleMode(outputHandle, outputMode); err == nil {
			err = outputErr
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
