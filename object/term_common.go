package object

import (
	"fmt"
	"os"
	"sync"
	"time"

	"golang.org/x/term"
)

// escapeSeqTimeout bounds how long termReadKey waits for the *next* byte
// after an ESC (0x1b) or after the previous byte of an escape sequence.
// A terminal may deliver the bytes of a single key press or SGR mouse
// report in separate reads; an idle timeout (rather than one deadline for
// the whole sequence) lets a slow but complete report assemble without
// turning it into stray keystrokes. A lone ESC still resolves after this
// timeout, matching std/term's "esc" key handling.
const escapeSeqTimeout = 100 * time.Millisecond

// escapeSeqTotalTimeout prevents a byte stream that never terminates from
// keeping the reader busy forever. Normal keys and mouse reports are far
// shorter than this.
const escapeSeqTotalTimeout = 1 * time.Second

// maxEscapeSeqBytes caps an assembled sequence (ESC plus up to 31 more
// bytes). 8 bytes was enough for function keys but truncated SGR mouse
// reports such as "\x1b[<0;120;40M" (15 bytes). 32 bytes covers the
// longest realistic SGR event (button + two four-digit coordinates).
const maxEscapeSeqBytes = 32

// Mouse tracking is opt-in. We use 1000 (button press/release), 1002
// (motion while a button is held) and 1006 (SGR encoding). 1002 is
// deliberate: 1003 would report every pointer movement even when no
// button is down, flooding the input queue.
const (
	mouseEnableSeq  = "\x1b[?1000h\x1b[?1002h\x1b[?1006h"
	mouseDisableSeq = "\x1b[?1006l\x1b[?1002l\x1b[?1000l"
	cursorShowSeq   = "\x1b[?25h"
	altEnterSeq     = "\x1b[?1049h"
	altLeaveSeq     = "\x1b[?1049l"
	// terminalResetSeq clears attributes, ends synchronized output, and
	// restores autowrap. Alternate-screen teardown is tracked separately
	// so plain scripts that never entered it do not emit DECRST 1049.
	terminalResetSeq = "\x1b[0m\x1b[?2026l\x1b[?7h"
)

var (
	// termStateMu makes enable/disable/cleanup transitions linearizable
	// with the signal-cleanup goroutine. Critical sections are short and
	// never wait for input.
	termStateMu sync.Mutex

	termMouseActive bool
	termAltActive   bool
	termRawActive   bool
	termOldState    *term.State

	// A truncated SGR report can leave its tail in the input stream.
	// These fields make termReadKey swallow only bytes that could still
	// belong to that report; any other byte ends the discard and is
	// returned as a normal key.
	termSGRDiscard      bool
	termSGRDiscardUntil time.Time
	termSGRDiscardBytes int
)

// stdoutIsTerminal reports whether stdout is a real terminal. A character
// device such as /dev/null is not enough: writing control sequences there
// would be silently lost, while pipes and files must stay byte-clean.
func stdoutIsTerminal() bool {
	return term.IsTerminal(int(os.Stdout.Fd()))
}

// termEnableMouse turns on button-event mouse tracking and SGR encoding.
// Raw mode must already be active so the terminal does not echo the
// escape sequence or line-buffer the mouse reports.
func termEnableMouse() error {
	termStateMu.Lock()
	if termMouseActive {
		termStateMu.Unlock()
		return nil
	}
	if !termRawActive {
		termStateMu.Unlock()
		return fmt.Errorf("raw mode must be enabled before mouse mode")
	}
	if !stdoutIsTerminal() {
		termStateMu.Unlock()
		return fmt.Errorf("stdout is not a terminal")
	}
	// Mark active before writing: if the write fails after a partial
	// sequence reached the terminal, cleanup must still try to disable it.
	termMouseActive = true
	termStateMu.Unlock()
	// Never hold termStateMu across a terminal write. A blocked write
	// must not prevent signal cleanup from restoring raw mode.
	if _, err := fmt.Fprint(os.Stdout, mouseEnableSeq); err != nil {
		return fmt.Errorf("could not enable mouse: %w", err)
	}
	return nil
}

// termDisableMouse restores the terminal's normal mouse behaviour. It is
// idempotent and safe to call from a cleanup defer.
func termDisableMouse() error {
	termStateMu.Lock()
	if !termMouseActive {
		termStateMu.Unlock()
		return nil
	}
	termMouseActive = false
	termStateMu.Unlock()
	if _, err := fmt.Fprint(os.Stdout, mouseDisableSeq); err != nil {
		return fmt.Errorf("could not disable mouse: %w", err)
	}
	return nil
}

