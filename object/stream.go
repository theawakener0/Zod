package object

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

// sseStream holds the in-flight state of one concurrent SSE stream.
// The fetching goroutine appends to textPending/toolPending; script code
// drains them via __stream_poll. All fields are guarded by mu.
type sseStream struct {
	mu          sync.Mutex
	textPending strings.Builder
	toolPending []Object
	done        bool
	errMsg      string
	cancel      context.CancelFunc
	manual      bool
}

// sseRegistry maps integer handles to live streams.
var sseRegistry = struct {
	sync.Mutex
	next int64
	live map[int64]*sseStream
}{next: 1, live: make(map[int64]*sseStream)}

func sseLookup(id int64) (*sseStream, bool) {
	sseRegistry.Lock()
	defer sseRegistry.Unlock()
	st, ok := sseRegistry.live[id]
	return st, ok
}

func makeToolCallHash(index int64, id, name, args string) *Hash {
	pairs := make(map[HashKey]HashPair, 4)
	order := make([]HashKey, 0, 4)
	add := func(k string, v Object) {
		ko := &String{Value: k}
		hk := ko.HashKey()
		pairs[hk] = HashPair{Key: ko, Value: v}
		order = append(order, hk)
	}
	add("index", NewInteger(index))
	if id == "" {
		add("id", NULL)
	} else {
		add("id", &String{Value: id})
	}
	if name == "" {
		add("name", NULL)
	} else {
		add("name", &String{Value: name})
	}
	if args == "" {
		add("args", NULL)
	} else {
		add("args", &String{Value: args})
	}
	return &Hash{Pairs: pairs, Order: order}
}

func makePollHash(text string, tools []Object, done bool, errMsg string) *Hash {
	pairs := make(map[HashKey]HashPair, 4)
	order := make([]HashKey, 0, 4)
	add := func(k string, v Object) {
		ko := &String{Value: k}
		hk := ko.HashKey()
		pairs[hk] = HashPair{Key: ko, Value: v}
		order = append(order, hk)
	}
	if tools == nil {
		tools = []Object{}
	}
	var doneObj Object = FALSE
	if done {
		doneObj = TRUE
	}
	var errObj Object = NULL
	if errMsg != "" {
		errObj = &String{Value: errMsg}
	}
	add("text", &String{Value: text})
	add("tools", &Array{Elements: tools})
	add("done", doneObj)
	add("error", errObj)
	return &Hash{Pairs: pairs, Order: order}
}

type sseToolCall struct {
	Index int    `json:"index"`
	ID    string `json:"id"`
	Type  string `json:"type"`
	Func  struct {
		Name      string `json:"name"`
		Arguments string `json:"arguments"`
	} `json:"function"`
}

type sseChunk struct {
	Choices []struct {
		Delta struct {
			Content   string        `json:"content"`
			ToolCalls []sseToolCall `json:"tool_calls"`
		} `json:"delta"`
	} `json:"choices"`
}

func runSSEStream(st *sseStream, client *http.Client, req *http.Request, url string, ctx context.Context) {
	finishErr := func(msg string) {
		st.mu.Lock()
		st.errMsg = msg
		st.done = true
		st.mu.Unlock()
	}
	finishDone := func() {
		st.mu.Lock()
		st.done = true
		st.mu.Unlock()
	}
	isManual := func() bool {
		st.mu.Lock()
		defer st.mu.Unlock()
		return st.manual
	}

	resp, err := client.Do(req)
	if err != nil {
		if isManual() {
			finishDone()
			return
		}
		if ctx.Err() == context.DeadlineExceeded {
			finishErr("http stream timed out")
			return
		}
		finishErr(fmt.Sprintf("http stream %q failed: %s", url, err.Error()))
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		data, _ := io.ReadAll(io.LimitReader(resp.Body, maxHTTPBody+1))
		finishErr(fmt.Sprintf("http stream %q failed with status %s: %s", url, resp.Status, string(data)))
		return
	}

	scanner := bufio.NewScanner(resp.Body)
	var totalBytes int64
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" || strings.HasPrefix(line, ":") {
			continue
		}
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		payload := strings.TrimSpace(strings.TrimPrefix(line, "data: "))
		if payload == "[DONE]" {
			break
		}
		var chunk sseChunk
		if err := json.Unmarshal([]byte(payload), &chunk); err != nil {
			continue
		}
		st.mu.Lock()
		overflow := false
		for _, ch := range chunk.Choices {
			if ch.Delta.Content != "" {
				if totalBytes+int64(len(ch.Delta.Content)) > maxHTTPBody {
					overflow = true
					break
				}
				totalBytes += int64(len(ch.Delta.Content))
				st.textPending.WriteString(ch.Delta.Content)
			}
			for _, tc := range ch.Delta.ToolCalls {
				if totalBytes+int64(len(tc.Func.Arguments)) > maxHTTPBody {
					overflow = true
					break
				}
				totalBytes += int64(len(tc.Func.Arguments))
				st.toolPending = append(st.toolPending, makeToolCallHash(int64(tc.Index), tc.ID, tc.Func.Name, tc.Func.Arguments))
			}
			if overflow {
				break
			}
		}
		if overflow {
			st.errMsg = "http stream response too large"
			st.done = true
			st.mu.Unlock()
			st.cancel()
			return
		}
		st.mu.Unlock()
	}

	if errScan := scanner.Err(); errScan != nil {
		if isManual() {
			finishDone()
			return
		}
		if ctx.Err() == context.DeadlineExceeded {
			finishErr("http stream timed out")
			return
		}
		finishErr(fmt.Sprintf("http stream %q read error: %s", url, errScan.Error()))
		return
	}
	finishDone()
}

