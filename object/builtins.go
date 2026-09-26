package object

import (
	"bufio"
	crand "crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"
	"unicode/utf8"
	"unsafe"
)

var (
	NULL  = &Null{}
	TRUE  = &Boolean{Value: true}
	FALSE = &Boolean{Value: false}
)

var Builtins = []struct {
	Name    string
	Builtin *Builtin
}{
	{"len", &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}

			switch arg := args[0].(type) {
			case *String:
				return NewInteger(int64(len(arg.Value)))
			case *Bytes:
				return NewInteger(int64(len(arg.Value)))
			case *Array:
				return NewInteger(int64(len(arg.Elements)))
			case *Matrix:
				return NewInteger(int64(arg.Rows))
			case *Hash:
				return NewInteger(int64(len(arg.Pairs)))
			default:
				return newError("argument to `len` not supported. got=%s", args[0].Type())
			}
		},
	}},
	{"println", &Builtin{
		Fn: func(args ...Object) Object {
			for _, arg := range args {
				fmt.Println(arg.Inspect())
			}

			return NULL
		},
	}},
	{"printf", &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) < 1 {
				return newError("wrong number of arguments. got=%d, want=1 or more", len(args))
			}

			format, ok := args[0].(*String)
			if !ok {
				return newError("first argument to `printf` not a string. got=%s", args[0].Type())
			}

			valArgs := make([]any, len(args)-1)
			for i, arg := range args[1:] {
				valArgs[i] = objectToValue(arg)
			}

			var buf strings.Builder
			_, _ = fmt.Fprintf(&buf, format.Value, valArgs...)
			out := buf.String()
			if strings.Contains(out, "%!") {
				return newError("printf format error: verb/type mismatch or missing argument in %q -> %q", format.Value, out)
			}
			fmt.Print(out)
			return NULL
		},
	}},
	{"input", &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) > 1 {
				return newError("wrong number of arguments. got=%d, want=0 or 1", len(args))
			}

			if len(args) == 1 {
				fmt.Print(args[0].Inspect())
			}

			reader := getStdinReader()
			text, _ := reader.ReadString('\n')
			text = strings.TrimSpace(text)

			return &String{Value: text}
		},
	}},
	{"int", &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}

			switch arg := args[0].(type) {
			case *String:
				s := arg.Value
				var val int64
				var err error

				if len(s) >= 2 && s[0] == '0' && strings.Contains("xXbBoO", string(s[1])) {
					val, err = strconv.ParseInt(s, 0, 64)
				} else {
					val, err = strconv.ParseInt(s, 10, 64)
				}
				if err != nil {
					return newError("could not parse %q as integer", arg.Value)
				}
				return NewInteger(val)
			case *Integer:
				return arg
			case *Float:
				return NewInteger(int64(arg.Value))
			case *Boolean:
				if arg.Value {
					return NewInteger(1)
				}
				return NewInteger(0)
			default:
				return newError("argument to `int` not supported. got=%s", args[0].Type())
			}
		},
	}},
	{"float", &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}

			switch arg := args[0].(type) {
			case *String:
				val, err := strconv.ParseFloat(arg.Value, 64)
				if err != nil {
					return newError("could not parse %q as float", arg.Value)
				}
				return &Float{Value: val}
			case *Float:
				return arg
			case *Integer:
				return &Float{Value: float64(arg.Value)}
			case *Boolean:
				if arg.Value {
					return &Float{Value: 1.0}
				}
				return &Float{Value: 0.0}
			default:
				return newError("argument to `float` not supported. got=%s", args[0].Type())
			}
		},
	}},
	{"string", &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}

			switch arg := args[0].(type) {
			case *String:
				return arg
			case *Integer:
				return &String{Value: arg.Inspect()}
			case *Float:
				return &String{Value: arg.Inspect()}
			case *Boolean:
				if arg.Value {
					return &String{Value: "true"}
				}
				return &String{Value: "false"}
			case *Bytes:
				return &String{Value: string(arg.Value)}
			case *Null:
				return &String{Value: "null"}
			case *Array:
				return &String{Value: arg.Inspect()}
			case *Hash:
				return &String{Value: arg.Inspect()}
			case *Matrix:
				return &String{Value: arg.Inspect()}
			case *Error:
				return &String{Value: arg.Inspect()}
			default:
				return newError("argument to `string` not supported. got=%s", args[0].Type())
			}
		},
	}},
	{"type", &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}
			return &String{Value: string(args[0].Type())}
		},
	}},
	{"first", &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}

			switch arg := args[0].(type) {
			case *Array:
				if len(arg.Elements) > 0 {
					return arg.Elements[0]
				}
				return NULL
			default:
				return newError("argument to `first` not supported. got=%s", args[0].Type())
			}
		},
	}},
	{"last", &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}

			switch arg := args[0].(type) {
			case *Array:
				if len(arg.Elements) > 0 {
					return arg.Elements[len(arg.Elements)-1]
				}
				return NULL
			default:
				return newError("argument to `last` not supported. got=%s", args[0].Type())
			}
		},
	}},
	{"pop", &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}
			switch arg := args[0].(type) {
			case *Array:
				length := len(arg.Elements)
				if length > 0 {
					newElements := make([]Object, length-1)
					copy(newElements, arg.Elements[:length-1])
					return &Array{Elements: newElements}
				}
				return NULL
			default:
				return newError("argument to `pop` not supported. got=%s", args[0].Type())
			}
		},
	}},
	{"push", &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 2 {
				return newError("wrong number of arguments. got=%d, want=2", len(args))
			}

			switch arg := args[0].(type) {
			case *Array:
				length := len(arg.Elements)

				newElements := make([]Object, length+1)
				copy(newElements, arg.Elements)
				newElements[length] = args[1]

				return &Array{Elements: newElements}
			default:
				return newError("argument to `push` not supported. got=%s", args[0].Type())
			}
		},
	}},
	{"insert", &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 3 {
				return newError("wrong number of arguments. got=%d, want=3", len(args))
			}

			hash, ok := args[0].(*Hash)
			if !ok {
				return newError("argument to `insert` not supported. got=%s", args[0].Type())
			}

			key, ok := args[1].(Hashable)
			if !ok {
				return newError("unusable as hash key: %s", args[1].Type())
			}

			newPairs := make(map[HashKey]HashPair, len(hash.Pairs)+1)
			newOrder := make([]HashKey, 0, len(hash.Order)+1)
			for k, v := range hash.Pairs {
				newPairs[k] = v
			}
			newOrder = append(newOrder, hash.Order...)

			hashKey := key.HashKey()
			if _, exists := newPairs[hashKey]; !exists {
				newOrder = append(newOrder, hashKey)
			}
			newPairs[hashKey] = HashPair{Key: args[1], Value: args[2]}

			return &Hash{Pairs: newPairs, Order: newOrder}
		},
	}},
	{"remove", &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 2 {
				return newError("wrong number of arguments. got=%d, want=2", len(args))
			}

			hash, ok := args[0].(*Hash)
			if !ok {
				return newError("argument to `remove` not supported. got=%s", args[0].Type())
			}

			key, ok := args[1].(Hashable)
			if !ok {
				return newError("unusable as hash key: %s", args[1].Type())
			}

			newPairs := make(map[HashKey]HashPair, len(hash.Pairs))
			newOrder := make([]HashKey, 0, len(hash.Order))
			for k, v := range hash.Pairs {
				newPairs[k] = v
			}

			hashKey := key.HashKey()
			delete(newPairs, hashKey)
			for _, k := range hash.Order {
				if k != hashKey {
					newOrder = append(newOrder, k)
				}
			}

			return &Hash{Pairs: newPairs, Order: newOrder}
		},
	}},
	{"keys", &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}

			hash, ok := args[0].(*Hash)
			if !ok {
				return newError("argument to `keys` not supported. got=%s", args[0].Type())
			}

			keys := make([]Object, 0, len(hash.Pairs))
			for _, k := range hash.Order {
				pair := hash.Pairs[k]
				keys = append(keys, pair.Key)
			}

			return &Array{Elements: keys}
		},
	}},
	{"vals", &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}

			hash, ok := args[0].(*Hash)
			if !ok {
				return newError("argument to `vals` not supported. got=%s", args[0].Type())
			}

			values := make([]Object, 0, len(hash.Pairs))
			for _, k := range hash.Order {
				pair := hash.Pairs[k]
				values = append(values, pair.Value)
			}

			return &Array{Elements: values}
		},
	}},
	{"contains", &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 2 {
				return newError("wrong number of arguments. got=%d, want=2", len(args))
			}

			hash, ok := args[0].(*Hash)
			if !ok {
				return newError("argument to `contains` not supported. got=%s", args[0].Type())
			}

			key, ok := args[1].(Hashable)
			if !ok {
				return newError("unusable as hash key: %s", args[1].Type())
			}

			if _, ok := hash.Pairs[key.HashKey()]; ok {
				return TRUE
			}
			return FALSE
		},
	}},
	{"random", &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) == 0 {
				return &Float{Value: rand.Float64()}
			}
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=0 or 1", len(args))
			}

			switch arg := args[0].(type) {
			case *Integer:
				if arg.Value <= 0 {
					return newError("argument to `random` must be positive. got=%d", arg.Value)
				}
				return NewInteger(rand.Int63n(arg.Value))
			case *Float:
				if arg.Value <= 0 {
					return newError("argument to `random` must be positive. got=%f", arg.Value)
				}
				return &Float{Value: rand.Float64() * arg.Value}
			default:
				return newError("argument to `random` not supported. got=%s", args[0].Type())
			}
		},
	}},
	{"matrix", &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 3 {
				return newError("wrong number of arguments. got=%d, want=3", len(args))
			}

			rows, ok := toInt(args[0])
			if !ok {
				return newError("first argument to `matrix` not an integer or float. got=%s", args[0].Type())
			}

			cols, ok := toInt(args[1])
			if !ok {
				return newError("second argument to `matrix` not an integer or float. got=%s", args[1].Type())
			}

			val, ok := args[2].(*Array)
			if !ok {
				return newError("third argument to `matrix` not an array. got=%s", args[2].Type())
			}

			if rows < 1 || cols < 1 {
				return newError("matrix dimensions must be positive. got=%d, %d", rows, cols)
			}

			if len(val.Elements) != rows*cols {
				return newError("array length does not match matrix dimensions. got=%d, want=%d", len(val.Elements), rows*cols)
			}

			data := make([][]Object, rows)
			for i := 0; i < rows; i++ {
				row := make([]Object, cols)
				for j := 0; j < cols; j++ {
					e := val.Elements[i*cols+j]
					switch e.(type) {
					case *Integer, *Float:
					default:
						return newError("matrix elements must be integers or floats. got=%s", e.Type())
					}
					row[j] = e
				}
				data[i] = row
			}

			return &Matrix{Rows: rows, Cols: cols, Data: data}
		},
	}},
	{"make", &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) < 1 || len(args) > 2 {
				return newError("wrong number of arguments. got=%d, want=1 or 2", len(args))
			}

			var size int64
			switch args[0].(type) {
			case *Integer:
				size = args[0].(*Integer).Value
			case *Float:
				fVal := args[0].(*Float).Value
				if math.IsNaN(fVal) {
					return newError("size to `make` must be integer, got NaN")
				}
				if math.IsInf(fVal, 0) {
					return newError("size to `make` must be finite, got %v", fVal)
				}
				size = int64(fVal)
			default:
				return newError("first argument to `make` not an integer or float. got=%s", args[0].Type())
			}

			if size < 0 {
				return newError("size to `make` must be non-negative. got=%d", size)
			}
			const maxArraySize int64 = 10000000
			if size > maxArraySize {
				return newError("size to `make` too large. got=%d, want <= %d", size, maxArraySize)
			}

			var fill Object = NULL
			if len(args) == 2 {
				fill = args[1]
			}

			elements := make([]Object, size)
			for i := range elements {
				elements[i] = fill
			}

			return &Array{Elements: elements}
		},
	}},
	{"color", &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 2 {
				return newError("wrong number of arguments. got=%d, want=2", len(args))
			}

			switch arg := args[0].(type) {
			case *String:
				switch arg.Value {
				case "RED":
					return &String{Value: "\033[31m" + args[1].Inspect() + "\033[0m"}
				case "GREEN":
					return &String{Value: "\033[32m" + args[1].Inspect() + "\033[0m"}
				case "YELLOW":
					return &String{Value: "\033[33m" + args[1].Inspect() + "\033[0m"}
				case "BLUE":
					return &String{Value: "\033[34m" + args[1].Inspect() + "\033[0m"}
				case "MAGENTA":
					return &String{Value: "\033[35m" + args[1].Inspect() + "\033[0m"}
				case "CYAN":
					return &String{Value: "\033[36m" + args[1].Inspect() + "\033[0m"}
				default:
					return newError("color not supported. got=%s", args[0].Type())
				}
			}

			return newError("first argument to `color` not supported. got=%s", args[0].Type())
		},
	}},
	{"sleep", &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}

			switch arg := args[0].(type) {
			case *Integer:
				time.Sleep(time.Duration(arg.Value) * time.Millisecond)
				return NULL
			default:
				return newError("argument to `sleep` not supported. got=%s", args[0].Type())
			}
		},
	}},
	{"exp", &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}

			switch arg := args[0].(type) {
			case *Integer:
				return &Float{Value: math.Exp(float64(arg.Value))}
			case *Float:
				return &Float{Value: math.Exp(arg.Value)}
			default:
				return newError("argument to `exp` not supported. got=%s", args[0].Type())
			}
		},
	}},
	{"pi", &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 0 {
				return newError("wrong number of arguments. got=%d, want=0", len(args))
			}
			return &Float{Value: math.Pi}
		},
	}},
	{"error", &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}
			s, ok := args[0].(*String)
			if !ok {
				return newError("argument to `error` must be STRING. got=%s", args[0].Type())
			}
			return &Error{Message: s.Value}
		},
	}},
	{"is_error", &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}
			if _, ok := args[0].(*Error); ok {
				return TRUE
			}
			return FALSE
		},
	}},
	{"__panic", &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}
			if s, ok := args[0].(*String); ok {
				return &Error{Message: s.Value}
			}
			if e, ok := args[0].(*Error); ok {
				return e
			}
			return newError("argument to `__panic` must be STRING. got=%s", args[0].Type())
		},
	}},
	{"__print", &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}
			fmt.Print(args[0].Inspect())
			return NULL
		},
	}},
	{"__eprint", &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}
			fmt.Fprint(os.Stderr, args[0].Inspect())
			return NULL
		},
	}},
	{"__sprintf", &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 2 {
				return newError("wrong number of arguments. got=%d, want=2", len(args))
			}
			format, ok := args[0].(*String)
			if !ok {
				return newError("first argument to `__sprintf` must be STRING. got=%s", args[0].Type())
			}
			arr, ok := args[1].(*Array)
			if !ok {
				return newError("second argument to `__sprintf` must be ARRAY. got=%s", args[1].Type())
			}
			valArgs := make([]any, len(arr.Elements))
			for i, e := range arr.Elements {
				valArgs[i] = objectToValue(e)
			}
			out := fmt.Sprintf(format.Value, valArgs...)
			if strings.Contains(out, "%!") {
				return newError("printf format error: verb/type mismatch or missing argument in %q -> %q", format.Value, out)
			}
			return &String{Value: out}
		},
	}},
	{"__read_line", &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) > 1 {
				return newError("wrong number of arguments. got=%d, want=0 or 1", len(args))
			}
			if len(args) == 1 {
				fmt.Print(args[0].Inspect())
			}
			reader := getStdinReader()
			text, _ := reader.ReadString('\n')
			text = strings.TrimSpace(text)
			return &String{Value: text}
		},
	}},
	{"__str_split", &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 2 {
				return newError("wrong number of arguments. got=%d, want=2", len(args))
			}
			s, ok := args[0].(*String)
			if !ok {
				return newError("first argument to `__str_split` must be STRING. got=%s", args[0].Type())
			}
			sep, ok := args[1].(*String)
			if !ok {
				return newError("second argument to `__str_split` must be STRING. got=%s", args[1].Type())
			}
			parts := strings.Split(s.Value, sep.Value)
			elems := make([]Object, len(parts))
			for i, p := range parts {
				elems[i] = &String{Value: p}
			}
			return &Array{Elements: elems}
		},
	}},
	{"__str_join", &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 2 {
				return newError("wrong number of arguments. got=%d, want=2", len(args))
			}
			arr, ok := args[0].(*Array)
			if !ok {
				return newError("first argument to `__str_join` must be ARRAY. got=%s", args[0].Type())
			}
			sep, ok := args[1].(*String)
			if !ok {
				return newError("second argument to `__str_join` must be STRING. got=%s", args[1].Type())
			}
			parts := make([]string, len(arr.Elements))
			for i, e := range arr.Elements {
				if s, ok := e.(*String); ok {
					parts[i] = s.Value
				} else {
					parts[i] = e.Inspect()
				}
			}
			return &String{Value: strings.Join(parts, sep.Value)}
		},
	}},
	{"__str_index", &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 2 {
				return newError("wrong number of arguments. got=%d, want=2", len(args))
			}
			s, ok := args[0].(*String)
			if !ok {
				return newError("first argument to `__str_index` must be STRING. got=%s", args[0].Type())
			}
			sub, ok := args[1].(*String)
			if !ok {
				return newError("second argument to `__str_index` must be STRING. got=%s", args[1].Type())
			}
			return NewInteger(int64(strings.Index(s.Value, sub.Value)))
		},
	}},
	{"__str_slice", &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 3 {
				return newError("wrong number of arguments. got=%d, want=3", len(args))
			}
			s, ok := args[0].(*String)
			if !ok {
				return newError("first argument to `__str_slice` must be STRING. got=%s", args[0].Type())
			}
			start, ok := args[1].(*Integer)
			if !ok {
				return newError("second argument to `__str_slice` must be INTEGER. got=%s", args[1].Type())
			}
			end, ok := args[2].(*Integer)
			if !ok {
				return newError("third argument to `__str_slice` must be INTEGER. got=%s", args[2].Type())
			}
			if start.Value < 0 || end.Value < 0 {
				return newError("string slice indices must be non-negative. got=%d, %d", start.Value, end.Value)
			}
			l := int64(len(s.Value))
			st := start.Value
			en := end.Value
			if st > l {
				st = l
			}
			if en > l {
				en = l
			}
			if st > en {
				return newError("string slice start after end. got=%d > %d", st, en)
			}
			return &String{Value: s.Value[int(st):int(en)]}
		},
	}},
	{"__ord", &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}
			s, ok := args[0].(*String)
			if !ok {
				return newError("argument to `__ord` must be STRING. got=%s", args[0].Type())
			}
			if s.Value == "" {
				return newError("argument to `__ord` must be non-empty string")
			}
			r, _ := utf8.DecodeRuneInString(s.Value)
			return NewInteger(int64(r))
		},
	}},
	{"__chr", &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}
			n, ok := args[0].(*Integer)
			if !ok {
				return newError("argument to `__chr` must be INTEGER. got=%s", args[0].Type())
			}
			if n.Value < 0 || n.Value > 0x10FFFF {
				return newError("argument to `__chr` out of range. got=%d", n.Value)
			}
			return &String{Value: string(rune(n.Value))}
		},
	}},
	{"__trim", &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}
			s, ok := args[0].(*String)
			if !ok {
				return newError("argument to `__trim` must be STRING. got=%s", args[0].Type())
			}
			return &String{Value: strings.TrimSpace(s.Value)}
		},
	}},
	{"__upper", &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}
			s, ok := args[0].(*String)
			if !ok {
				return newError("argument to `__upper` must be STRING. got=%s", args[0].Type())
			}
			return &String{Value: strings.ToUpper(s.Value)}
		},
	}},
	{"__lower", &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}
			s, ok := args[0].(*String)
			if !ok {
				return newError("argument to `__lower` must be STRING. got=%s", args[0].Type())
			}
			return &String{Value: strings.ToLower(s.Value)}
		},
	}},
	{"__replace", &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 3 {
				return newError("wrong number of arguments. got=%d, want=3", len(args))
			}
			s, ok := args[0].(*String)
			if !ok {
				return newError("first argument to `__replace` must be STRING. got=%s", args[0].Type())
			}
			old, ok := args[1].(*String)
			if !ok {
				return newError("second argument to `__replace` must be STRING. got=%s", args[1].Type())
			}
			newS, ok := args[2].(*String)
			if !ok {
				return newError("third argument to `__replace` must be STRING. got=%s", args[2].Type())
			}
			return &String{Value: strings.ReplaceAll(s.Value, old.Value, newS.Value)}
		},
	}},
	{"__str_repeat", &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 2 {
				return newError("wrong number of arguments. got=%d, want=2", len(args))
			}
			s, ok := args[0].(*String)
			if !ok {
				return newError("first argument to `__str_repeat` must be STRING. got=%s", args[0].Type())
			}
			n, ok := args[1].(*Integer)
			if !ok {
				return newError("second argument to `__str_repeat` must be INTEGER. got=%s", args[1].Type())
			}
			if n.Value < 0 {
				return newError("repeat count must be non-negative. got=%d", n.Value)
			}
			const maxRepeat int64 = 10000000
			if int64(len(s.Value))*n.Value > maxRepeat {
				return newError("repeat result too large")
			}
			return &String{Value: strings.Repeat(s.Value, int(n.Value))}
		},
	}},
	{"__arr_slice", &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 3 {
				return newError("wrong number of arguments. got=%d, want=3", len(args))
			}
			arr, ok := args[0].(*Array)
			if !ok {
				return newError("first argument to `__arr_slice` must be ARRAY. got=%s", args[0].Type())
			}
			start, ok := args[1].(*Integer)
			if !ok {
				return newError("second argument to `__arr_slice` must be INTEGER. got=%s", args[1].Type())
			}
			end, ok := args[2].(*Integer)
			if !ok {
				return newError("third argument to `__arr_slice` must be INTEGER. got=%s", args[2].Type())
			}
			if start.Value < 0 || end.Value < 0 {
				return newError("array slice indices must be non-negative. got=%d, %d", start.Value, end.Value)
			}
			l := int64(len(arr.Elements))
			st := start.Value
			en := end.Value
			if st > l {
				st = l
			}
			if en > l {
				en = l
			}
			if st > en {
				return newError("array slice start after end. got=%d > %d", st, en)
			}
			newElems := make([]Object, en-st)
			copy(newElems, arr.Elements[int(st):int(en)])
			return &Array{Elements: newElems}
		},
	}},
	{"__arr_sort", &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}
			arr, ok := args[0].(*Array)
			if !ok {
				return newError("argument to `__arr_sort` must be ARRAY. got=%s", args[0].Type())
			}
			n := len(arr.Elements)
			newElems := make([]Object, n)
			copy(newElems, arr.Elements)
			if n <= 1 {
				return &Array{Elements: newElems}
			}
			hasNum := false
			hasStr := false
			hasOther := false
			for _, e := range newElems {
				switch e.(type) {
				case *Integer, *Float:
					hasNum = true
				case *String:
					hasStr = true
				default:
					hasOther = true
				}
			}
			if hasOther || (hasNum && hasStr) {
				return newError("cannot sort mixed or unsupported array types")
			}
			if hasNum {
				sort.Slice(newElems, func(i, j int) bool {
					return toFloatVal(newElems[i]) < toFloatVal(newElems[j])
				})
			} else if hasStr {
				sort.Slice(newElems, func(i, j int) bool {
					return newElems[i].(*String).Value < newElems[j].(*String).Value
				})
			} else {
				return newError("cannot sort array of type %s", newElems[0].Type())
			}
			return &Array{Elements: newElems}
		},
	}},
	{"__arr_join", &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 2 {
				return newError("wrong number of arguments. got=%d, want=2", len(args))
			}
			arr, ok := args[0].(*Array)
			if !ok {
				return newError("first argument to `__arr_join` must be ARRAY. got=%s", args[0].Type())
			}
			sep, ok := args[1].(*String)
			if !ok {
				return newError("second argument to `__arr_join` must be STRING. got=%s", args[1].Type())
			}
			parts := make([]string, len(arr.Elements))
			for i, e := range arr.Elements {
				parts[i] = e.Inspect()
			}
			return &String{Value: strings.Join(parts, sep.Value)}
		},
	}},
	{"__bytes", &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}
			switch v := args[0].(type) {
			case *String:
				cp := make([]byte, len(v.Value))
				copy(cp, v.Value)
				return &Bytes{Value: cp}
			case *Bytes:
				cp := make([]byte, len(v.Value))
				copy(cp, v.Value)
				return &Bytes{Value: cp}
			default:
				return newError("argument to `__bytes` must be STRING. got=%s", args[0].Type())
			}
		},
	}},
	{"__bytes_to_str", &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}
			b, ok := args[0].(*Bytes)
			if !ok {
				return newError("argument to `__bytes_to_str` must be BYTES. got=%s", args[0].Type())
			}
			return &String{Value: string(b.Value)}
		},
	}},
	{"__bytes_len", &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}
			b, ok := args[0].(*Bytes)
			if !ok {
				return newError("argument to `__bytes_len` must be BYTES. got=%s", args[0].Type())
			}
			return NewInteger(int64(len(b.Value)))
		},
	}},
	{"__args", &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 0 {
				return newError("wrong number of arguments. got=%d, want=0", len(args))
			}
			cli := os.Args[1:]
			elems := make([]Object, len(cli))
			for i, a := range cli {
				elems[i] = &String{Value: a}
			}
			return &Array{Elements: elems}
		},
	}},
	{"__env_get", &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}
			k, ok := args[0].(*String)
			if !ok {
				return newError("argument to `__env_get` must be STRING. got=%s", args[0].Type())
			}
			if v, ok := os.LookupEnv(k.Value); ok {
				return &String{Value: v}
			}
			return NULL
		},
	}},
	{"__env_set", &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 2 {
				return newError("wrong number of arguments. got=%d, want=2", len(args))
			}
			k, ok := args[0].(*String)
			if !ok {
				return newError("first argument to `__env_set` must be STRING. got=%s", args[0].Type())
			}
			v, ok := args[1].(*String)
			if !ok {
				return newError("second argument to `__env_set` must be STRING. got=%s", args[1].Type())
			}
			if err := os.Setenv(k.Value, v.Value); err != nil {
				return newError("could not set env %q: %s", k.Value, err.Error())
			}
			return NULL
		},
	}},
	{"__exit", &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}
			n, ok := args[0].(*Integer)
			if !ok {
				return newError("argument to `__exit` must be INTEGER. got=%s", args[0].Type())
			}
			os.Exit(int(n.Value))
			return NULL
		},
	}},
	{"__cwd", &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 0 {
				return newError("wrong number of arguments. got=%d, want=0", len(args))
			}
			dir, err := os.Getwd()
			if err != nil {
				return newError("could not get cwd: %s", err.Error())
			}
			return &String{Value: dir}
		},
	}},
	{"__read_file", &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}
			p, ok := args[0].(*String)
			if !ok {
				return newError("argument to `__read_file` must be STRING. got=%s", args[0].Type())
			}
			clean := filepath.Clean(p.Value)
			data, err := os.ReadFile(clean)
			if err != nil {
				return newError("could not read file %q: %s", p.Value, err.Error())
			}
			return &String{Value: string(data)}
		},
	}},
	{"__write_file", &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 2 {
				return newError("wrong number of arguments. got=%d, want=2", len(args))
			}
			p, ok := args[0].(*String)
			if !ok {
				return newError("first argument to `__write_file` must be STRING. got=%s", args[0].Type())
			}
			s, ok := args[1].(*String)
			if !ok {
				return newError("second argument to `__write_file` must be STRING. got=%s", args[1].Type())
			}
			clean := filepath.Clean(p.Value)
			if err := os.WriteFile(clean, []byte(s.Value), 0644); err != nil {
				return newError("could not write file %q: %s", p.Value, err.Error())
			}
			return NULL
		},
	}},
	{"__append_file", &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 2 {
				return newError("wrong number of arguments. got=%d, want=2", len(args))
			}
			p, ok := args[0].(*String)
			if !ok {
				return newError("first argument to `__append_file` must be STRING. got=%s", args[0].Type())
			}
			s, ok := args[1].(*String)
			if !ok {
				return newError("second argument to `__append_file` must be STRING. got=%s", args[1].Type())
			}
			clean := filepath.Clean(p.Value)
			f, err := os.OpenFile(clean, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				return newError("could not open file %q: %s", p.Value, err.Error())
			}
			defer f.Close()
			if _, err := f.WriteString(s.Value); err != nil {
				return newError("could not append file %q: %s", p.Value, err.Error())
			}
			return NULL
		},
	}},
	{"__ls", &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}
			p, ok := args[0].(*String)
			if !ok {
				return newError("argument to `__ls` must be STRING. got=%s", args[0].Type())
			}
			clean := filepath.Clean(p.Value)
			entries, err := os.ReadDir(clean)
			if err != nil {
				return newError("could not list dir %q: %s", p.Value, err.Error())
			}
			elems := make([]Object, len(entries))
			for i, e := range entries {
				elems[i] = &String{Value: e.Name()}
			}
			return &Array{Elements: elems}
		},
	}},
	{"__stat", &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}
			p, ok := args[0].(*String)
			if !ok {
				return newError("argument to `__stat` must be STRING. got=%s", args[0].Type())
			}
			clean := filepath.Clean(p.Value)
			info, err := os.Stat(clean)
			if err != nil {
				return newError("could not stat %q: %s", p.Value, err.Error())
			}
			return makeStatHash(info)
		},
	}},
	{"__mkdir", &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}
			p, ok := args[0].(*String)
			if !ok {
				return newError("argument to `__mkdir` must be STRING. got=%s", args[0].Type())
			}
			clean := filepath.Clean(p.Value)
			if err := os.MkdirAll(clean, 0755); err != nil {
				return newError("could not mkdir %q: %s", p.Value, err.Error())
			}
			return NULL
		},
	}},
	{"__rm", &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}
			p, ok := args[0].(*String)
			if !ok {
				return newError("argument to `__rm` must be STRING. got=%s", args[0].Type())
			}
			clean := filepath.Clean(p.Value)
			if err := os.RemoveAll(clean); err != nil {
				return newError("could not remove %q: %s", p.Value, err.Error())
			}
			return NULL
		},
	}},
	{"__exists", &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}
			p, ok := args[0].(*String)
			if !ok {
				return newError("argument to `__exists` must be STRING. got=%s", args[0].Type())
			}
			clean := filepath.Clean(p.Value)
			if _, err := os.Stat(clean); err == nil {
				return TRUE
			}
			return FALSE
		},
	}},
	{"__now_ms", &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 0 {
				return newError("wrong number of arguments. got=%d, want=0", len(args))
			}
			return NewInteger(time.Now().UnixMilli())
		},
	}},
	{"__now_ns", &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 0 {
				return newError("wrong number of arguments. got=%d, want=0", len(args))
			}
			return NewInteger(time.Now().UnixNano())
		},
	}},
	{"__sleep_ms", &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}
			n, ok := args[0].(*Integer)
			if !ok {
				return newError("argument to `__sleep_ms` must be INTEGER. got=%s", args[0].Type())
			}
			time.Sleep(time.Duration(n.Value) * time.Millisecond)
			return NULL
		},
	}},
	{"__json_parse", &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}
			s, ok := args[0].(*String)
			if !ok {
				return newError("argument to `__json_parse` must be STRING. got=%s", args[0].Type())
			}
			var v any
			if err := json.Unmarshal([]byte(s.Value), &v); err != nil {
				return newError("could not parse JSON: %s", err.Error())
			}
			return jsonToObject(v)
		},
	}},
	{"__json_stringify", &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}
			goVal, err := objectToJSON(args[0])
			if err != nil {
				return newError("could not stringify JSON: %s", err.Error())
			}
			data, err := json.Marshal(goVal)
			if err != nil {
				return newError("could not stringify JSON: %s", err.Error())
			}
			return &String{Value: string(data)}
		},
	}},
	{"__b64_encode", &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}
			switch v := args[0].(type) {
			case *String:
				return &String{Value: base64.StdEncoding.EncodeToString([]byte(v.Value))}
			case *Bytes:
				return &String{Value: base64.StdEncoding.EncodeToString(v.Value)}
			default:
				return newError("argument to `__b64_encode` must be STRING. got=%s", args[0].Type())
			}
		},
	}},
	{"__b64_decode", &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}
			s, ok := args[0].(*String)
			if !ok {
				return newError("argument to `__b64_decode` must be STRING. got=%s", args[0].Type())
			}
			data, err := base64.StdEncoding.DecodeString(strings.TrimSpace(s.Value))
			if err != nil {
				return newError("could not decode base64: %s", err.Error())
			}
			return &String{Value: string(data)}
		},
	}},
	{"__hex_encode", &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}
			switch v := args[0].(type) {
			case *String:
				return &String{Value: hex.EncodeToString([]byte(v.Value))}
			case *Bytes:
				return &String{Value: hex.EncodeToString(v.Value)}
			default:
				return newError("argument to `__hex_encode` must be STRING. got=%s", args[0].Type())
			}
		},
	}},
	{"__sha256", &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}
			var data []byte
			switch v := args[0].(type) {
			case *String:
				data = []byte(v.Value)
			case *Bytes:
				data = v.Value
			default:
				return newError("argument to `__sha256` must be STRING. got=%s", args[0].Type())
			}
			sum := sha256.Sum256(data)
			return &String{Value: hex.EncodeToString(sum[:])}
		},
	}},
	{"__seed_rand", &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}
			n, ok := args[0].(*Integer)
			if !ok {
				return newError("argument to `__seed_rand` must be INTEGER. got=%s", args[0].Type())
			}
			rand.Seed(n.Value)
			return NULL
		},
	}},
	{"__rand_float", &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 0 {
				return newError("wrong number of arguments. got=%d, want=0", len(args))
			}
			return &Float{Value: rand.Float64()}
		},
	}},
	{"__rand_int", &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}
			n, ok := args[0].(*Integer)
			if !ok {
				return newError("argument to `__rand_int` must be INTEGER. got=%s", args[0].Type())
			}
			if n.Value <= 0 {
				return newError("argument to `__rand_int` must be positive. got=%d", n.Value)
			}
			return NewInteger(rand.Int63n(n.Value))
		},
	}},
	{"__arr_push", &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 2 {
				return newError("wrong number of arguments. got=%d, want=2", len(args))
			}
			arr, ok := args[0].(*Array)
			if !ok {
				return newError("first argument to `__arr_push` must be ARRAY. got=%s", args[0].Type())
			}
			newElems := make([]Object, len(arr.Elements)+1)
			copy(newElems, arr.Elements)
			newElems[len(arr.Elements)] = args[1]
			return &Array{Elements: newElems}
		},
	}},
	{"__hash_keys", &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}
			h, ok := args[0].(*Hash)
			if !ok {
				return newError("argument to `__hash_keys` must be HASH. got=%s", args[0].Type())
			}
			keys := make([]Object, 0, len(h.Pairs))
			for _, k := range h.Order {
				keys = append(keys, h.Pairs[k].Key)
			}
			return &Array{Elements: keys}
		},
	}},
	{"__hash_vals", &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}
			h, ok := args[0].(*Hash)
			if !ok {
				return newError("argument to `__hash_vals` must be HASH. got=%s", args[0].Type())
			}
			vals := make([]Object, 0, len(h.Pairs))
			for _, k := range h.Order {
				vals = append(vals, h.Pairs[k].Value)
			}
			return &Array{Elements: vals}
		},
	}},
	{"__arr_pop", &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}
			arr, ok := args[0].(*Array)
			if !ok {
				return newError("argument to `__arr_pop` must be ARRAY. got=%s", args[0].Type())
			}
			if len(arr.Elements) == 0 {
				return &Array{Elements: []Object{}}
			}
			newElems := make([]Object, len(arr.Elements)-1)
			copy(newElems, arr.Elements[:len(arr.Elements)-1])
			return &Array{Elements: newElems}
		},
	}},
	{"__hash_has", &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 2 {
				return newError("wrong number of arguments. got=%d, want=2", len(args))
			}
			h, ok := args[0].(*Hash)
			if !ok {
				return newError("first argument to `__hash_has` must be HASH. got=%s", args[0].Type())
			}
			key, ok := args[1].(Hashable)
			if !ok {
				return newError("unusable as hash key: %s", args[1].Type())
			}
			if _, ok := h.Pairs[key.HashKey()]; ok {
				return TRUE
			}
			return FALSE
		},
	}},
	{"__hash_del", &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 2 {
				return newError("wrong number of arguments. got=%d, want=2", len(args))
			}
			h, ok := args[0].(*Hash)
			if !ok {
				return newError("first argument to `__hash_del` must be HASH. got=%s", args[0].Type())
			}
			key, ok := args[1].(Hashable)
			if !ok {
				return newError("unusable as hash key: %s", args[1].Type())
			}
			newPairs := make(map[HashKey]HashPair, len(h.Pairs))
			newOrder := make([]HashKey, 0, len(h.Order))
			for k, v := range h.Pairs {
				newPairs[k] = v
			}
			hashKey := key.HashKey()
			delete(newPairs, hashKey)
			for _, k := range h.Order {
				if k != hashKey {
					newOrder = append(newOrder, k)
				}
			}
			return &Hash{Pairs: newPairs, Order: newOrder}
		},
	}},
	{"__str_starts", &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 2 {
				return newError("wrong number of arguments. got=%d, want=2", len(args))
			}
			s, ok := args[0].(*String)
			if !ok {
				return newError("first argument to `__str_starts` must be STRING. got=%s", args[0].Type())
			}
			prefix, ok := args[1].(*String)
			if !ok {
				return newError("second argument to `__str_starts` must be STRING. got=%s", args[1].Type())
			}
			if strings.HasPrefix(s.Value, prefix.Value) {
				return TRUE
			}
			return FALSE
		},
	}},
	{"__str_ends", &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 2 {
				return newError("wrong number of arguments. got=%d, want=2", len(args))
			}
			s, ok := args[0].(*String)
			if !ok {
				return newError("first argument to `__str_ends` must be STRING. got=%s", args[0].Type())
			}
			suffix, ok := args[1].(*String)
			if !ok {
				return newError("second argument to `__str_ends` must be STRING. got=%s", args[1].Type())
			}
			if strings.HasSuffix(s.Value, suffix.Value) {
				return TRUE
			}
			return FALSE
		},
	}},
	{"__bytes_slice", &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 3 {
				return newError("wrong number of arguments. got=%d, want=3", len(args))
			}
			b, ok := args[0].(*Bytes)
			if !ok {
				return newError("first argument to `__bytes_slice` must be BYTES. got=%s", args[0].Type())
			}
			start, ok := args[1].(*Integer)
			if !ok {
				return newError("second argument to `__bytes_slice` must be INTEGER. got=%s", args[1].Type())
			}
			end, ok := args[2].(*Integer)
			if !ok {
				return newError("third argument to `__bytes_slice` must be INTEGER. got=%s", args[2].Type())
			}
			if start.Value < 0 || end.Value < 0 {
				return newError("bytes slice indices must be non-negative. got=%d, %d", start.Value, end.Value)
			}
			l := int64(len(b.Value))
			st := start.Value
			en := end.Value
			if st > l {
				st = l
			}
			if en > l {
				en = l
			}
			if st > en {
				return newError("bytes slice start after end. got=%d > %d", st, en)
			}
			cp := make([]byte, en-st)
			copy(cp, b.Value[int(st):int(en)])
			return &Bytes{Value: cp}
		},
	}},
	{"__bytes_concat", &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 2 {
				return newError("wrong number of arguments. got=%d, want=2", len(args))
			}
			a, ok := args[0].(*Bytes)
			if !ok {
				return newError("first argument to `__bytes_concat` must be BYTES. got=%s", args[0].Type())
			}
			b, ok := args[1].(*Bytes)
			if !ok {
				return newError("second argument to `__bytes_concat` must be BYTES. got=%s", args[1].Type())
			}
			cp := make([]byte, len(a.Value)+len(b.Value))
			copy(cp, a.Value)
			copy(cp[len(a.Value):], b.Value)
			return &Bytes{Value: cp}
		},
	}},
	{"__env_list", &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 0 {
				return newError("wrong number of arguments. got=%d, want=0", len(args))
			}
			env := os.Environ()
			sort.Strings(env)
			elems := make([]Object, len(env))
			for i, kv := range env {
				elems[i] = &String{Value: kv}
			}
			return &Array{Elements: elems}
		},
	}},
	{"__uuid", &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 0 {
				return newError("wrong number of arguments. got=%d, want=0", len(args))
			}
			b := make([]byte, 16)
			if _, err := crand.Read(b); err != nil {
				return newError("could not generate uuid: %s", err.Error())
			}
			b[6] = (b[6] & 0x0f) | 0x40
			b[8] = (b[8] & 0x3f) | 0x80
			s := hex.EncodeToString(b[0:4]) + "-" + hex.EncodeToString(b[4:6]) + "-" + hex.EncodeToString(b[6:8]) + "-" + hex.EncodeToString(b[8:10]) + "-" + hex.EncodeToString(b[10:16])
			return &String{Value: s}
		},
	}},
	{"__random_bytes", &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}
			n, ok := args[0].(*Integer)
			if !ok {
				return newError("argument to `__random_bytes` must be INTEGER. got=%s", args[0].Type())
			}
			if n.Value < 0 {
				return newError("argument to `__random_bytes` must be non-negative. got=%d", n.Value)
			}
			if n.Value > 1048576 {
				return newError("argument to `__random_bytes` too large. got=%d, want <= 1048576", n.Value)
			}
			buf := make([]byte, int(n.Value))
			if _, err := crand.Read(buf); err != nil {
				return newError("could not generate random bytes: %s", err.Error())
			}
			cp := make([]byte, len(buf))
			copy(cp, buf)
			return &Bytes{Value: cp}
		},
	}},
	{"__file_open", &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 2 {
				return newError("wrong number of arguments. got=%d, want=2", len(args))
			}
			p, ok := args[0].(*String)
			if !ok {
				return newError("first argument to `__file_open` must be STRING. got=%s", args[0].Type())
			}
			m, ok := args[1].(*String)
			if !ok {
				return newError("second argument to `__file_open` must be STRING. got=%s", args[1].Type())
			}
			if m.Value != "r" && m.Value != "w" && m.Value != "a" {
				return newError("second argument to `__file_open` must be one of \"r\", \"w\", \"a\". got=%q", m.Value)
			}
			clean := filepath.Clean(p.Value)
			var f *os.File
			var err error
			switch m.Value {
			case "r":
				f, err = os.Open(clean)
			case "w":
				f, err = os.Create(clean)
			case "a":
				f, err = os.OpenFile(clean, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			}
			if err != nil {
				return newError("could not open file %q: %s", p.Value, err.Error())
			}
			return &File{F: f, Path: clean, Mode: m.Value}
		},
	}},
	{"__file_read", &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 2 {
				return newError("wrong number of arguments. got=%d, want=2", len(args))
			}
			f, ok := args[0].(*File)
			if !ok {
				return newError("first argument to `__file_read` must be FILE. got=%s", args[0].Type())
			}
			n, ok := args[1].(*Integer)
			if !ok {
				return newError("second argument to `__file_read` must be INTEGER. got=%s", args[1].Type())
			}
			if n.Value < 0 {
				return newError("second argument to `__file_read` must be non-negative. got=%d", n.Value)
			}
			if n.Value > 1048576 {
				return newError("second argument to `__file_read` too large. got=%d, want <= 1048576", n.Value)
			}
			if f.F == nil {
				return newError("could not read file %q: file is closed", f.Path)
			}
			if n.Value == 0 {
				return &String{Value: ""}
			}
			buf := make([]byte, int(n.Value))
			m, err := io.ReadFull(f.F, buf)
			if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
				return newError("could not read file %q: %s", f.Path, err.Error())
			}
			return &String{Value: string(buf[:m])}
		},
	}},
	{"__file_write", &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 2 {
				return newError("wrong number of arguments. got=%d, want=2", len(args))
			}
			f, ok := args[0].(*File)
			if !ok {
				return newError("first argument to `__file_write` must be FILE. got=%s", args[0].Type())
			}
			s, ok := args[1].(*String)
			if !ok {
				return newError("second argument to `__file_write` must be STRING. got=%s", args[1].Type())
			}
			if f.F == nil {
				return newError("could not write file %q: file is closed", f.Path)
			}
			n, err := f.F.WriteString(s.Value)
			if err != nil {
				return newError("could not write file %q: %s", f.Path, err.Error())
			}
			return NewInteger(int64(n))
		},
	}},
	{"__file_close", &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}
			f, ok := args[0].(*File)
			if !ok {
				return newError("argument to `__file_close` must be FILE. got=%s", args[0].Type())
			}
			if f.F == nil {
				return newError("could not close file %q: file is already closed", f.Path)
			}
			err := f.F.Close()
			f.F = nil
			if err != nil {
				return newError("could not close file %q: %s", f.Path, err.Error())
			}
			return NULL
		},
	}},
	{"__http_get", &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 2 {
				return newError("wrong number of arguments. got=%d, want=2", len(args))
			}
			u, ok := args[0].(*String)
			if !ok {
				return newError("first argument to `__http_get` must be STRING. got=%s", args[0].Type())
			}
			ms, ok := args[1].(*Integer)
			if !ok {
				return newError("second argument to `__http_get` must be INTEGER. got=%s", args[1].Type())
			}
			if ms.Value < 1 || ms.Value > 60000 {
				return newError("second argument to `__http_get` must be between 1 and 60000. got=%d", ms.Value)
			}
			client := &http.Client{Timeout: time.Duration(ms.Value) * time.Millisecond}
			resp, err := client.Get(u.Value)
			if err != nil {
				return newError("http get %q failed: %s", u.Value, err.Error())
			}
			defer resp.Body.Close()
			if resp.StatusCode < 200 || resp.StatusCode >= 300 {
				return newError("http get %q failed with status %s", u.Value, resp.Status)
			}
			if resp.ContentLength > maxHTTPBody {
				return newError("http response body too large")
			}
			data, err := io.ReadAll(io.LimitReader(resp.Body, maxHTTPBody+1))
			if err != nil {
				return newError("http get %q failed: %s", u.Value, err.Error())
			}
			if int64(len(data)) > maxHTTPBody {
				return newError("http response body too large")
			}
			return &String{Value: string(data)}
		},
	}},
	{"__http_post", &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 3 {
				return newError("wrong number of arguments. got=%d, want=3", len(args))
			}
			u, ok := args[0].(*String)
			if !ok {
				return newError("first argument to `__http_post` must be STRING. got=%s", args[0].Type())
			}
			body, ok := args[1].(*String)
			if !ok {
				return newError("second argument to `__http_post` must be STRING. got=%s", args[1].Type())
			}
			ms, ok := args[2].(*Integer)
			if !ok {
				return newError("third argument to `__http_post` must be INTEGER. got=%s", args[2].Type())
			}
			if ms.Value < 1 || ms.Value > 60000 {
				return newError("third argument to `__http_post` must be between 1 and 60000. got=%d", ms.Value)
			}
			client := &http.Client{Timeout: time.Duration(ms.Value) * time.Millisecond}
			resp, err := client.Post(u.Value, "text/plain", strings.NewReader(body.Value))
			if err != nil {
				return newError("http post %q failed: %s", u.Value, err.Error())
			}
			defer resp.Body.Close()
			if resp.StatusCode < 200 || resp.StatusCode >= 300 {
				return newError("http post %q failed with status %s", u.Value, resp.Status)
			}
			if resp.ContentLength > maxHTTPBody {
				return newError("http response body too large")
			}
			data, err := io.ReadAll(io.LimitReader(resp.Body, maxHTTPBody+1))
			if err != nil {
				return newError("http post %q failed: %s", u.Value, err.Error())
			}
			if int64(len(data)) > maxHTTPBody {
				return newError("http response body too large")
			}
			return &String{Value: string(data)}
		},
	}},
	{"__term_width", &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 0 {
				return newError("wrong number of arguments. got=%d, want=0", len(args))
			}
			if n := termWinsizeDim(true); n > 0 {
				return NewInteger(int64(n))
			}
			if v := os.Getenv("COLUMNS"); v != "" {
				if n, err := strconv.Atoi(strings.TrimSpace(v)); err == nil && n > 0 {
					return NewInteger(int64(n))
				}
			}
			return newError("could not determine terminal width: not a TTY")
		},
	}},
	{"__term_height", &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 0 {
				return newError("wrong number of arguments. got=%d, want=0", len(args))
			}
			if n := termWinsizeDim(false); n > 0 {
				return NewInteger(int64(n))
			}
			if v := os.Getenv("LINES"); v != "" {
				if n, err := strconv.Atoi(strings.TrimSpace(v)); err == nil && n > 0 {
					return NewInteger(int64(n))
				}
			}
			return newError("could not determine terminal height: not a TTY")
		},
	}},
	// NOTE: __read_key and __has_input read the raw stdin fd directly,
	// bypassing getStdinReader()'s buffered bufio.Reader, so buffered line
	// input and raw key polling never share (or deadlock on) one buffer.
	{"__read_key", &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 0 {
				return newError("wrong number of arguments. got=%d, want=0", len(args))
			}
			fd := int(os.Stdin.Fd())
			var orig [termiosSize]byte
			if !termGetAttr(uintptr(fd), &orig) {
				// Not a TTY (e.g. piped stdin in tests): plain single-byte read.
				var b [1]byte
				n, err := syscall.Read(fd, b[:])
				if err != nil {
					return newError("could not read key: %s", err.Error())
				}
				if n == 0 {
					return newError("could not read key: EOF")
				}
				return &String{Value: string(b[:n])}
			}
			raw := orig
			termSetLflag(&raw, termGetLflag(&raw)&^(termICANON|termECHO))
			raw[termVMIN] = 1
			raw[termVTIME] = 0
			if !termSetAttr(uintptr(fd), &raw) {
				return newError("could not set terminal to raw mode")
			}
			defer termSetAttr(uintptr(fd), &orig)
			var b [1]byte
			n, err := syscall.Read(fd, b[:])
			if err != nil {
				return newError("could not read key: %s", err.Error())
			}
			if n == 0 {
				return newError("could not read key: EOF")
			}
			if b[0] != 0x1B {
				return &String{Value: string(b[:1])}
			}
			// ESC: allow 0.1s per follow-up byte for escape sequences; read up to 2 more.
			timed := raw
			timed[termVMIN] = 0
			timed[termVTIME] = 1
			termSetAttr(uintptr(fd), &timed)
			seq := []byte{b[0]}
			for i := 0; i < 2; i++ {
				var eb [1]byte
				m, err := syscall.Read(fd, eb[:])
				if err != nil || m == 0 {
					break
				}
				seq = append(seq, eb[0])
			}
			return &String{Value: string(seq)}
		},
	}},
	{"__has_input", &Builtin{
		Fn: func(args ...Object) Object {
			if len(args) != 0 {
				return newError("wrong number of arguments. got=%d, want=0", len(args))
			}
			fd := int(os.Stdin.Fd())
			var rfds syscall.FdSet
			rfds.Bits[fd/64] |= int64(1) << (uint(fd) % 64)
			tv := syscall.Timeval{Sec: 0, Usec: 0}
			n, err := syscall.Select(fd+1, &rfds, nil, nil, &tv)
			if err != nil {
				return newError("could not poll stdin: %s", err.Error())
			}
			if n > 0 {
				return TRUE
			}
			return FALSE
		},
	}},
}

