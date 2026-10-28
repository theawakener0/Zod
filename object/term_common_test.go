package object

import (
	"os"
	"strings"
	"testing"
	"time"
)

// collectWithString feeds seq[1:] through the escape-sequence assembler
// one byte at a time, mimicking a terminal that has already delivered the
// leading ESC.
func collectWithString(seq string) string {
	i := 1
	return collectEscapeSequence(seq[0], func() (byte, bool) {
		if i >= len(seq) {
			return 0, false
		}
		b := seq[i]
		i++
		return b, true
	})
}

func TestCollectEscapeSequenceSGRMouse(t *testing.T) {
	cases := []string{
		"\x1b[<0;10;5M",     // left press
		"\x1b[<0;120;40m",   // release, large coordinates
		"\x1b[<35;80;30M",   // motion without a button
		"\x1b[<64;1;1M",     // wheel up
		"\x1b\x1b[<0;10;5M", // ESC merged with a click
	}
	for _, seq := range cases {
		if got := collectWithString(seq); got != seq {
			t.Fatalf("collectEscapeSequence(%q) = %q, want the full report", seq, got)
		}
	}
}

func TestCollectEscapeSequenceArrowKey(t *testing.T) {
	const seq = "\x1b[A"
	if got := collectWithString(seq); got != seq {
		t.Fatalf("collectEscapeSequence(%q) = %q", seq, got)
	}
}

func TestCollectEscapeSequenceCap(t *testing.T) {
	seq := "\x1b[" + strings.Repeat("1", 40)
	got := collectWithString(seq)
	if len(got) != maxEscapeSeqBytes {
		t.Fatalf("assembled %d bytes, want cap %d", len(got), maxEscapeSeqBytes)
	}
	if !strings.HasPrefix(got, "\x1b[") {
		t.Fatalf("assembled sequence lost its prefix: %q", got)
	}
}

// A malformed SGR report containing letters must still be consumed as one
// unit; ending the escape sequence at the first letter would leak the rest
// of the report into the key queue (and could trigger game hotkeys).
func TestCollectEscapeSequenceMalformedSGR(t *testing.T) {
	const seq = "\x1b[<0;abc;5M"
	if got := collectWithString(seq); got != seq {
		t.Fatalf("collectEscapeSequence(%q) = %q, want the whole report", seq, got)
	}
}

// M/m is the SGR terminator even when the preceding fields are
// malformed. Stopping there leaves the rest of the bytes to the tail
// filter, which preserves a following ordinary key.
func TestCollectEscapeSequenceMalformedSGRTerminators(t *testing.T) {
	cases := []struct {
		seq  string
		want string
	}{
		{"\x1b[<0;M;5M", "\x1b[<0;M"},
		{"\x1b[<0;1m;2M", "\x1b[<0;1m"},
		// No M/m: the assembler can only stop at the idle timeout and
		// returns everything it collected.
		{"\x1b[<0;10;5X", "\x1b[<0;10;5X"},
	}
	for _, tc := range cases {
		if got := collectWithString(tc.seq); got != tc.want {
			t.Fatalf("collectEscapeSequence(%q) = %q, want %q", tc.seq, got, tc.want)
		}
	}
}

// A truncated SGR report at the byte cap must stay one unit; termReadKey
// is responsible for discarding it (and its delayed tail) without
// emitting keys.
func TestCollectEscapeSequenceSGRCap(t *testing.T) {
	seq := "\x1b[<" + strings.Repeat("1", 40)
	got := collectWithString(seq)
	if len(got) != maxEscapeSeqBytes {
		t.Fatalf("assembled %d bytes, want cap %d", len(got), maxEscapeSeqBytes)
	}
	if !strings.HasPrefix(got, "\x1b[<") {
		t.Fatalf("assembled sequence lost its prefix: %q", got)
	}
}

