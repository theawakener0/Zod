//go:build windows

package object

import (
	"fmt"
	"os"

	"golang.org/x/term"
)

var (
	termRawActive bool
	termOldState  *term.State
)

func TermDisableRaw() error {
	return termDisableRaw()
}

func termEnableRaw() error {
	if termRawActive {
		return nil
	}
	fd := int(os.Stdin.Fd())
	if !term.IsTerminal(fd) {
		return fmt.Errorf("not a TTY")
	}
	old, err := term.MakeRaw(fd)
	if err != nil {
		return err
	}
	termOldState = old
	termRawActive = true

	return nil
}

func termDisableRaw() error {
	if !termRawActive {
		return nil
	}
	fd := int(os.Stdin.Fd())
	if termOldState != nil {
		err := term.Restore(fd, termOldState)
		if err != nil {
			return err
		}
	}
	termRawActive = false
	termOldState = nil

	return nil
}

func termWinsizeDim(width bool) int {
	for _, f := range []int{int(os.Stdout.Fd()), int(os.Stdin.Fd())} {
		w, h, err := term.GetSize(f)
		if err != nil {
			continue
		}
		if width && w > 0 {
			return w
		}
		if !width && h > 0 {
			return h
		}
	}
	return 0
}

func termHasInput() (bool, error) {
	// I don't know what to do here
	return false, nil
}

func termReadKey() (string, error) {
	var b [1]byte
	n, err := os.Stdin.Read(b[:])
	if err != nil {
		return "", fmt.Errorf("could not read key: %w", err)
	}
	if n == 0 {
		return "", fmt.Errorf("could not read key: EOF")
	}
	return string(b[:1]), nil
}