var stdinReader *bufio.Reader

const maxHTTPBody = 5 * 1024 * 1024

// Terminal helpers (Linux only; no new dependencies).
// termios is manipulated as a raw byte buffer to avoid C struct padding
// mismatches. Linux layout: Lflag u32 LE at byte 12, line discipline at 16,
// Cc[0..31] at bytes 17..48 with VTIME=Cc[5] and VMIN=Cc[6].
const (
	termiosSize = 64
	termLflag   = 12
	termVTIME   = 22
	termVMIN    = 23
	termICANON  = 0x0002
	termECHO    = 0x0008
)

type termWinsize struct {
	Row    uint16
	Col    uint16
	Xpixel uint16
	Ypixel uint16
}

func termWinsizeTry(fd uintptr) (termWinsize, bool) {
	var ws termWinsize
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, fd, uintptr(syscall.TIOCGWINSZ), uintptr(unsafe.Pointer(&ws)))
	if errno != 0 {
		return ws, false
	}
	return ws, true
}

// termWinsizeDim returns terminal cols (width=true) or rows via ioctl on
// stdout then stdin, or 0 when unavailable.
func termWinsizeDim(width bool) int {
	for _, f := range []uintptr{os.Stdout.Fd(), os.Stdin.Fd()} {
		if ws, ok := termWinsizeTry(f); ok {
			if width && ws.Col > 0 {
				return int(ws.Col)
			}
			if !width && ws.Row > 0 {
				return int(ws.Row)
			}
		}
	}
	return 0
}

