package object

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// drainStream polls until done or timeout, returning the last poll hash.
func drainStream(t *testing.T, handle int64) *Hash {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	var last *Hash
	for time.Now().Before(deadline) {
		p := streamPoll(NewInteger(handle))
		h, ok := p.(*Hash)
		if !ok {
			t.Fatalf("stream_poll returned %T, want *Hash", p)
		}
		last = h
		if done, _ := lookupHashKey(h, "done"); done == TRUE {
			return h
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("stream did not finish; last poll: %v", last)
	return nil
}

func TestStreamUsageAndFinishReason(t *testing.T) {
	sse := "data: {\"choices\":[{\"delta\":{\"content\":\"hi\"},\"finish_reason\":null}]}\n\n" +
		"data: {\"choices\":[{\"delta\":{},\"finish_reason\":\"stop\"}],\"usage\":{\"prompt_tokens\":11,\"completion_tokens\":22,\"total_tokens\":33}}\n\n" +
		"data: [DONE]\n\n"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		fmt.Fprint(w, sse)
	}))
	defer srv.Close()

	h := streamStart(&String{Value: srv.URL}, &String{Value: "{}"}, &Hash{}, NewInteger(10000))
	handle, ok := h.(*Integer)
	if !ok {
		t.Fatalf("stream_start returned %T (%v), want *Integer", h, h)
	}

	poll := drainStream(t, handle.Value)
	defer streamClose(NewInteger(handle.Value))

	text, _ := lookupHashKey(poll, "text")
	if s, ok := text.(*String); !ok || s.Value != "hi" {
		t.Fatalf("text = %v, want \"hi\"", text)
	}

	fr, _ := lookupHashKey(poll, "finish_reason")
	if s, ok := fr.(*String); !ok || s.Value != "stop" {
		t.Fatalf("finish_reason = %v, want \"stop\"", fr)
	}

	usageObj, _ := lookupHashKey(poll, "usage")
	usage, ok := usageObj.(*Hash)
	if !ok {
		t.Fatalf("usage = %v, want *Hash", usageObj)
	}
	for key, want := range map[string]int64{"prompt_tokens": 11, "completion_tokens": 22, "total_tokens": 33} {
		v, _ := lookupHashKey(usage, key)
		n, ok := v.(*Integer)
		if !ok || n.Value != want {
			t.Fatalf("usage[%s] = %v, want %d", key, v, want)
		}
	}
}

func TestStreamWithoutUsage(t *testing.T) {
	sse := "data: {\"choices\":[{\"delta\":{\"content\":\"x\"},\"finish_reason\":\"length\"}]}\n\n" +
		"data: [DONE]\n\n"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		fmt.Fprint(w, sse)
	}))
	defer srv.Close()

	h := streamStart(&String{Value: srv.URL}, &String{Value: "{}"}, &Hash{}, NewInteger(10000))
	handle, ok := h.(*Integer)
	if !ok {
		t.Fatalf("stream_start returned %T (%v), want *Integer", h, h)
	}

	poll := drainStream(t, handle.Value)
	defer streamClose(NewInteger(handle.Value))

	usageObj, _ := lookupHashKey(poll, "usage")
	if usageObj != NULL {
		t.Fatalf("usage = %v, want NULL", usageObj)
	}
	fr, _ := lookupHashKey(poll, "finish_reason")
	if s, ok := fr.(*String); !ok || s.Value != "length" {
		t.Fatalf("finish_reason = %v, want \"length\"", fr)
	}
}