// termEnterAltScreen switches to the alternate screen. Tracking this
// explicitly lets TermCleanup leave the alternate screen only when it was
// actually entered; unconditional DECRST 1049 moves the cursor on some
// terminals and breaks plain non-fullscreen scripts.
func termEnterAltScreen() error {
	termStateMu.Lock()
	if termAltActive {
		termStateMu.Unlock()
		return nil
	}
	if !stdoutIsTerminal() {
		termStateMu.Unlock()
		return fmt.Errorf("stdout is not a terminal")
	}
	termAltActive = true
	termStateMu.Unlock()
	if _, err := fmt.Fprint(os.Stdout, altEnterSeq); err != nil {
		return fmt.Errorf("could not enter alternate screen: %w", err)
	}
	return nil
}

func termLeaveAltScreen() error {
	termStateMu.Lock()
	if !termAltActive {
		termStateMu.Unlock()
		return nil
	}
	termAltActive = false
	termStateMu.Unlock()
	if _, err := fmt.Fprint(os.Stdout, altLeaveSeq); err != nil {
		return fmt.Errorf("could not leave alternate screen: %w", err)
	}
	return nil
}

// termShowCursor makes the text cursor visible again. It is kept
// separate from mouse teardown so releasing the mouse mid-game does not
// make the cursor blink while the game is still rendering. Piped output
// never hid the cursor, so the guard keeps captured output clean.
func termShowCursor() error {
	if !stdoutIsTerminal() {
		return nil
	}
	if _, err := fmt.Fprint(os.Stdout, cursorShowSeq); err != nil {
		return fmt.Errorf("could not show cursor: %w", err)
	}
	return nil
}

// TermDisableMouse is the exported mouse teardown hook.
func TermDisableMouse() error {
	return termDisableMouse()
}

// TermCleanup restores every terminal mode this process may have
// changed. It is safe to call more than once and is used by main's exit
// path so runtime errors do not leave the terminal in raw/mouse mode.
func TermCleanup() {
	_ = termDisableMouse()
	_ = termLeaveAltScreen()
	if stdoutIsTerminal() {
		_, _ = fmt.Fprint(os.Stdout, terminalResetSeq)
	}
	_ = termShowCursor()
	_ = termDisableRaw()
}

// TermSignalCleanup is the signal-handler variant of TermCleanup. It
// restores raw mode first, synchronously and without writing to stdout,
// because a PTY whose output buffer is full would otherwise block the
// handler before termios is restored. The remaining best-effort screen
// writes run in a goroutine with a short deadline; the caller exits
// immediately afterwards.
func TermSignalCleanup() {
	termRestoreRawOnly()
	done := make(chan struct{})
	go func() {
		_ = termDisableMouse()
		_ = termLeaveAltScreen()
		if stdoutIsTerminal() {
			_, _ = fmt.Fprint(os.Stdout, terminalResetSeq)
		}
		_ = termShowCursor()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(100 * time.Millisecond):
	}
}

// collectEscapeSequence finishes an escape sequence whose first byte (the
// ESC itself) has already been read. tryRead must return the next byte when
// one is already available and ok=false otherwise; collectEscapeSequence
// polls it every 2ms until a sequence terminator is seen (see
// isEscapeTerminator), the idle timeout expires, the total timeout
// expires, or the byte cap is reached. A lone ESC key press therefore
// resolves to "\x1b" after escapeSeqTimeout.
func collectEscapeSequence(first byte, tryRead func() (byte, bool)) string {
	sep := []byte{first}
	start := time.Now()
	lastByte := start
	for len(sep) < maxEscapeSeqBytes && time.Since(start) < escapeSeqTotalTimeout {
		if b, ok := tryRead(); ok {
			sep = append(sep, b)
			lastByte = time.Now()
			if isEscapeTerminator(sep, b) {
				break
			}
			continue
		}
		if time.Since(lastByte) >= escapeSeqTimeout {
			break
		}
		time.Sleep(2 * time.Millisecond)
	}
	return string(sep)
}

// sgrPayloadStart returns the index of the Cb field in an SGR mouse
// report, or -1 if sep is not one. A lone ESC pressed just before a click
// can be merged by the escape assembler ("ESC ESC [ <"), so that optional
// leading ESC is recognized here too.
func sgrPayloadStart(sep []byte) int {
	if len(sep) >= 3 && sep[0] == 0x1b && sep[1] == '[' && sep[2] == '<' {
		return 3
	}
	if len(sep) >= 4 && sep[0] == 0x1b && sep[1] == 0x1b && sep[2] == '[' && sep[3] == '<' {
		return 4
	}
	return -1
}

