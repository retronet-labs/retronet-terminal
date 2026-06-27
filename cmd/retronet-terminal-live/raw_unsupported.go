//go:build !windows && !linux

package main

import (
	"errors"
	"os"
)

func enterConsoleRaw(input *os.File, output *os.File) (func() error, error) {
	return nil, errors.New("raw console non supportata su questo sistema")
}