func TestValidSGRMousePayload(t *testing.T) {
	valid := []string{
		"\x1b[<0;1;1M",
		"\x1b[<64;120;40m",
		"\x1b[<255;9999;9999M",
		"\x1b\x1b[<0;10;5M",
	}
	for _, seq := range valid {
		if !validSGRMousePayload([]byte(seq)) {
			t.Fatalf("validSGRMousePayload(%q) = false, want true", seq)
		}
	}
	invalid := []string{
		"\x1b[<0;1;1",
		"\x1b[<0;1M",
		"\x1b[<;1;1M",
		"\x1b[<0;;1M",
		"\x1b[<0;1;M",
		"\x1b[<0;1;1;2M",
		"\x1b[<0;1;1X",
		"\x1b[<0;M;1M",
		"\x1b[<0;+1;1M",
		"\x1b[<0;1 ;1M",
	}
	for _, seq := range invalid {
		if validSGRMousePayload([]byte(seq)) {
			t.Fatalf("validSGRMousePayload(%q) = true, want false", seq)
		}
	}
}

func resetSGRDiscard() {
	termSGRDiscard = false
	termSGRDiscardUntil = time.Time{}
	termSGRDiscardBytes = 0
}

// A truncated report must swallow its delayed tail, but a legitimate key
// must end the discard immediately instead of disappearing.
func TestSGRTailDiscardKey(t *testing.T) {
	resetSGRDiscard()
	beginSGRTailDiscard("\x1b[<0;10;5")
	if !termSGRDiscard {
		t.Fatal("truncated report did not start tail discard")
	}
	if out, skip := filterSGRTail("q"); skip || out != "q" {
		t.Fatalf("filterSGRTail(q) = %q, %v; want q, false", out, skip)
	}
	if termSGRDiscard {
		t.Fatal("discard still active after a normal key")
	}
}

func TestSGRTailDiscardTerminator(t *testing.T) {
	resetSGRDiscard()
	beginSGRTailDiscard("\x1b[<0;10;5")
	for _, s := range []string{"6", ";", "7"} {
		if out, skip := filterSGRTail(s); !skip || out != "" {
			t.Fatalf("filterSGRTail(%q) = %q, %v; want swallow", s, out, skip)
		}
	}
	if out, skip := filterSGRTail("M"); !skip || out != "" {
		t.Fatalf("filterSGRTail(M) = %q, %v; want swallow", out, skip)
	}
	if termSGRDiscard {
		t.Fatal("discard still active after terminator")
	}
}

func TestSGRTailDiscardFreshSequence(t *testing.T) {
	resetSGRDiscard()
	beginSGRTailDiscard("\x1b[<0;10;5")
	const seq = "\x1b[<0;1;1M"
	if out, skip := filterSGRTail(seq); skip || out != seq {
		t.Fatalf("filterSGRTail(fresh SGR) = %q, %v; want the sequence", out, skip)
	}
	if termSGRDiscard {
		t.Fatal("discard still active after a fresh sequence")
	}
}

// A complete malformed report must not start a discard: its bytes are
// already consumed, and a following key has to work normally.
func TestSGRTailDiscardCompleteMalformed(t *testing.T) {
	for _, seq := range []string{
		"\x1b[<0;abc;5M",
		"\x1b[<0;M;5M",
		"\x1b[<0;1m;2M",
	} {
		resetSGRDiscard()
		beginSGRTailDiscard(seq)
		if termSGRDiscard {
			t.Fatalf("complete malformed report %q started a discard", seq)
		}
	}
}

// An M/m before the third field is an early terminator, not a complete
// report: the remaining fields must be swallowed.
// A malformed report that ends in a letter (no M/m yet) may still have a
// delayed numeric tail. Arm the discard even though the last byte is not
// a digit, semicolon, or angle bracket.
func TestSGRTailDiscardLetterTail(t *testing.T) {
	resetSGRDiscard()
	beginSGRTailDiscard("\x1b[<0;abc")
	if !termSGRDiscard {
		t.Fatal("letter-ended malformed report did not start tail discard")
	}
	for _, s := range []string{";", "5", "M"} {
		if out, skip := filterSGRTail(s); !skip || out != "" {
			t.Fatalf("tail byte %q = %q, %v; want swallow", s, out, skip)
		}
	}
	if out, skip := filterSGRTail("q"); skip || out != "q" {
		t.Fatalf("filterSGRTail(q) = %q, %v; want q, false", out, skip)
	}
}

