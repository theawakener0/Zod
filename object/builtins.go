package object

import (
	"bufio"
	"fmt"
	"math"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"
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
}

var stdinReader *bufio.Reader

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
