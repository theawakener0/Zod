//go:build !windows

package object

import (
	"fmt"
	"os"
	"syscall"

	"golang.org/x/sys/unix"
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
	fd := int(os.Stdin.Fd())
	var rfds unix.FdSet
	rfds.Set(fd)
	tv := unix.NsecToTimeval(0)
	n, err := unix.Select(fd+1, &rfds, nil, nil, &tv)
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

func termReadKey() (string, error) {
	fd := int(os.Stdin.Fd())

	var saved *term.State
	if !termRawActive {
		if term.IsTerminal(fd) {
			old, err := term.MakeRaw(fd)
			if err != nil {
				return "", fmt.Errorf("could not set terminal to raw mode: %w", err)
			}
			saved = old
			defer func() {
				_ = term.Restore(fd, saved)
			}()
		}
	}

	var b [1]byte
	n, err := syscall.Read(fd, b[:])
	if err != nil {
		return "", fmt.Errorf("could not read key: %w", err)
	}
	if n == 0 {
		return "", fmt.Errorf("could not read key: EOF")
	}
	if b[0] != 0x1b {
		return string(b[:1]), nil
	}

	sep := []byte{b[0]}
	for i := 0; i < 5; i++ {
		ok, err := termHasInput()
		if err != nil || !ok {
			break
		}
		var eb [1]byte
		m, err := syscall.Read(fd, eb[:])
		if err != nil || m == 0 {
			break
		}
		sep = append(sep, eb[0])
		if (eb[0] >= 'A' && eb[0] <= 'Z') || (eb[0] >= 'a' && eb[0] <= 'z') || eb[0] == '~' {
			break
		}
	}
	return string(sep), nil
}