// A swallowed tail may be the last buffered byte. termReadKey must
// return "no key" instead of blocking for an unrelated future key;
// poll_events depends on this to keep rendering after a malformed report.
func TestTermReadKeyTailDoesNotBlock(t *testing.T) {
	resetSGRDiscard()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	oldStdin := os.Stdin
	os.Stdin = r
	defer func() {
		os.Stdin = oldStdin
		_ = r.Close()
		_ = w.Close()
		resetSGRDiscard()
	}()

	if _, err := w.WriteString("\x1b[<0;abc"); err != nil {
		t.Fatal(err)
	}
	got, err := termReadKey()
	if err != nil {
		t.Fatal(err)
	}
	if got != "\x1b[<0;abc" {
		t.Fatalf("termReadKey() = %q, want the truncated report", got)
	}
	if !termSGRDiscard {
		t.Fatal("truncated report did not arm the tail discard")
	}

	if _, err := w.WriteString(";5M"); err != nil {
		t.Fatal(err)
	}
	type result struct {
		s   string
		err error
	}
	done := make(chan result, 1)
	go func() {
		s, err := termReadKey()
		done <- result{s, err}
	}()
	select {
	case res := <-done:
		if res.err != nil {
			t.Fatal(res.err)
		}
		if res.s != "" {
			t.Fatalf("termReadKey() = %q, want an empty no-key result", res.s)
		}
	case <-time.After(time.Second):
		t.Fatal("termReadKey blocked after consuming the last tail byte")
	}

	if _, err := w.WriteString("q"); err != nil {
		t.Fatal(err)
	}
	got, err = termReadKey()
	if err != nil {
		t.Fatal(err)
	}
	if got != "q" {
		t.Fatalf("termReadKey() = %q, want q after the tail", got)
	}
}

func TestSGRTailDiscardEarlyTerminator(t *testing.T) {
	resetSGRDiscard()
	beginSGRTailDiscard("\x1b[<0;123M")
	if !termSGRDiscard {
		t.Fatal("early M did not start tail discard")
	}
	for _, s := range []string{";", "5", "M"} {
		if _, skip := filterSGRTail(s); !skip {
			t.Fatalf("early-tail byte %q was not swallowed", s)
		}
	}
	if out, skip := filterSGRTail("w"); skip || out != "w" {
		t.Fatalf("filterSGRTail(w) = %q, %v; want w, false", out, skip)
	}
}

func TestSGRTailDiscardCap(t *testing.T) {
	resetSGRDiscard()
	seq := "\x1b[<" + strings.Repeat("1", maxEscapeSeqBytes-3)
	beginSGRTailDiscard(seq)
	if !termSGRDiscard {
		t.Fatal("cap-truncated report did not start tail discard")
	}
	resetSGRDiscard()
}

func TestSGRTailDiscardExpiry(t *testing.T) {
	resetSGRDiscard()
	beginSGRTailDiscard("\x1b[<0;10;5")
	termSGRDiscardUntil = time.Now().Add(-time.Millisecond)
	if out, skip := filterSGRTail("5"); skip || out != "5" {
		t.Fatalf("expired discard swallowed a key: %q, %v", out, skip)
	}
}

// A terminal may deliver an SGR report across several reads with gaps
// longer than the idle timeout, as long as no single gap exceeds it. The
// old single-deadline implementation truncated this case.
func TestCollectEscapeSequenceSlowSplitReport(t *testing.T) {
	const seq = "\x1b[<0;1;1M"
	start := time.Now()
	i := 1
	got := collectEscapeSequence(seq[0], func() (byte, bool) {
		if i >= len(seq) {
			return 0, false
		}
		if time.Since(start) < time.Duration(i)*30*time.Millisecond {
			return 0, false
		}
		b := seq[i]
		i++
		return b, true
	})
	if got != seq {
		t.Fatalf("slow split report = %q, want %q", got, seq)
	}
}