// isSGRMousePrefix reports whether sep starts an SGR mouse report,
// optionally with one extra ESC merged in front.
func isSGRMousePrefix(sep []byte) bool {
	return sgrPayloadStart(sep) >= 0
}

// validSGRMousePayload checks the grammar of a complete SGR report:
// ESC [ < digits ; digits ; digits M/m. It deliberately accepts values
// outside the protocol's ranges; parse_mouse applies those limits.
func validSGRMousePayload(sep []byte) bool {
	start := sgrPayloadStart(sep)
	if start < 0 || len(sep)-start < 6 {
		return false
	}
	last := sep[len(sep)-1]
	if last != 'M' && last != 'm' {
		return false
	}
	digits := 0
	semis := 0
	for _, c := range sep[start : len(sep)-1] {
		if c >= '0' && c <= '9' {
			digits++
			continue
		}
		if c != ';' || digits == 0 || semis == 2 {
			return false
		}
		semis++
		digits = 0
	}
	return semis == 2 && digits > 0
}

// sgrStructurallyComplete reports whether sep ends in M/m and has at
// least three semicolon-separated fields. It does not require the fields
// to be decimal: a malformed but complete report must terminate at its
// final M/m so the bytes after it are not swallowed by the assembler.
func sgrStructurallyComplete(sep []byte) bool {
	start := sgrPayloadStart(sep)
	if start < 0 || len(sep)-start < 2 {
		return false
	}
	last := sep[len(sep)-1]
	if last != 'M' && last != 'm' {
		return false
	}
	semis := 0
	for _, c := range sep[start : len(sep)-1] {
		if c == ';' {
			semis++
		}
	}
	return semis >= 2
}

// isEscapeTerminator reports whether b completes the sequence assembled in
// sep. Most ANSI sequences end with a letter or '~'. An SGR mouse report
// ("ESC [ < ... M" or "m") is special: M/m is its terminator even when the
// preceding fields are malformed, because a malformed report cannot become
// valid after its terminator. Stopping there leaves the remaining bytes to
// the tail filter instead of consuming a following key.
func isEscapeTerminator(sep []byte, b byte) bool {
	if isSGRMousePrefix(sep) {
		return b == 'M' || b == 'm'
	}
	return (b >= 'A' && b <= 'Z') || (b >= 'a' && b <= 'z') || b == '~'
}

// beginSGRTailDiscard starts swallowing the tail of an incomplete SGR
// report. Valid reports and structurally complete malformed reports are
// exempt: their bytes are already consumed, and a following key must
// still work. Any other SGR-prefixed string may have a delayed tail
// (including one that ended in a letter or an early M/m), so discard is
// armed unconditionally for it.
func beginSGRTailDiscard(s string) {
	b := []byte(s)
	if !isSGRMousePrefix(b) || validSGRMousePayload(b) || sgrStructurallyComplete(b) {
		return
	}
	termSGRDiscard = true
	termSGRDiscardUntil = time.Now().Add(500 * time.Millisecond)
	termSGRDiscardBytes = 0
}

// filterSGRTail applies the pending truncated-report discard to one key.
// Only bytes that can legally continue an SGR report are swallowed; any
// other byte (including ESC, which may start a fresh sequence) ends the
// discard and is processed normally.
func filterSGRTail(s string) (string, bool) {
	if !termSGRDiscard {
		return s, false
	}
	if time.Now().After(termSGRDiscardUntil) || termSGRDiscardBytes >= 256 {
		termSGRDiscard = false
		return s, false
	}
	if len(s) == 1 {
		c := s[0]
		if c == 'M' || c == 'm' {
			termSGRDiscard = false
			return "", true
		}
		if (c >= '0' && c <= '9') || c == ';' || c == '<' {
			termSGRDiscardBytes++
			return "", true
		}
	}
	termSGRDiscard = false
	return s, false
}

// termReadKey wraps the platform reader with truncated-SGR filtering so
// tail bytes never reach the VM as hotkeys.
func termReadKey() (string, error) {
	for {
		s, err := termReadKeyRaw()
		if err != nil {
			return "", err
		}
		out, skip := filterSGRTail(s)
		if skip {
			// poll_events only calls __read_key when has_input was true.
			// A swallowed tail may have been the last buffered byte; do
			// not block the frame waiting for an unrelated future key.
			if ok, _ := termHasInput(); !ok {
				return "", nil
			}
			continue
		}
		beginSGRTailDiscard(out)
		return out, nil
	}
}
