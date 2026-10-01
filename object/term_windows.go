//go:build windows

package object

import (
	"fmt"
	"os"

	"golang.org/x/sys/windows"
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
	handle := windows.Handle(os.Stdin.Fd())
	if handle == 0 || handle == windows.InvalidHandle {
		return false, nil
	}
	// Console input handles are signaled when the input buffer is
	// non-empty; pipe handles when data is available. Never return an
	// error: an Error object is truthy in Zod and poll_key would block
	// in read_key.
	event, err := windows.WaitForSingleObject(handle, 0)
	if err != nil {
		return false, nil
	}
	return event == windows.WAIT_OBJECT_0, nil
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
	if b[0] != 0x1b {
		return string(b[:1]), nil
	}

	// Assemble the rest of the escape sequence. With
	// ENABLE_VIRTUAL_TERMINAL_INPUT (set by x/term's raw mode) arrow keys
	// arrive as "\x1b[A" etc., possibly across several reads.
	return collectEscapeSequence(b[0], func() (byte, bool) {
		ok, err := termHasInput()
		if err != nil || !ok {
			return 0, false
		}
		var eb [1]byte
		m, err := os.Stdin.Read(eb[:])
		if err != nil || m == 0 {
			return 0, false
		}
		return eb[0], true
	}), nil
}
