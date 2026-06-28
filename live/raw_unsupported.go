//go:build !windows && !linux

package live

import (
	"errors"
	"os"
)

func EnterConsoleRaw(input *os.File, output *os.File) (func() error, error) {
	return nil, errors.New("raw console non supportata su questo sistema")
}
