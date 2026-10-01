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
	// MakeRaw clears OPOST, so the kernel stops translating "\n" to
	// "\r\n". Zod renders full-screen frames (std/canvas) with bare
	// newlines, which would otherwise print as a staircase while raw
	// mode is active. Put OPOST/ONLCR back; term.Restore reverts this
	// together with everything else MakeRaw changed.
	if err := termKeepOutputTranslation(fd); err != nil {
		_ = term.Restore(fd, old)
		return fmt.Errorf("could not restore output translation: %w", err)
	}
	termOldState = old
	termRawActive = true

	return nil
}

// termKeepOutputTranslation re-enables OPOST/ONLCR on fd, which
// term.MakeRaw disabled. Raw mode must only change *input* processing;
// Zod's output relies on the kernel newline translation.
func termKeepOutputTranslation(fd int) error {
	t, err := unix.IoctlGetTermios(fd, ioctlReadTermios)
	if err != nil {
		return err
	}
	t.Oflag |= unix.OPOST | unix.ONLCR
	return unix.IoctlSetTermios(fd, ioctlWriteTermios, t)
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
	for {
		var rfds unix.FdSet
		rfds.Set(fd)
		tv := unix.NsecToTimeval(0)
		n, err := unix.Select(fd+1, &rfds, nil, nil, &tv)
		if err == unix.EINTR {
			continue
		}
		if err != nil {
			// Report "no input" rather than an error: an Error object is
			// truthy in Zod, and std/term's poll_key would treat it as
			// "input ready" and block in read_key (freezing the games).
			return false, nil
		}
		return n > 0, nil
	}
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

	// The press started with ESC: collect the rest of the escape
	// sequence ("\x1b[A" etc.), tolerating split delivery.
	return collectEscapeSequence(b[0], func() (byte, bool) {
		ok, err := termHasInput()
		if err != nil || !ok {
			return 0, false
		}
		var eb [1]byte
		m, err := syscall.Read(fd, eb[:])
		if err != nil || m == 0 {
			return 0, false
		}
		return eb[0], true
	}), nil
}
