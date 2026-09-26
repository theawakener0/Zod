# Introduction

Zod is a small (let's say it's a medium sized), interpreted, dynamically typed with a first-class functions, closures, arrays, hashes, and more.

> [!CAUTION]
> Zod is pre-1.0 and under active development — APIs may change. Not yet recommended for production.

## What Is Zod?

Zod is a small, interpreted, dynamically-typed language implemented in Go and inspired by Thorsten Ball's *Writing an Interpreter in Go* and *Writing a Compiler in Go*. It offers first-class functions and closures, plus core types including integers, floats, booleans, strings, arrays, hashes, matrices, and null. Programs run via script files or an interactive REPL.

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

Or with Go (requires Go 1.27+):

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
- Operators including `+`, `-`, `*`, `/`, `%`, comparisons, and logical operators
- A top-level denfition 

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

## Standard Library

New code should prefer the standard library in `stdlib/` (21 modules: `fmt`, `str`, `array`, `hash`, `math`, `matrix`, `rand`, `time`, `os`, `fs`, `path`, `io`, `term`, `json`, `http`, `sort`, `bytes`, `log`, `test`, `crypto`, `iter`) over the legacy core built-ins above. Import a module by path with an alias, then call its `pub` functions as properties:

```zod
import "std/fmt" as fmt
fmt.println("Hello, World!")
```

Individual names can be imported too:

```zod
from "std/fmt" import println
println("via from-import")
```

### How `std/` is resolved

A `"std/..."` path is looked up in this order: `$ZOD_STDLIB` if set, `stdlib/` next to the executable, `./stdlib` in the working directory, then `stdlib/` in parent directories of the script. Running from the project root uses `./stdlib`.

### Example: strings and arrays

```zod
import "std/str" as str
import "std/array" as array
import "std/fmt" as fmt

fmt.println(str.upper("hello"))                    // HELLO
fmt.println(str.join(["a", "b"], ","))             // a,b
fmt.println(array.map([1, 2, 3], fn(x) { x * 2 })) // [2, 4, 6]
```

### Example: files and JSON

```zod
import "std/json" as json
import "std/fs" as fs
import "std/fmt" as fmt

fs.write_file("hello.json", json.stringify({"name": "Zod"}))
let h = json.parse(fs.read_file("hello.json"))
fmt.println(h["name"]) // Zod
```

### Example: aliases avoid collisions

Several modules export the same name (`join` in `str`/`array`/`path`, `contains` in `str`/`array`/`hash`). Always import with an alias and call through it:

```zod
import "std/str" as str
import "std/path" as path
import "std/fmt" as fmt

fmt.println(str.join(["a", "b"], "-")) // a-b
fmt.println(path.join("a", "b"))       // a/b
```

An imported name also shadows a same-named core built-in (e.g. `array.push` vs `push`) on both the eval and VM engines.

### Example: bytes and error values

```zod
import "std/bytes" as bytes
import "std/fmt" as fmt

let b = bytes.from_str("hello")
fmt.println(bytes.len(b))                  // 5
fmt.println(bytes.to_str(bytes.slice(b, 1, 4))) // ell

// string() formats errors too: nested string(error()) works on both engines.
fmt.println(string(error("boom")))         // Error: boom
fmt.println(type(error("boom")))           // ERROR
fmt.println(is_error(error("boom")))       // true
```

Capture errors with `try()` instead of binding them directly — a bare `let e = error("x")` halts evaluation by design:

```zod
let r = try(error("x"))
fmt.println(r[0]) // false
fmt.println(r[1]) // x
```

### Example: crypto (deterministic lengths)

```zod
import "std/crypto" as crypto
println(len(crypto.uuid())) // 36
println(len(crypto.random_hex(8))) // 16
println(len(crypto.sha256("hi"))) // 64
```

### Example: iter

```zod
import "std/iter" as iter
import "std/fmt" as fmt

let sum = iter.reduce([1, 2, 3, 4], fn(acc, x) { acc + x }, 0)
fmt.println(sum) // 10
fmt.println(iter.chain([1, 2], [3, 4])) // [1, 2, 3, 4]
fmt.println(iter.take([1, 2, 3, 4, 5], 3)) // [1, 2, 3]
fmt.println(iter.drop([1, 2, 3, 4, 5], 2)) // [3, 4, 5]
```

See `examples/std_iter_demo.zd` (run with `go run . examples/std_iter_demo.zd` and `go run . --eng=eval examples/std_iter_demo.zd`).

### Example: http (offline-safe)