func termGetLflag(b *[termiosSize]byte) uint32 {
	return uint32(b[termLflag]) | uint32(b[termLflag+1])<<8 | uint32(b[termLflag+2])<<16 | uint32(b[termLflag+3])<<24
}

func termSetLflag(b *[termiosSize]byte, v uint32) {
	b[termLflag] = byte(v)
	b[termLflag+1] = byte(v >> 8)
	b[termLflag+2] = byte(v >> 16)
	b[termLflag+3] = byte(v >> 24)
}

func termGetAttr(fd uintptr, buf *[termiosSize]byte) bool {
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, fd, uintptr(syscall.TCGETS), uintptr(unsafe.Pointer(buf)))
	return errno == 0
}

func termSetAttr(fd uintptr, buf *[termiosSize]byte) bool {
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, fd, uintptr(syscall.TCSETS), uintptr(unsafe.Pointer(buf)))
	return errno == 0
}

func getStdinReader() *bufio.Reader {
	if stdinReader == nil {
		stdinReader = bufio.NewReader(os.Stdin)
	}
	return stdinReader
}

func toInt(o Object) (int, bool) {
	switch v := o.(type) {
	case *Integer:
		return int(v.Value), true
	case *Float:
		return int(v.Value), true
	}
	return 0, false
}

