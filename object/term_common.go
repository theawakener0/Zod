package object

import "time"

// escapeSeqTimeout bounds how long termReadKey waits for the bytes that
// follow an ESC (0x1b) when assembling an ANSI escape sequence such as
// "\x1b[A". A terminal may deliver the bytes of a single key press in
// separate reads; without this wait a split arrow key would be reported
// as a lone ESC, which std/term's key_name maps to "esc" (the demo games
// quit on it). Matches the ~100ms-per-byte wait of the original
// implementation (VMIN=0/VTIME=1).
const escapeSeqTimeout = 100 * time.Millisecond

// maxEscapeSeqBytes caps an assembled sequence (ESC plus up to 7 more bytes).
const maxEscapeSeqBytes = 8

// collectEscapeSequence finishes an escape sequence whose first byte (the
// ESC itself) has already been read. tryRead must return the next byte when
// one is already available and ok=false otherwise; collectEscapeSequence
// polls it every 2ms until a sequence terminator (A-Z, a-z or '~') is seen,
// the deadline expires, or the byte cap is reached. A lone ESC key press
// therefore resolves to "\x1b" after escapeSeqTimeout.
func collectEscapeSequence(first byte, tryRead func() (byte, bool)) string {
	sep := []byte{first}
	deadline := time.Now().Add(escapeSeqTimeout)
	for time.Now().Before(deadline) && len(sep) < maxEscapeSeqBytes {
		if b, ok := tryRead(); ok {
			sep = append(sep, b)
			if (b >= 'A' && b <= 'Z') || (b >= 'a' && b <= 'z') || b == '~' {
				break
			}
			continue
		}
		time.Sleep(2 * time.Millisecond)
	}
	return string(sep)
}
