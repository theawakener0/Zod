# Zodex

A Codex-style coding agent written in ZodLang: full-screen TUI, streaming
turn engine, tool calling with `apply_patch`, approvals and sandbox policy,
rollout sessions with resume, an exec mode with JSONL events, and self-tests.
It runs on the stock Zod interpreter (one tiny interpreter change is required,
see [Development notes](#development-notes)).

```
examples/demo/harness.zd       4-line runner (imports agentic/harness_lib.zd)
examples/demo/agentic/         all Zodex modules + tests
```

## Quickstart

```sh
go build -o zod .                       # build the interpreter (repo root)
export HACKCLUB_AI_API_KEY=...          # or ZODEX_API_KEY / OPENAI_API_KEY
./zod examples/demo/harness.zd          # TUI on a terminal, line mode otherwise
```

The default provider is the HackClub AI proxy
(`https://ai.hackclub.com/proxy/v1`); any OpenAI-compatible endpoint works
(set `ZODEX_BASE_URL` or `base_url` in the config).

### Modes

```sh
./zod examples/demo/harness.zd                      # interactive (TUI or line mode)
./zod examples/demo/harness.zd -p "fix the tests"   # exec: one prompt, stream to stdout
./zod examples/demo/harness.zd exec "fix the tests" # same
./zod examples/demo/harness.zd exec "..." --json    # JSONL events (tool.start, tool.result, text, turn.completed)
./zod examples/demo/harness.zd --resume             # resume the latest session
./zod examples/demo/harness.zd --resume <id|latest> # same, explicit target
./zod examples/demo/harness.zd --self-test          # 49-check UI self-test, exit 1 on failure
```

Flags: `--model <id>`, `--approvals <mode>`, `--sandbox <mode>`, `--cwd <dir>`,
`--yolo` (alias `--dangerously-bypass-approvals-and-sandbox`), `--json`,
`--resume [id|latest]`, `--help`, `--version`.

When stdin is not a TTY the interactive mode degrades to a line REPL
(streams to stdout, commands on `/`, shell on `!`; a blank line or EOF exits).
In exec mode, approvals default to `on-failure` unless `--approvals` is given,
so unattended runs do not hang on prompts. `--yolo` sets approvals to `never`
(never ask) and the sandbox to `danger-full-access`, prints a warning and
marks the TUI status line with a red `YOLO` badge.

## Keys (TUI)

| Key | Action |
| --- | --- |
| `enter` | send the turn; while a turn is running: queue the message |
| `ctrl-j` | insert a newline (multiline composer, up to 10 visible lines) |
| `esc` | clear the composer; during a turn: interrupt |
| `ctrl-c` | clear; when empty: quit; during a turn: interrupt |
| `ctrl-d` | quit when the composer is empty |
| `:q` / `:quit` | quit (`:q!` cancels the running turn and quits) |
| `ctrl-l` | clear the transcript |
| `ctrl-u` / `ctrl-w` / `ctrl-k` | clear line / delete word / kill to end |
| `ctrl-a` / `ctrl-e` | home / end |
| `left` / `right` | move the cursor |
| `up` / `down` | history (and transcript scroll with `pgup`/`pgdn`) |
| `tab` | complete `/command` names |
| approval overlay | `y` allow once, `a` allow for session, `n`/`esc`/`enter` deny, `v` expand args |
| model picker (`/model` with no argument) | up/down + enter, esc cancels |

The composer is borderless (`›` prompt with a dim placeholder), a dim status
line under it shows `model · Context NN% left · directory`, and a
`• Working (Ns • esc to interrupt)` row appears above the composer during a
turn. Queued messages run automatically when the current turn finishes; the
status line shows the queue depth.

## Commands

```
/help /status /model /stats /approvals /sandbox /new /resume
/compact /init /diff /review /btw /tree /clear /quit
!<cmd>  run a shell command directly (user-initiated, bypasses approvals;
        the sandbox still applies)
```

`/quit` (or `:q` / `:quit`) exits the TUI.

## Approvals and sandbox

Approvals (`untrusted` default, `on-failure`, `on-request`, `never`) gate every
non-read-only tool call. `untrusted` and `on-request` ask; `on-failure` allows
without asking; `never` (Codex semantics) allows without ever asking — it is
the approvals half of yolo mode. Read-only tools (`read_file`, `glob`, `grep`)
run unconditionally. Writes and `shell` ask the user, who can allow once or for
the session. Non-interactive runs answer `deny` automatically; exec mode
switches the default to `on-failure`. Allow rules can be configured as
`"tool"` or `"tool:pattern"` entries in `approvals_rules`.

Sandbox modes (`workspace-write` default, `read-only`, `danger-full-access`)
are enforced by shared guards inside the tools:

- `read-only`: all writes fail; `shell` is restricted to a read-only allowlist
  (`git status/diff/log/...`, `ls`, `cat`, `rg`, `sed -n`, ...).
- `workspace-write`: writes allowed under the working directory.
- `danger-full-access`: no path restrictions.

This is an advisory, in-process sandbox. OS-level sandboxing (seatbelt/landlock
style) is documented future work.

## Config

Precedence: built-in defaults → `~/.zodex/config.json` → `<cwd>/.zodex/config.json`
→ environment variables. Keys: `model`, `base_url`, `api_key`, `api_key_env`,
`approvals`, `approvals_rules`, `sandbox`, `max_steps`, `auto_compact`,
`compact_threshold`, `context_window`, `stream_usage`, `allow_exec`,
`max_bytes`, `temperature`, `timeout_ms`, `theme`.

Environment: `ZODEX_HOME`, `ZODEX_MODEL`, `ZODEX_BASE_URL`, `ZODEX_API_KEY`,
`ZODEX_APPROVALS`, `ZODEX_SANDBOX`, `ZODEX_MAX_STEPS`, `ZODEX_MAX_BYTES`,
`ZODEX_ALLOW_EXEC`, `ZODEX_AUTO_COMPACT`, `ZODEX_STREAM_USAGE`,
`ZODEX_TIMEOUT_MS`. API keys are also read from `HACKCLUB_AI_API_KEY`,
`HACK_CLUB_AI_API_KEY` and `OPENAI_API_KEY`.

`timeout_ms` is an *inactivity* timeout for streaming responses: the stream
keeps running while data keeps arriving and fails only after that many
milliseconds of silence. On a mid-stream timeout the turn engine keeps the
partial output, appends it to the conversation and continues from there
instead of replaying the request.

## Sessions

Every conversation is an append-only rollout at
`$ZODEX_HOME/sessions/<id>.jsonl` (default `~/.zodex/sessions/`), recording
meta, messages and tool traffic as they happen. `--resume` and `/resume` load
a rollout into the agent conversation and continue appending to the same file.
Session files are created lazily on their first record, so resuming never picks
an empty stub.

## Modules

| Module | Role |
| --- | --- |
| `harness_lib.zd` | TUI render loop, key handling, exec/line modes, app assembly |
| `agent.zd` | streaming turn engine: tool loop, retries, cancellation, compaction |
| `client.zd` | OpenAI-compatible HTTP client (`chat`, `stream_events`, `models`, stats) |
| `tools.zd` / `coding_tools.zd` | tool registry, read/write/edit/glob/grep/shell/apply_patch/update_plan |
| `apply_patch.zd` | Codex-style `*** Begin Patch` parser and all-or-nothing applier |
| `approvals.zd` | decision engine (modes, rules, session grants) + tool wrapping |
| `commands.zd` | slash commands and `!` passthrough |
| `pi.zd` | AGENTS.md context loading, system prompt, worktree/branch helpers |
| `render.zd` / `fx.zd` | ANSI-aware markdown/cells/diff rendering; spinner kit |
| `config.zd` / `cli.zd` | config precedence/env; argument parsing |
| `session.zd` / `history.zd` | rollout persistence; input history |
| `messages.zd` | conversation model |

## Tests

```sh
# pure modules + UI core: 116 checks, exit 1 on failure
ZODEX_HOME=/tmp/zodex_tests ./zod examples/demo/agentic/tests.zd

# UI self-test only (no disk writes): 49 checks
./zod examples/demo/harness.zd --self-test

# client + agent against the mock server: 38 checks
python3 examples/demo/agentic/mock_openai.py 8791 &
./zod examples/demo/agentic/_check_agent.zd
```

`tests.zd` skips the four session/history disk checks unless `ZODEX_HOME` is
set, so plain runs never touch a real `~/.zodex`. The mock server
(`mock_openai.py`, stdlib-only test tooling) is stateful: restart it between
runs that exercise retries.

## Parity with Codex

| Area | Status |
| --- | --- |
| Streaming TUI chat, markdown, diffs | done |
| Tool calling: shell, apply_patch, update_plan, read/write/edit/glob/grep | done |
| `apply_patch` format (add/update/delete, `@@` anchors, move) | done |
| Approvals modes + session grants + rules + yolo bypass (`--yolo`) | done |
| Sandbox modes (workspace-write / read-only / danger-full-access) | done (advisory; OS-level sandboxing is future work) |
| Sessions, resume, rollout JSONL (with transcript replay) | done |
| Config files + env vars | done |
| Exec mode, JSONL events, exit codes | done |
| Automatic context compaction | done (token-budget based) |
| Interruptions, retries, idle-timeout recovery, queued messages | done |
| MCP servers | not in scope |
| OAuth / ChatGPT login | not in scope (API key only) |
| Cloud tasks / Codex web | not in scope |
| Image inputs, web search tools | not in scope |
| OS-level sandbox enforcement | future work |

MCP, OAuth and cloud features are excluded by design: they need long-lived
external services and credentials that this local, dependency-free
implementation intentionally avoids. The provider interface is a plain
OpenAI-compatible endpoint, so staying local keeps the whole harness
inspectable and testable.

## Development notes

- The one interpreter change this work required is included here:
  `object/stream.go` now surfaces token usage and `finish_reason` from SSE
  `stream_poll` (plus `object/stream_usage_test.go`), which the agent needs for
  `/stats`, compaction budgets and length-continuation handling. Stream
  deadlines are enforced as an *inactivity* timeout: a watchdog resets on
  every byte and fails the stream only after `timeout_ms` of silence
  (`object/stream_idle_test.go`). `__http_post_headers` accepts up to
  `300000` ms.
- Modules avoid letting `error()` values cross more than one function
  boundary: on the VM engine, an error propagating further fails with
  `executing bytecode failed:` instead of being catchable. All fallible calls
  use envelopes or are wrapped in `try()` at the direct call site. The eval
  engine (`--eng=eval`) does not have this restriction.
- All definitions precede their call sites (the VM bytecode compiler rejects
  forward references); a few cycles are broken with hash indirection.
