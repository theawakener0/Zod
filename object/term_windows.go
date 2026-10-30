//go:build windows

package object

import (
	"fmt"
	"os"
	"unsafe"

	"golang.org/x/sys/windows"
	"golang.org/x/term"
)

// PeekNamedPipe is not wrapped by x/sys/windows. It reports how many
// bytes are buffered in a pipe without consuming them or blocking.
var procPeekNamedPipe = windows.NewLazySystemDLL("kernel32.dll").NewProc("PeekNamedPipe")

func TermDisableRaw() error {
	return termDisableRaw()
}

func termEnableRaw() error {
	termStateMu.Lock()
	defer termStateMu.Unlock()
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
	mouseErr := termDisableMouse()
	termStateMu.Lock()
	defer termStateMu.Unlock()
	if !termRawActive {
		return mouseErr
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

	return mouseErr
}

// termRestoreRawOnly restores the console input mode without touching
// stdout, so signal cleanup cannot block on a full output buffer.
func termRestoreRawOnly() {
	termStateMu.Lock()
	old := termOldState
	active := termRawActive
	termRawActive = false
	termOldState = nil
	termStateMu.Unlock()
	if active && old != nil {
		_ = term.Restore(int(os.Stdin.Fd()), old)
	}
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
	// WaitForSingleObject reports input readiness only for console
	// input handles. A synchronous pipe handle is effectively always
	// signaled, even when its buffer is empty, so waiting on it would
	// claim input and the subsequent read would block. Query the pipe
	// directly instead.
	if ft, err := windows.GetFileType(handle); err == nil && ft == windows.FILE_TYPE_PIPE {
		var avail uint32
		r1, _, _ := procPeekNamedPipe.Call(
			uintptr(handle),
			0,                               // lpBuffer
			0,                               // nBufferSize
			0,                               // lpBytesRead
			uintptr(unsafe.Pointer(&avail)), // lpTotalBytesAvail
			0,                               // lpBytesLeftThisMessage
		)
		if r1 == 0 {
			// A broken or unqueryable pipe must not look readable, or
			// poll_key would enter a blocking read.
			return false, nil
		}
		return avail > 0, nil
	}
	// Console input handles are signaled when the input buffer is
	// non-empty. Never return an error: an Error object is truthy in
	// Zod and poll_key would block in read_key.
	event, err := windows.WaitForSingleObject(handle, 0)
	if err != nil {
		return false, nil
	}
	return event == windows.WAIT_OBJECT_0, nil
}

func termReadKeyRaw() (string, error) {
	if termSGRDiscard {
		ok, _ := termHasInput()
		if !ok {
			return "", nil
		}
	}
	var b [1]byte
	n, err := os.Stdin.Read(b[:])
	if err != nil {
		return "", fmt.Errorf("could not read key: %w", err)
	}
	if n == 0 {
		return "", nil
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
