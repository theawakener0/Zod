# Zod Programming Language

<p align="center">
    <img src="media/Zod.jpg" alt="Zod Banner">
</p>

Zod is a small (let's say it's a medium sized), interpreted, dynamically typed with a first-class functions, closures, arrays, hashes, and more.

## Install

Prebuilt binaries for Linux, macOS, and Windows are attached to every [GitHub Release](https://github.com/theawakener0/Zod/releases).

### Option 1: Download script

```sh
curl -fsSL https://raw.githubusercontent.com/theawakener0/Zod/main/install.sh | sh
```

### Option 2: go install

```sh
go install github.com/theawakener0/Zod@latest
```

Requires [Go](https://go.dev/dl/) 1.22 or newer and installs the `Zod` executable into your Go bin directory.

## Demo

### Conway's Game of Life

![Conway's Game of Life](media/demo.gif)

## Features

- First-class functions and closures
- Data types: integers, floats, booleans, strings, arrays, hashes, matrices, and `null`
- `if` / `elseif` / `else` control flow
- C-style `for` loops, and infinite `loop`
- `break` and `continue`
- `try(expr)` for recoverable errors
- Standard library in [`stdlib/`](stdlib/) (`std/fmt`, `std/str`, `std/array`, …) via `import "std/fmt" as fmt`
- Legacy core built-ins for I/O (`println()`, `printf()`, `input()`, etc.), arrays, and hashes (kept for compat)
- An interactive REPL

## Getting Started

### Prerequisites

- [Go](https://go.dev/dl/) 1.22 or newer

### Build

```sh
# if you installed manually run this
go build -o zod .
```

> [!NOTE]
> ZodLang now has a new engine, the VM, which is faster and more powerful. You can run the old engine with `--eng=eval`.

### Run the REPL

```sh
# if you installed using the script
zod

# if you installed using go install
Zod
```

Inside the REPL, type `/clear` to clear the screen and `/exit` to quit.

### Run a script

```sh
# if you installed using the script
zod file.zd

# if you installed using the script and want to run the eval engine
zod --eng=eval file.zd

# if you installed using go install
Zod file.zd

# if you installed using go install and want to run the eval engine
Zod --eng=eval file.zd
```

## Examples

### Hello, World!

```zod
println("Hello, World!")
```

More runnable examples live in [`examples/`](examples/).

## Language Reference

### Data types

| Type      | Example                         |
| --------- | ------------------------------- |
| Integer   | `42`, `-7`, `0`                 |
| Float     | `3.14`, `-0.5`, `20.0`          |
| Boolean   | `true`, `false`                 |
| String    | `"Hello, World!"`               |
| Array     | `[1, 2, 3]`, `[]`               |
| Hash      | `{"name": "Zod"}`, `{}`         |
| Matrix    | `matrix(2, 2, [1, 2, 3, 4])`    |
| Bytes     | `__bytes("hi")` (private intrinsic; use `std/*`) |
| Error     | `error("boom")`                 |
| Null      | `null`                          |

### Operators

| Category      | Operators                                |
| ------------- | ---------------------------------------- |
| Arithmetic    | `+`  `-`  `*`  `/`                       |
| Comparison    | `==`  `!=`  `<`  `>`  `<=`  `>=`         |
| Logical       | `&&`  `\|\|`  `!`                        |
| Increment     | `++`  `--` (prefix and postfix)          |

### Assignment

| Operator | Meaning                                   |
| -------- | ----------------------------------------- |
| `let`    | Declare a new variable: `let x = 5`       |
| `:=`     | Assign (declares if needed): `x := 5`     |

Index assignment works on arrays and hashes too: `nums[0] = 10`, `user["age"] += 1`.

### Modules

Bindings are private by default. Prefix a top-level `let` or `:=` with `pub` to export it.

```zod
// pub_lib.zd
pub let version = 1
let secret = "hidden"
pub greet := fn(name) { "hello, " + name }
```

```zod
from "examples/pub_lib.zd" import *      // all public bindings
from "examples/pub_lib.zd" import greet  // listed public bindings only
import "examples/pub_lib.zd" as m        // property access: m.version, m.greet("hi")
```

Importing or accessing a private binding is an error (e.g. `secret is private`). See `docs/introduction.md` (Modules & pub) and `examples/modules_pub.zd` for details.

## Standard Library (preferred)

New code should use the standard library in [`stdlib/`](stdlib/) (20 modules) instead of the legacy core built-ins below. Each module is imported by path with an alias:

```zod
import "std/fmt" as fmt
fmt.println("Hello, World!")
```

| Module   | Import example              | Key exports                                                              |
| -------- | --------------------------- | ------------------------------------------------------------------------ |
| `fmt`    | `import "std/fmt" as fmt`   | `print`, `println`, `eprint`, `printf`, `sprintf`                        |
| `str`    | `import "std/str" as str`   | `split`, `join`, `trim`, `upper`, `lower`, `replace`, `contains`, `has_prefix`, `has_suffix`, `to_int`, `to_float` |
| `array`  | `import "std/array" as array` | `first`, `last`, `push`, `pop`, `pop_last`, `map`, `filter`, `find`, `any`, `all`, `reverse`, `flatten`, `join`, `sort`, `slice`, `range`, `make` |
| `hash`   | `import "std/hash" as hash` | `get`, `set`, `contains`, `has`, `del`, `keys`, `vals`, `merge`, `remove`, `from_pairs` |
| `math`   | `import "std/math" as math` | `pi`, `e`, `tau`, `abs`, `min`, `max`, `clamp`, `floor`, `ceil`, `round`, `exp`, `log`, `sqrt`, `pow`, `sin`, `cos`, `tan` |
| `matrix` | `import "std/matrix" as matrix` | `new`, `eye`, `zeros`, `ones`, `rows`, `row`, `get`, `transpose`, `mul`, `det2`      |
| `rand`   | `import "std/rand" as rand` | `seed`, `float`, `int`, `range`, `choice`, `shuffle`                     |
| `time`   | `import "std/time" as time` | `sleep`, `now_ms`, `now_s`, `elapsed_note`                               |
| `os`     | `import "std/os" as os`     | `args`, `env_get`, `env_set`, `env_list`, `exit`, `cwd`                              |
| `fs`     | `import "std/fs" as fs`     | `read_file`, `write_file`, `append_file`, `exists`, `ls`, `mkdir`, `rm`, `stat` (includes `mtime` unix field) |
| `path`   | `import "std/path" as path` | `join`, `basename`, `dirname`, `ext`                                     |
| `io`     | `import "std/io" as io`     | `read_stdin`, `write_stdout`, `read_lines`                               |
| `term`   | `import "std/term" as term` | `color`, `input`, `clear`                                                |
| `json`   | `import "std/json" as json` | `parse`, `stringify`, `read`, `write`                                    |
| `sort`   | `import "std/sort" as sort` | `sorted`, `sort_by` (comparator must return int)                         |
| `bytes`  | `import "std/bytes" as bytes` | `from_str`, `to_str`, `len`, `slice`, `concat`                         |
| `log`    | `import "std/log" as log`   | `info`, `warn`, `error`, `debug`                                         |
| `test`   | `import "std/test" as test` | `assert`, `assert_eq`, `assert_true`                                     |
| `crypto` | `import "std/crypto" as crypto` | `uuid`, `random_bytes`, `random_hex`, `sha256`                       |
| `iter`   | `import "std/iter" as iter` | `reduce`, `chain`, `take`, `drop`                                        |

- The `__*` names (about 50 intrinsics such as `__print`, `__str_split`, `__read_file`) are private implementation details backing `std/*`. Do not call them directly.
- New types: `Bytes` (`type(x)` returns `"BYTES"`) and `Error` values (`type(error("x"))` returns `"ERROR"`). New globals are `error(msg)` (creates an error value) and `is_error(x)` (tests for one); `string()` also formats `Array`, `Hash`, `Matrix`, `Bytes`, `null`, and `Error` (`string(error("x"))` returns `"Error: x"`).
- Capture errors with `try()`: `let r = try(error("x"))` returns `[false, "x"]`. A bare `let e = error("x")` halts evaluation by design, so always wrap error-producing expressions in `try()` (or nest them, e.g. `string(error("x"))`, which works on both engines).
- An imported name shadows a same-named core built-in (e.g. `array.push` vs `push`) on both engines. Several modules export the same name (`join` in `str`/`array`/`path`, `contains` in `str`/`array`/`hash`), so always import with an alias and call through it.
- `std/` lookup order and runnable examples are documented in [docs](docs/introduction.md).

### Legacy core (kept for compat)

The original 24 built-ins below still work, but new code should use `std/*` instead.

| Function              | Description                                    |
| --------------------- | ---------------------------------------------- |
| `len(x)`              | Length of a string, array, hash, or matrix     |
| `println(...)`        | Print each argument on its own line            |
| `printf(fmt, ...)`    | Formatted output (Go-style verbs)              |
| `input(prompt)`       | Read a line of input from the user             |
| `int(x)`              | Convert `x` to integer                         |
| `float(x)`            | Convert `x` to float                           |
| `string(x)`           | Convert `x` to string (also Array, Hash, Matrix, Bytes, null, Error → `"Error: msg"`) |
| `type(x)`             | Return the type name of `x`                    |
| `first(arr)`          | First element of an array                      |
| `last(arr)`           | Last element of an array                       |
| `push(arr, x)`        | New array with `x` appended                    |
| `pop(arr)`            | New array without its last element             |
| `insert(hash, k, v)`  | New hash with `k: v` added                     |
| `remove(hash, k)`     | New hash without key `k`                       |
| `keys(hash)`          | Array of keys in insertion order               |
| `vals(hash)`          | Array of values in insertion order             |
| `contains(hash, k)`   | Whether hash contains key `k`                  |
| `random(...)`         | Random float in `[0, 1)`; with one arg, in `[0, x)` |
| `matrix(r, c, data)`  | Matrix with `r` rows and `c` cols from `data`  |
| `make(size) / make(size, value)` | Create a new array with `size` elements, and value |
| `color(color, text)`  | Change the color of `text` to `color`          |
| `sleep(ms)`           | Pause execution for `ms` milliseconds          |
| `exp(x)`              | e^x |
| `pi()`                | π |

> [!NOTE]
> The `color` function is still under development so it doesn't have many colors.

for more details. Go to the [docs](docs/introduction.md).

## Acknowledgments

Inspired by Thorsten Ball's *Writing an Interpreter in Go* and *Writing a Compiler in Go*.

## License

MIT License. See [`LICENSE`](LICENSE).
