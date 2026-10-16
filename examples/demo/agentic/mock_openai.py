#!/usr/bin/env python3
"""Mock OpenAI-compatible server for Zodex tests (stdlib only).

Scenarios are triggered by markers in the last user message:
  __mock_tool__          -> emits a read_file tool call (first request);
                            once a tool message exists, streams "All done."
  __mock_retry__         -> first request: 429; later: normal echo
  __mock_reject_usage__  -> requests carrying stream_options get a 400;
                            caller must retry without it
  __mock_stall__         -> first request streams two chunks, stalls ~2.5s
                            past the client idle timeout, then finishes;
                            continuations echo normally
  __mock_slow__          -> streams chunks slowly (for cancel tests)
  anything else          -> streams "echo: <last user content>"

usage is attached to the final chunk only when stream_options.include_usage
was requested (like real providers). Every request is appended as one JSON
line to $MOCK_LOG (default /tmp/opencode/mock_requests.jsonl).
"""
import json
import os
import sys
import time
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer

LOG = os.environ.get("MOCK_LOG", "/tmp/opencode/mock_requests.jsonl")
RETRY_COUNTER = {"n": 0}


def log_req(record):
    with open(LOG, "a", encoding="utf-8") as f:
        f.write(json.dumps(record) + "\n")


def sse_line(payload):
    return ("data: " + json.dumps(payload) + "\n\n").encode("utf-8")


def usage_chunk(prompt, completion):
    return {
        "choices": [],
        "usage": {
            "prompt_tokens": prompt,
            "completion_tokens": completion,
            "total_tokens": prompt + completion,
        },
    }


class Handler(BaseHTTPRequestHandler):
    protocol_version = "HTTP/1.1"

    def log_message(self, fmt, *args):
        pass

    def _read_json(self):
        length = int(self.headers.get("Content-Length", "0"))
        raw = self.rfile.read(length) if length else b"{}"
        try:
            return json.loads(raw.decode("utf-8")), raw.decode("utf-8", "replace")
        except Exception:
            return {}, raw.decode("utf-8", "replace")

    def _json_response(self, code, obj):
        body = json.dumps(obj).encode("utf-8")
        self.send_response(code)
        self.send_header("Content-Type", "application/json")
        self.send_header("Content-Length", str(len(body)))
        self.end_headers()
        self.wfile.write(body)

    def do_GET(self):
        if self.path.endswith("/models"):
            log_req({"method": "GET", "path": self.path})
            self._json_response(200, {
                "data": [
                    {"id": "mock-1", "name": "Mock One", "context_length": 8192,
                     "pricing": {"prompt": "0.00001", "completion": "0.00002"}},
                    {"id": "mock-2", "name": "Mock Two", "context_length": 4096,
                     "pricing": {"prompt": "0.00002", "completion": "0.00004"}},
                ]
            })
            return
        self._json_response(404, {"error": {"message": "not found"}})

    def do_POST(self):
        body, raw = self._read_json()
        body_text = raw
        messages = body.get("messages", []) or []
        log_req({"method": "POST", "path": self.path, "body": body})
        last_user = ""
        for m in messages:
            if m.get("role") == "user" and isinstance(m.get("content"), str):
                last_user = m["content"]
        has_tool_msg = any(m.get("role") == "tool" for m in messages)
        want_usage = bool(body.get("stream_options", {}).get("include_usage"))
        stream = bool(body.get("stream"))

        if "__mock_reject_usage__" in last_user and "stream_options" in body_text:
            self._json_response(400, {"error": {"message": "unknown parameter: stream_options"}})
            return

        if "__mock_retry__" in last_user:
            RETRY_COUNTER["n"] += 1
            if RETRY_COUNTER["n"] == 1:
                self._json_response(429, {"error": {"message": "slow down (429)"}})
                return

        if not stream:
            self._json_response(200, {
                "choices": [{"message": {"role": "assistant", "content": "non-stream reply"}, "finish_reason": "stop"}],
                "usage": {"prompt_tokens": 5, "completion_tokens": 3, "total_tokens": 8},
            })
            return

        self.send_response(200)
        self.send_header("Content-Type", "text/event-stream")
        self.send_header("Cache-Control", "no-cache")
        self.send_header("Connection", "keep-alive")
        self.end_headers()

        try:
            has_assistant = any(m.get("role") == "assistant" for m in messages)
            if "__mock_length__" in last_user and not has_assistant:
                self.wfile.write(sse_line({"choices": [{"delta": {"content": "partial "}, "finish_reason": None}]}))
                self.wfile.write(sse_line({"choices": [{"delta": {}, "finish_reason": "length"}]}))
            elif "__mock_stall__" in last_user and not has_assistant:
                self.wfile.write(sse_line({"choices": [{"delta": {"content": "partial-one "}, "finish_reason": None}]}))
                self.wfile.flush()
                self.wfile.write(sse_line({"choices": [{"delta": {"content": "partial-two "}, "finish_reason": None}]}))
                self.wfile.flush()
                time.sleep(2.5)
                self.wfile.write(sse_line({"choices": [{"delta": {"content": "stalled-tail "}, "finish_reason": None}]}))
                self.wfile.write(sse_line({"choices": [{"delta": {}, "finish_reason": "stop"}]}))
            elif "__mock_tool__" in last_user and not has_tool_msg:
                self.wfile.write(sse_line({"choices": [{"delta": {"tool_calls": [{
                    "index": 0, "id": "call_1", "type": "function",
                    "function": {"name": "read_file", "arguments": "{\"path\": \""}}]},
                    "finish_reason": None}]}))
                time.sleep(0.05)
                self.wfile.write(sse_line({"choices": [{"delta": {"tool_calls": [{
                    "index": 0, "function": {"arguments": "a.txt\"}"}}]},
                    "finish_reason": None}]}))
                self.wfile.write(sse_line({"choices": [{"delta": {}, "finish_reason": "tool_calls"}]}))
            elif "__mock_slow__" in last_user:
                words = ("slow " * 40).split()
                for w in words:
                    self.wfile.write(sse_line({"choices": [{"delta": {"content": w + " "}, "finish_reason": None}]}))
                    self.wfile.flush()
                    time.sleep(0.12)
                self.wfile.write(sse_line({"choices": [{"delta": {}, "finish_reason": "stop"}]}))
            else:
                text = "echo: " + last_user
                if has_tool_msg:
                    text = "All done. [tools present]"
                for i in range(0, len(text), 128):
                    self.wfile.write(sse_line({"choices": [{"delta": {"content": text[i:i + 128]}, "finish_reason": None}]}))
                self.wfile.write(sse_line({"choices": [{"delta": {}, "finish_reason": "stop"}]}))
            if want_usage:
                self.wfile.write(sse_line(usage_chunk(21, 34)))
            self.wfile.write(b"data: [DONE]\n\n")
            self.wfile.flush()
        except (BrokenPipeError, ConnectionResetError):
            pass


if __name__ == "__main__":
    port = int(sys.argv[1]) if len(sys.argv) > 1 else 8791
    if os.path.exists(LOG):
        os.remove(LOG)
    srv = ThreadingHTTPServer(("127.0.0.1", port), Handler)
    print("mock listening on %d" % port, flush=True)
    srv.serve_forever()
