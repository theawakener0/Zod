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
go run . --eng=eval examples/demo/zlisp/tests.zd
```

`tests.zd` is a self-contained suite (32 checks) that exercises the lexer,
parser, evaluator, closures, quoting, errors, and both engines.

## Example session

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

Only `nil` and `false` are falsey. `/` returns an integer when the division is
exact, otherwise a float.

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
- `input()` cannot distinguish an empty line from end-of-file; when piping
  input, finish with `exit` or `quit`.
- Strings are byte-oriented in ZodLang. The lexer uses `str.slice` to read
  characters because strings do not support `s[i]` indexing.
- Local forward references are rejected by the VM, so the parser's mutually
  recursive functions go through a small hash indirection (`exprRef`).
- No macros, vectors, `set!`, or tail-call optimization yet.
