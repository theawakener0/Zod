# Z-Lisp

A small Lisp interpreter written in ZodLang, in the spirit of the Monkey
interpreter from *Writing an Interpreter in Go*.

## Run

From the repository root:

```sh
go run . examples/demo/zlisp/repl.zd
```

Or with a built binary:

```sh
go build -o zod .
./zod examples/demo/zlisp/repl.zd
```

## Test

```sh
go run . examples/demo/zlisp/tests.zd
```

`tests.zd` is a self-contained suite (98 checks) that exercises the lexer,
parser, evaluator, closures, quoting, and errors. It runs on the VM.

## Example session

The startup banner is omitted below.

```text
Z-Lisp REPL (type 'exit' or 'quit' to quit)
z-lisp> (+ 1 2)
3
z-lisp> (define square (lambda (x) (* x x)))
<function>
z-lisp> (square 12)
144
z-lisp> (print (list 1 2 3))
(1 2 3)
z-lisp> (if 0 "truthy" "falsey")
"truthy"
z-lisp> '(a b c)
(a b c)
z-lisp> nope
error: unbound variable: nope
```

## Language

### Values

Integers (`42`, `-5`), floats (`3.14`), strings (`"hello"`), symbols (`foo`),
booleans (`true`, `false`), `nil`, and lists.

### Special forms

| Form | Meaning |
| --- | --- |
| `(quote x)` / `'x` | Return `x` without evaluating it |
| `(if test then else)` | Conditional; `else` is optional |
| `(define name value)` | Bind `name` in the current environment |
| `(lambda (params...) body...)` | Create a closure; the body is an implicit `begin` |
| `(begin body...)` | Evaluate forms in order, returning the last |

### Builtins

`+`, `-`, `*`, `/`, `=`, `<`, `>`, `<=`, `>=`, `list`, `print`.

Only `nil` and `false` are falsey. `/` returns an integer when both operands
are integers and the division is exact; otherwise it returns a float. `=` is
type-strict (`1` is not `1.0`) and structural for lists; functions and builtins
never compare equal. `<`, `>`, `<=`, and `>=` compare numbers only. Numeric
literals outside the 64-bit range produce an error instead of aborting.

## Layout

| File | Role |
| --- | --- |
| `token.zd` | Token type constants |
| `lexer.zd` | Hand-written lexer |
| `parser.zd` | Recursive-descent parser producing hash-based AST nodes |
| `ast.zd` | AST pretty-printer |
| `environment.zd` | Lexical environments (hash store + parent link) |
| `evaluator.zd` | Tree-walking evaluator, special forms, closures |
| `builtins.zd` | Global environment and primitive functions |
| `repl.zd` | Interactive REPL |
| `tests.zd` | Regression suite |

## Notes and limitations

- The REPL reads one line at a time, so multi-line forms are not supported.
- `input()` cannot distinguish an empty line from end-of-file. When piping
  input, finish with `exit` or `quit`; closing stdin without one makes the REPL
  loop, because every read returns an empty string.
- Strings are byte-oriented in ZodLang. The lexer uses `str.slice` to read
  characters because strings do not support `s[i]` indexing.
- Local forward references are rejected by the VM, so the parser's mutually
  recursive functions go through a small hash indirection (`exprRef`).
- Numeric literals are 64-bit and do not support exponent notation (`1e3` is
  rejected as an invalid number). Runtime integer arithmetic wraps on overflow;
  only out-of-range literals produce an error. Float arithmetic follows
  IEEE-754, so overflow can produce `+Inf.0` or `NaN.0`, which the reader
  cannot parse back.
- Deep non-tail recursion can exceed the VM call-depth limit and abort the
  process.
- The empty form `()` evaluates to `nil`.
- Signed literals (`-5`, `+5`) are single tokens. Without whitespace, a sign
  continues the token: `1-2` is rejected as an invalid number rather than read
  as `1` followed by `-2`.
- Identifiers are ASCII only; unknown string escapes are preserved literally.
- No macros, vectors, `set!`, or tail-call optimization yet.
