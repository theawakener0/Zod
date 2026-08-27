# Introduction

Zod is a small, interpreted, dynamically-typed language for joyful programming from 2D terminal games to tiny machine-learning experiments.

> [!CAUTION]
> Zod is pre-1.0 and under active development — APIs may change. Not yet recommended for production.

## What Is Zod?

Zod is a small, interpreted, dynamically-typed language implemented in Go and inspired by Thorsten Ball's *Writing an Interpreter in Go*. It offers first-class functions and closures, plus core types including integers, floats, booleans, strings, arrays, hashes, matrices, and null. Programs run via script files or an interactive REPL.

## Try It in 30 Seconds

Create `hello.zd`:

```zod
println("Hello, World!")
```

Run it, or start the REPL:

```sh
# Run a file
zod hello.zd
# Or start the REPL
zod
```

Inside the REPL, use `/clear` to clear the screen and `/exit` to leave. More examples live in `examples/` — try `hello.zd`, `guess_the_number.zd`, and `variables.zd`.

Install Zod with the install script:

```sh
curl -fsSL https://raw.githubusercontent.com/theawakener0/Zod/main/install.sh | sh
```

Or with Go (requires Go 1.22+):

```sh
go install github.com/theawakener0/Zod@latest
```

## Philosophy

Zod is designed to be small, simple, and easy to use. It is not a general-purpose replacement for Python or Ruby. Instead, it is built for people who want to enjoy programming and learn at the same time. The language is intentionally easy to migrate from, it is ideal if your goal is to learn fundamentals before moving to Go, Rust, or Zig. Above all, Zod emphasizes joy, fundamentals, and building what you love.

## Key Features

- First-class functions and closures
- Integers, floats, booleans, strings, arrays, hashes, matrices, and null
- Conditional branching with `if` / `elseif` / `else` (supports both `elseif` and `else if`)
- Loops: C-style `for`, while-style `for`, and infinite loops with `break` and `continue`
- Recoverable errors with `try(expr)` returning `[ok, value]`
- Rich built-ins for I/O, collections, and math
- Interactive REPL and script-file execution
- Operators including `+`, `-`, `*`, `/`, comparisons, and logical operators

## At a Glance: Built-in Functions

### I/O

| Function | Description |
| --- | --- |
| `println(...)` | Print each argument on its own line |
| `printf(fmt, ...)` | Formatted output with Go-style verbs |
| `input([prompt])` | Read a line from stdin; accepts 0 or 1 argument (optional prompt) |
| `sleep(ms)` | Pause execution for `ms` milliseconds; `ms` must be an integer |
| `color(color, text)` | Wrap `text` with ANSI color; `color` must be one of `RED`, `GREEN`, `YELLOW`, `BLUE`, `MAGENTA`, `CYAN` (uppercase) |

### Collections

| Function | Description |
| --- | --- |
| `len(x)` | Length of `x` where `x` is a string, array, hash, or matrix; for matrices returns row count, for strings returns byte length |
| `first(arr)` | First element of an array |
| `last(arr)` | Last element of an array |
| `push(arr, x)` | New array with `x` appended |
| `pop(arr)` | New array without its last element; returns `null` if the array is empty |
| `insert(hash, k, v)` | New hash with `k: v` added |
| `remove(hash, k)` | New hash without key `k` |
| `keys(hash)` | Array of keys in insertion order |
| `vals(hash)` | Array of values in insertion order |
| `contains(hash, k)` | Whether hash contains key `k` |
| `make(size [, value])` | Create an array with `size` elements; `size` is `int` or `float` (truncated), `NaN`/`Inf` rejected, max 10M; `value` defaults to `null` if omitted |

### Types & Math

| Function | Description |
| --- | --- |
| `int(x)` | Convert `x` to integer; accepts `string`, `int`, `float`, or `bool` |
| `float(x)` | Convert `x` to float; accepts `string`, `int`, `float`, or `bool` |
| `string(x)` | Convert `x` to string; accepts `string`, `int`, `float`, or `bool` |
| `type(x)` | Return the type name of `x` as a string |
| `random()` | Returns a random float in `[0, 1)` |
| `random(x)` | With one argument: if `x` is an `int`, returns a random `int` in `[0, x)`; if `x` is a `float`, returns a random `float` in `[0, x)`; `x` must be positive |
| `matrix(r, c, data)` | Create a matrix with `r` rows and `c` cols from flat array `data`; `r` and `c` are `int` or `float` (truncated), must be `>= 1`, `len(data)` must equal `r * c`, elements must be `int` or `float` |
| `exp(x)` | Returns e raised to `x`; accepts `int` or `float`, returns `float` |
| `pi()` | Returns π as float; accepts 0 arguments |

> [!NOTE]
> `try(expr)` is a special form, not a built-in, and is documented separately below. It requires exactly one argument.

For complete signatures, see `evaluator/builtins.go`.

## Peek: Matrices

`matrix(r, c, data)` builds a row-major grid of integers or floats. Index rows with `a[0]` and elements with `a[0][1]`; `len(a)` returns the row count.

```zod
let a = matrix(2, 3, [1, 2, 3, 4, 5, 6])
println(a[0])    // [1, 2, 3]
println(a[0][1]) // 2
println(len(a))  // 2
```

Matrices support element-wise `+` and `-` (same dimensions), scalar `+`, `-`, `*`, `/`, and matrix multiplication with `*` (columns of the left matrix must equal rows of the right, otherwise an error is returned).

## Peek: Error Handling with try()

`try(expr)` evaluates `expr` and always returns a two-element array. On success it returns `[true, result]`; on failure it returns `[false, "error message"]`. It requires exactly one argument.

```zod
let r = try(int("42"))
println(r[0]) // true
println(r[1]) // 42

let r = try(int("abc"))
if (r[0]) {
    println("parsed:", r[1])
} else {
    println("failed:", r[1]) // failed: could not parse "abc" as integer
}
```

## Next Steps

- Explore runnable examples in `examples/` — start with `hello.zd`, `variables.zd`, and `guess_the_number.zd`.
- Read the Language Reference in `README.md` for the full syntax and built-in details.
- Try the REPL (`zod`) to experiment interactively.


