package object

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// drainStreamAll polls until done, accumulating the text delivered across
// polls and returning the concatenation plus the final poll hash.
func drainStreamAll(t *testing.T, handle int64) (string, *Hash) {
	t.Helper()
	deadline := time.Now().Add(6 * time.Second)
	var all strings.Builder
	var last *Hash
	for time.Now().Before(deadline) {
		p := streamPoll(NewInteger(handle))
		h, ok := p.(*Hash)
		if !ok {
			t.Fatalf("stream_poll returned %T, want *Hash", p)
		}
		last = h
		if text, _ := lookupHashKey(h, "text"); text != nil {
			if s, ok := text.(*String); ok {
				all.WriteString(s.Value)
			}
		}
		if done, _ := lookupHashKey(h, "done"); done == TRUE {
			return all.String(), h
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("stream did not finish; last poll: %v", last)
	return "", nil
}

// The stream emits a chunk every 120ms for ~1.2s while the idle timeout is
// 400ms: every chunk must reset the inactivity watchdog so the stream
// completes without error.
func TestStreamIdleTimeoutResetsOnData(t *testing.T) {
	const chunks = 10
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		flusher, _ := w.(http.Flusher)
		for i := 0; i < chunks; i++ {
			fmt.Fprintf(w, "data: {\"choices\":[{\"delta\":{\"content\":\"c%d\"}}]}\n\n", i)
			flusher.Flush()
			time.Sleep(120 * time.Millisecond)
		}
		fmt.Fprint(w, "data: [DONE]\n\n")
		flusher.Flush()
	}))
	defer srv.Close()

	start := time.Now()
	h := streamStart(&String{Value: srv.URL}, &String{Value: "{}"}, &Hash{}, NewInteger(400))
	handle, ok := h.(*Integer)
	if !ok {
		t.Fatalf("stream_start returned %T (%v), want *Integer", h, h)
	}
	text, poll := drainStreamAll(t, handle.Value)
	defer streamClose(NewInteger(handle.Value))
	elapsed := time.Since(start)

	// Without idle-timeout resets the stream would be killed after ~400ms.
	if elapsed < 800*time.Millisecond {
		t.Fatalf("stream finished in %v, want it to outlive the 400ms idle window", elapsed)
	}

	errObj, _ := lookupHashKey(poll, "error")
	if errObj != NULL {
		t.Fatalf("error = %v, want NULL", errObj)
	}
	var want strings.Builder
	for i := 0; i < chunks; i++ {
		fmt.Fprintf(&want, "c%d", i)
	}
	if text != want.String() {
		t.Fatalf("text = %q, want %q", text, want.String())
	}
}

// One chunk is delivered, then the server goes silent for 2s while the idle
// timeout is 300ms: the watchdog must fire with the inactivity error and the
// chunk delivered before the stall must remain visible.
func TestStreamIdleTimeoutFires(t *testing.T) {
	const first = "first-chunk"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		flusher, _ := w.(http.Flusher)
		fmt.Fprintf(w, "data: {\"choices\":[{\"delta\":{\"content\":%q}}]}\n\n", first)
		flusher.Flush()
		select {
		case <-time.After(2 * time.Second):
		case <-r.Context().Done():
			return
		}
		fmt.Fprint(w, "data: {\"choices\":[{\"delta\":{\"content\":\"late\"}}]}\n\n")
		flusher.Flush()
	}))
	defer srv.Close()

	h := streamStart(&String{Value: srv.URL}, &String{Value: "{}"}, &Hash{}, NewInteger(300))
	handle, ok := h.(*Integer)
	if !ok {
		t.Fatalf("stream_start returned %T (%v), want *Integer", h, h)
	}
	text, poll := drainStreamAll(t, handle.Value)
	defer streamClose(NewInteger(handle.Value))

	errObj, _ := lookupHashKey(poll, "error")
	s, ok := errObj.(*String)
	if !ok {
		t.Fatalf("error = %v (%T), want non-NULL *String", errObj, errObj)
	}
	if !strings.Contains(s.Value, "timed out") {
		t.Fatalf("error = %q, want it to mention \"timed out\"", s.Value)
	}
	if !strings.Contains(s.Value, "300ms of inactivity") {
		t.Fatalf("error = %q, want it to mention \"300ms of inactivity\"", s.Value)
	}
	if text != first {
		t.Fatalf("text = %q, want %q delivered before the stall", text, first)
	}
}

// A manual streamClose must finish the stream without surfacing an error.
// streamClose removes the handle from the registry, so the returned handle
// can no longer be polled; observe the stream state directly instead.
func TestStreamCloseStaysSilent(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.(http.Flusher).Flush()
		<-r.Context().Done()
	}))
	defer srv.Close()

	h := streamStart(&String{Value: srv.URL}, &String{Value: "{}"}, &Hash{}, NewInteger(10000))
	handle, ok := h.(*Integer)
	if !ok {
		t.Fatalf("stream_start returned %T (%v), want *Integer", h, h)
	}
	st, ok := sseLookup(handle.Value)
	if !ok {
		t.Fatalf("stream handle %d not in registry", handle.Value)
	}
	if obj := streamClose(NewInteger(handle.Value)); obj != NULL {
		t.Fatalf("stream_close = %v, want NULL", obj)
	}

	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		st.mu.Lock()
		done, errMsg := st.done, st.errMsg
		st.mu.Unlock()
		if done {
			if errMsg != "" {
				t.Fatalf("manual close surfaced error %q, want none", errMsg)
			}
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatal("stream did not finish after manual close")
}