No external network required — safe for CI. The demo below targets an unroutable loopback port and captures the expected connection error with `try()`:

```zod
import "std/http" as http
import "std/fmt" as fmt

let r = try(http.get_timeout("http://127.0.0.1:1/", 500))
fmt.println(r[0]) // false (connection refused)
```

`std/http` exports `get(url)`, `get_timeout(url, ms)`, and `post(url, body)` backed by `__http_get`/`__http_post`. `get`/`post` use a 10s timeout. See `examples/std_http_demo.zd`.

### API conventions

- Functional vs mutate: `array.*` returns new arrays and never mutates its
  input; `hash.set` mutates in place and returns the same hash (documented
  as-is, do not rely on a copy).
- Printing: `fmt.println(x)` takes a single argument; the legacy variadic
  `println(...)` prints each argument on its own line and is kept for
  compatibility.
- Imports: prefer `import "std/x" as x` and call through the alias; avoid
  star-imports because names like `join` (`str`/`array`/`path`) and
  `contains` (`str`/`array`/`hash`) collide.
- Example:
  ```zod
  import "std/str" as str
  import "std/fmt" as fmt
  fmt.println(str.join(["a", "b"], "-"))
  ```

## Peek: Matrices

`matrix(r, c, data)` builds a row-major grid of integers or floats. Index rows with `a[0]` and elements with `a[0][1]`; `len(a)` returns the row count.

```zod
let a = matrix(2, 3, [1, 2, 3, 4, 5, 6])
println(a[0])    // [1, 2, 3]
println(a[0][1]) // 2
println(len(a))  // 2
```

Matrices support element-wise `+` and `-` (same dimensions), scalar `+`, `-`, `*`, `/`, and matrix multiplication with `*` (columns of the left matrix must equal rows of the right, otherwise an error is returned).

The `std/matrix` module wraps this with `new`, `eye`, `zeros`, `ones`, `rows`, `row`, `get`, `transpose`, `mul`, `det2`, `det3`, `det`, `inv2`, `inv`. `det` is the general NxN determinant (`det2`/`det3` are specialized 2x2/3x3 helpers); `inv` is the general NxN inverse (`inv2` is the specialized 2x2 helper). A singular matrix is an error — capture it with `try()`:

```zod
import "std/matrix" as m
import "std/fmt" as fmt

let a = m.new(2, 2, [1, 2, 3, 4])
fmt.println(m.mul(a, m.eye(2))) // [[1, 2], [3, 4]]
fmt.println(m.det(a))           // -2
fmt.println(m.det3(m.eye(3)))   // 1
fmt.println(m.inv2(m.new(2, 2, [4, 7, 2, 6]))) // [[0.6, -0.7], [-0.2, 0.4]]
let r = try(m.inv2(m.new(2, 2, [1, 2, 2, 4])))
fmt.println(r[0]) // false (singular matrix)
fmt.println(r[1]) // singular matrix
let r2 = try(m.inv(m.new(2, 2, [1, 2, 2, 4])))
fmt.println(r2[0]) // false (singular matrix)
```

See `examples/std_matrix_demo.zd` and `examples/std_matrix2_demo.zd` (run each with `go run . <file>` and `go run . --eng=eval <file>`).

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

## Modules & pub

Bindings are private by default. Mark a top-level `let` or `:=` definition with `pub` to export it from a module. Only public bindings can be imported.

```zod
// pub_lib.zd
pub let version = 1
let secret = "hidden"

pub greet := fn(name) { "hello, " + name }
```

```zod
// Star import: brings all public bindings into scope.
from "examples/pub_lib.zd" import *
println(version)       // 1
println(greet("world")) // hello, world

// Named import: brings only the listed public bindings into scope.
from "examples/pub_lib.zd" import greet

// Module import with alias: access public bindings as properties.
import "examples/pub_lib.zd" as m
println(m.version)
println(m.greet("alias"))
```

Accessing a private binding fails:

```zod
from "examples/pub_lib.zd" import secret // error: module ... has no public export named secret (secret is private)
println(m.secret)                        // error: undefined property secret on module m (secret is private)
```

`pub` only applies to top-level `let` and `:=` definitions. A `pub` inside a block or function body stays local to that scope and is never exported. See `examples/pub_lib.zd` and `examples/modules_pub.zd` for a runnable demo.

## Next Steps

- Explore runnable examples in `examples/` — start with `hello.zd`, `variables.zd`, and `guess_the_number.zd`.
- Read the Language Reference in `README.md` for the full syntax and built-in details.
- Try the REPL (`zod`) to experiment interactively.