func objectToValue(object Object) any {
	switch val := object.(type) {
	case *Integer:
		return val.Value
	case *Float:
		return val.Value
	case *String:
		return val.Value
	case *Boolean:
		return val.Value
	default:
		return val.Inspect()
	}

}

func toFloatVal(o Object) float64 {
	switch v := o.(type) {
	case *Integer:
		return float64(v.Value)
	case *Float:
		return v.Value
	}
	return 0
}

func makeStatHash(info os.FileInfo) *Hash {
	pairs := make(map[HashKey]HashPair, 5)
	order := make([]HashKey, 0, 5)
	add := func(k string, v Object) {
		ko := &String{Value: k}
		hk := ko.HashKey()
		pairs[hk] = HashPair{Key: ko, Value: v}
		order = append(order, hk)
	}
	var isDir Object = FALSE
	if info.IsDir() {
		isDir = TRUE
	}
	add("size", NewInteger(info.Size()))
	add("is_dir", isDir)
	add("mode", &String{Value: info.Mode().String()})
	add("name", &String{Value: info.Name()})
	add("mtime", NewInteger(info.ModTime().Unix()))
	return &Hash{Pairs: pairs, Order: order}
}

func jsonToObject(v any) Object {
	switch val := v.(type) {
	case nil:
		return NULL
	case bool:
		if val {
			return TRUE
		}
		return FALSE
	case string:
		return &String{Value: val}
	case float64:
		if math.Trunc(val) == val {
			iv := int64(val)
			if float64(iv) == val {
				return NewInteger(iv)
			}
		}
		return &Float{Value: val}
	case []any:
		elems := make([]Object, len(val))
		for i, e := range val {
			elems[i] = jsonToObject(e)
		}
		return &Array{Elements: elems}
	case map[string]any:
		keys := make([]string, 0, len(val))
		for k := range val {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		pairs := make(map[HashKey]HashPair, len(val))
		order := make([]HashKey, 0, len(val))
		for _, k := range keys {
			ko := &String{Value: k}
			hk := ko.HashKey()
			pairs[hk] = HashPair{Key: ko, Value: jsonToObject(val[k])}
			order = append(order, hk)
		}
		return &Hash{Pairs: pairs, Order: order}
	default:
		return newError("unsupported JSON value type %T", v)
	}
}

func objectToJSON(o Object) (any, error) {
	switch v := o.(type) {
	case *Integer:
		return v.Value, nil
	case *Float:
		return v.Value, nil
	case *String:
		return v.Value, nil
	case *Boolean:
		return v.Value, nil
	case *Null:
		return nil, nil
	case *Bytes:
		return string(v.Value), nil
	case *Array:
		arr := make([]any, len(v.Elements))
		for i, e := range v.Elements {
			jv, err := objectToJSON(e)
			if err != nil {
				return nil, err
			}
			arr[i] = jv
		}
		return arr, nil
	case *Hash:
		m := make(map[string]any, len(v.Pairs))
		if len(v.Order) == 0 && len(v.Pairs) > 0 {
			for _, pair := range v.Pairs {
				keyStr := pair.Key.Inspect()
				jv, err := objectToJSON(pair.Value)
				if err != nil {
					return nil, err
				}
				m[keyStr] = jv
			}
		} else {
			for _, hk := range v.Order {
				pair := v.Pairs[hk]
				keyStr := pair.Key.Inspect()
				jv, err := objectToJSON(pair.Value)
				if err != nil {
					return nil, err
				}
				m[keyStr] = jv
			}
		}
		return m, nil
	default:
		return nil, fmt.Errorf("unsupported type for JSON: %s", o.Type())
	}
}

func newError(format string, a ...any) *Error {
	return &Error{Message: fmt.Sprintf(format, a...)}
}

func GetBuiltinByName(name string) *Builtin {
	// O(1) map lookup; linear scan kept as fallback for compat
	// (e.g. if Builtins were appended to after init).
	if b, ok := builtinIndex[name]; ok {
		return b
	}
	for _, def := range Builtins {
		if def.Name == name {
			return def.Builtin
		}
	}
	return nil
}

var builtinIndex = func() map[string]*Builtin {
	m := make(map[string]*Builtin, len(Builtins))
	for _, def := range Builtins {
		m[def.Name] = def.Builtin
	}
	return m
}()