func streamStart(args ...Object) Object {
	if len(args) != 4 {
		return newError("wrong number of arguments. got=%d, want=4", len(args))
	}
	u, ok := args[0].(*String)
	if !ok {
		return newError("first argument to `__http_post_stream` must be STRING. got=%s", args[0].Type())
	}
	body, ok := args[1].(*String)
	if !ok {
		return newError("second argument to `__http_post_stream` must be STRING. got=%s", args[1].Type())
	}
	headersObj, ok := args[2].(*Hash)
	if !ok {
		return newError("third argument to `__http_post_stream` must be HASH. got=%s", args[2].Type())
	}
	ms, ok := args[3].(*Integer)
	if !ok {
		return newError("fourth argument to `__http_post_stream` must be INTEGER. got=%s", args[3].Type())
	}
	if ms.Value < 1 || ms.Value > 300000 {
		return newError("fourth argument to `__http_post_stream` must be between 1 and 300000. got=%d", ms.Value)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(ms.Value)*time.Millisecond)
	req, err := http.NewRequestWithContext(ctx, "POST", u.Value, strings.NewReader(body.Value))
	if err != nil {
		cancel()
		return newError("http stream %q failed: %s", u.Value, err.Error())
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "text/event-stream")
	if errObj := applyHeaders(headersObj, req.Header); errObj != NULL {
		cancel()
		return errObj
	}

	st := &sseStream{cancel: cancel}
	sseRegistry.Lock()
	id := sseRegistry.next
	sseRegistry.next++
	sseRegistry.live[id] = st
	sseRegistry.Unlock()

	client := &http.Client{}
	go runSSEStream(st, client, req, u.Value, ctx)

	return NewInteger(id)
}

func streamPoll(args ...Object) Object {
	if len(args) != 1 {
		return newError("wrong number of arguments. got=%d, want=1", len(args))
	}
	h, ok := args[0].(*Integer)
	if !ok {
		return newError("argument to `__stream_poll` must be INTEGER. got=%s", args[0].Type())
	}
	st, ok := sseLookup(h.Value)
	if !ok {
		return newError("unknown stream handle: %d", h.Value)
	}
	st.mu.Lock()
	text := st.textPending.String()
	st.textPending.Reset()
	tools := st.toolPending
	st.toolPending = nil
	done := st.done
	errMsg := st.errMsg
	st.mu.Unlock()
	return makePollHash(text, tools, done, errMsg)
}

func streamCancel(args ...Object) Object {
	if len(args) != 1 {
		return newError("wrong number of arguments. got=%d, want=1", len(args))
	}
	h, ok := args[0].(*Integer)
	if !ok {
		return newError("argument to `__stream_cancel` must be INTEGER. got=%s", args[0].Type())
	}
	st, ok := sseLookup(h.Value)
	if !ok {
		return newError("unknown stream handle: %d", h.Value)
	}
	st.mu.Lock()
	st.manual = true
	st.mu.Unlock()
	st.cancel()
	return NULL
}

func streamClose(args ...Object) Object {
	if len(args) != 1 {
		return newError("wrong number of arguments. got=%d, want=1", len(args))
	}
	h, ok := args[0].(*Integer)
	if !ok {
		return newError("argument to `__stream_close` must be INTEGER. got=%s", args[0].Type())
	}
	st, ok := sseLookup(h.Value)
	if !ok {
		return newError("unknown stream handle: %d", h.Value)
	}
	st.mu.Lock()
	st.manual = true
	st.mu.Unlock()
	st.cancel()
	sseRegistry.Lock()
	delete(sseRegistry.live, h.Value)
	sseRegistry.Unlock()
	return NULL
}
