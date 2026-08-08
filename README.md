# Zod Programming Language

<p align="center">
    <img src="media/Zod.png" alt="Zod Banner">
</p>

Zod is a small, interpreted, dynamically typed with a first-class functions, closures, arrays, hashes, and more.

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

### Option 3: Manual download

Grab the tarball for your OS/architecture from the [Releases page](https://github.com/theawakener0/Zod/releases), extract it, and put the `zod` binary on your `PATH`.

## Features

- First-class functions and closures
- Data types: integers, floats, booleans, strings, arrays, hashes, matrices, and `null`
- `if` / `elseif` / `else` control flow
- C-style `for` loops, and infinite `loop`
- `break` and `continue`
- `try(expr)` for recoverable errors
- Built-in functions for I/O (`println()`, `printf()`, `input()`, etc.), arrays, and hashes
- An interactive REPL

## Getting Started

### Prerequisites

- [Go](https://go.dev/dl/) 1.22 or newer

### Build

```sh
go build -o zod .
# if you installed using the script or go install you can just run zod
```

### Run the REPL

```sh
./zod
```

Inside the REPL, type `/clear` to clear the screen and `/exit` to quit.

### Run a script

```sh
./zod file.zd
```

### Run the examples

```sh
./zod examples/hello.zd
./zod examples/guess_the_number.zd
```

## Examples

### Hello, World!

```zod
println("Hello, World!")
```

### Variables and data types

```zod
let name = "Zod"
version := 1
let isAwesome = true
scores := [10, 20, 30]
let config = {"lang": "Zod", "year": 2026}
nothing := null
```

### Functions and closures

```zod
let add = fn(x, y) { x + y }
println(add(2, 3))  // 5

makeAdder := fn(x) {
    fn(y) { x + y }
}
let addFive = makeAdder(5)
println(addFive(10))  // 15
```

### Control flow

```zod
let grade = fn(score) {
    if (score >= 90) {
        "A"
    } elseif (score >= 80) {
        "B"
    } else {
        "C"
    }
}

println(grade(95))  // A
```

### Loops

```zod
// C-style for loop
for (i := 0; i < 5; i++) {
    println(i)  
/*
    0
    1
    2
    3
    4
*/
}

// While-style for loop
let x = 0
for (x < 3) { x++ }
println(x)  // 3

// Infinite loop with break / continue
let i = 0
loop {
    i++
    if (i == 3) { continue }
    if (i == 6) { break }
}
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
| Null      | `null`                          |

> [!NOTE]
> Hashes preserve insertion order: iterating with `keys()` / `vals()` and printing
> a hash always follow the order the keys were added.

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

## Built-in Functions

| Function              | Description                                    |
| --------------------- | ---------------------------------------------- |
| `len(x)`              | Length of a string, array, hash, or matrix     |
| `println(...)`        | Print each argument on its own line            |
| `printf(fmt, ...)`    | Formatted output (Go-style verbs)              |
| `input(prompt)`       | Read a line of input from the user             |
| `int(x)`              | Convert a string/bool to an integer            |
| `float(x)`            | Convert a string/int to a float                |
| `string(x)`           | Convert an int/bool to a string                |
| `type(x)`             | Return the type name of `x`                    |
| `first(arr)`          | First element of an array                      |
| `last(arr)`           | Last element of an array                       |
| `push(arr, x)`        | New array with `x` appended                    |
| `pop(arr)`            | New array without its last element             |
| `insert(hash, k, v)`  | New hash with `k: v` added                     |
| `remove(hash, k)`     | New hash without key `k`                       |
| `keys(hash)`          | Array of keys in a hash (insertion order)      |
| `vals(hash)`          | Array of values in a hash (insertion order)    |
| `contains(hash, k)`   | Whether a hash contains key `k`                |
| `random()`            | Random float in `[0, 1)`                       |
| `random(x)`           | Random integer in `[0, x)`                     |
| `matrix(r, c, data)`  | Matrix with `r` rows and `c` cols from `data`  |
| `make(size) / make(size, value)` | Create a new array with `size` elements, and value |
| `color(color, text)`  | Change the color of `text` to `color`          |
| `sleep(ms)`           | Pause execution for `ms` milliseconds          |
| `exp(x)`              | e^x |
| `pi()`                | π |

### Matrices

`matrix(r, c, data)` builds a `MATRIX` value (a row-major grid of integers or
floats) from a flat `data` array whose length must equal `r * c`.

```zod
let a = matrix(2, 3, [1, 2, 3, 4, 5, 6])
a[0]        // [1, 2, 3]  (first row)
a[0][1]     // 2
len(a)      // 2          (number of rows)
```

Matrices support element-wise `+` and `-` (same dimensions), scalar
`+ - * /`, and matrix multiplication with `*` (columns of the left must match
rows of the right). Mismatched dimensions raise an error.

```zod
let a = matrix(2, 3, [1, 2, 3, 4, 5, 6])
let b = matrix(2, 3, [1, 1, 1, 1, 1, 1])
println(a + b)             // [[2, 3, 4], [5, 6, 7]]
println(a * 2)             // [[2, 4, 6], [8, 10, 12]]

let c = matrix(3, 2, [7, 8, 9, 10, 11, 12])
println(a * c)             // [[58, 64], [139, 154]]
```

### Error handling with `try()`

`try(expr)` evaluates `expr` and always returns a two-element array
`[ok, value]` instead of aborting the program:

- On success it returns `[true, result]`.
- On failure it returns `[false, "error message"]`.

```zod
let r = try(int("42"))
println(r[0])              // true
println(r[1])              // 42

let r = try(int("abc"))
if (r[0]) {
    println("parsed:", r[1])
} else {
    println("failed:", r[1])   // failed: could not parse "abc" as integer
}
```

> [!NOTE]
> The `color` function is still under development so it doesn't have many colors.

## Acknowledgments

Inspired by Thorsten Ball's *Writing an Interpreter in Go*.

## License

MIT License. See [`LICENSE`](LICENSE).
