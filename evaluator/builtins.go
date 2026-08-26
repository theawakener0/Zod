package evaluator

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
	"math"
	"math/rand"

	obj "github.com/theawakener0/Zod/object"
)

var builtins = map[string]*obj.Builtin{
	"len": {
		Fn: func(args ...obj.Object) obj.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}

			switch arg := args[0].(type) {
			case *obj.String:
				return &obj.Integer{Value: int64(len(arg.Value))}
			case *obj.Array:
				return &obj.Integer{Value: int64(len(arg.Elements))}
			case *obj.Matrix:
				return &obj.Integer{Value: int64(arg.Rows)}
			case *obj.Hash:
				return &obj.Integer{Value: int64(len(arg.Pairs))}
			default:
				return newError("argument to `len` not supported. got=%s", args[0].Type())
			}
		},
	},
	"println": {
		Fn: func(args ...obj.Object) obj.Object {
			for _, arg := range args {
				fmt.Println(arg.Inspect())
			}

			return NULL
		},
	},
	"printf": {
		Fn: func(args ...obj.Object) obj.Object {
			if len(args) < 1 {
				return newError("wrong number of arguments. got=%d, want=1 or more", len(args))
			}

			format, ok := args[0].(*obj.String)
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
	},
	"input": {
		Fn: func(args ...obj.Object) obj.Object {
			if len(args) > 1 {
				return newError("wrong number of arguments. got=%d, want=0 or 1", len(args))
			}

			if len(args) == 1 {
				fmt.Print(args[0].Inspect())
			}

			reader := getStdinReader()
			text, _ := reader.ReadString('\n')
			text = strings.TrimSpace(text)

			return &obj.String{Value: text}
		},
	},
	"int": {
		Fn: func(args ...obj.Object) obj.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}

			switch arg := args[0].(type) {
			case *obj.String:
				val, err := strconv.ParseInt(arg.Value, 0, 64)
				if err != nil {
					return newError("could not parse %q as integer", arg.Value)
				}
				return &obj.Integer{Value: val}
			case *obj.Integer:
				return arg
			case *obj.Float:
				return &obj.Integer{Value: int64(arg.Value)}
			case *obj.Boolean:
				if arg.Value {
					return &obj.Integer{Value: 1}
				}
				return &obj.Integer{Value: 0}
			default:
				return newError("argument to `int` not supported. got=%s", args[0].Type())
			}
		},
	},
	"float": {
		Fn: func(args ...obj.Object) obj.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}

			switch arg := args[0].(type) {
			case *obj.String:
				val, err := strconv.ParseFloat(arg.Value, 64)
				if err != nil {
					return newError("could not parse %q as float", arg.Value)
				}
				return &obj.Float{Value: val}
			case *obj.Float:
				return arg
			case *obj.Integer:
				return &obj.Float{Value: float64(arg.Value)}
			case *obj.Boolean:
				if arg.Value {
					return &obj.Float{Value: 1.0}
				}
				return &obj.Float{Value: 0.0}
			default:
				return newError("argument to `float` not supported. got=%s", args[0].Type())
			}
		},
	},
	"string": {
		Fn: func(args ...obj.Object) obj.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}

			switch arg := args[0].(type) {
			case *obj.String:
				return arg
			case *obj.Integer:
				return &obj.String{Value: fmt.Sprintf("%s", arg.Inspect())}
			case *obj.Float:
				return &obj.String{Value: arg.Inspect()}
			case *obj.Boolean:
				if arg.Value {
					return &obj.String{Value: "true"}
				}
				return &obj.String{Value: "false"}
			default:
				return newError("argument to `string` not supported. got=%s", args[0].Type())
			}
		},
	},
	"type": {
		Fn: func(args ...obj.Object) obj.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}
			return &obj.String{Value: fmt.Sprintf("%s", args[0].Type())}
		},
	},
	"first": {
		Fn: func(args ...obj.Object) obj.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}

			switch arg := args[0].(type) {
			case *obj.Array:
				if len(arg.Elements) > 0 {
					return arg.Elements[0]
				}
				return NULL
			default:
				return newError("argument to `first` not supported. got=%s", args[0].Type())
			}
		},
	},
	"last": {
		Fn: func(args ...obj.Object) obj.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}

			switch arg := args[0].(type) {
			case *obj.Array:
				if len(arg.Elements) > 0 {
					return arg.Elements[len(arg.Elements)-1]
				}
				return NULL
			default:
				return newError("argument to `last` not supported. got=%s", args[0].Type())
			}
		},
	},
	"pop": {
		Fn: func(args ...obj.Object) obj.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}
			switch arg := args[0].(type) {
			case *obj.Array:
				length := len(arg.Elements)
				if length > 0 {
					newElements := make([]obj.Object, length-1)
					copy(newElements, arg.Elements[:length-1])
					return &obj.Array{Elements: newElements}
				}
				return NULL
			default:
				return newError("argument to `pop` not supported. got=%s", args[0].Type())
			}
		},
	},
	"push": {
		Fn: func(args ...obj.Object) obj.Object {
			if len(args) != 2 {
				return newError("wrong number of arguments. got=%d, want=2", len(args))
			}

			switch arg := args[0].(type) {
			case *obj.Array:
				length := len(arg.Elements)

				newElements := make([]obj.Object, length+1)
				copy(newElements, arg.Elements)
				newElements[length] = args[1]

				return &obj.Array{Elements: newElements}
			default:
				return newError("argument to `push` not supported. got=%s", args[0].Type())
			}
		},
	},
	"insert": {
		Fn: func(args ...obj.Object) obj.Object {
			if len(args) != 3 {
				return newError("wrong number of arguments. got=%d, want=3", len(args))
			}

			hash, ok := args[0].(*obj.Hash)
			if !ok {
				return newError("argument to `insert` not supported. got=%s", args[0].Type())
			}

			key, ok := args[1].(obj.Hashable)
			if !ok {
				return newError("unusable as hash key: %s", args[1].Type())
			}

			newPairs := make(map[obj.HashKey]obj.HashPair, len(hash.Pairs)+1)
			newOrder := make([]obj.HashKey, 0, len(hash.Order)+1)
			for k, v := range hash.Pairs {
				newPairs[k] = v
			}
			newOrder = append(newOrder, hash.Order...)

			hashKey := key.HashKey()
			if _, exists := newPairs[hashKey]; !exists {
				newOrder = append(newOrder, hashKey)
			}
			newPairs[hashKey] = obj.HashPair{Key: args[1], Value: args[2]}

			return &obj.Hash{Pairs: newPairs, Order: newOrder}
		},
	},
	"remove": {
		Fn: func(args ...obj.Object) obj.Object {
			if len(args) != 2 {
				return newError("wrong number of arguments. got=%d, want=2", len(args))
			}

			hash, ok := args[0].(*obj.Hash)
			if !ok {
				return newError("argument to `remove` not supported. got=%s", args[0].Type())
			}

			key, ok := args[1].(obj.Hashable)
			if !ok {
				return newError("unusable as hash key: %s", args[1].Type())
			}

			newPairs := make(map[obj.HashKey]obj.HashPair, len(hash.Pairs))
			newOrder := make([]obj.HashKey, 0, len(hash.Order))
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

			return &obj.Hash{Pairs: newPairs, Order: newOrder}
		},
	},
	"keys": {
		Fn: func(args ...obj.Object) obj.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}

			hash, ok := args[0].(*obj.Hash)
			if !ok {
				return newError("argument to `keys` not supported. got=%s", args[0].Type())
			}

			keys := make([]obj.Object, 0, len(hash.Pairs))
			for _, k := range hash.Order {
				pair := hash.Pairs[k]
				keys = append(keys, pair.Key)
			}

			return &obj.Array{Elements: keys}
		},
	},
	"vals": {
		Fn: func(args ...obj.Object) obj.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}

			hash, ok := args[0].(*obj.Hash)
			if !ok {
				return newError("argument to `vals` not supported. got=%s", args[0].Type())
			}

			values := make([]obj.Object, 0, len(hash.Pairs))
			for _, k := range hash.Order {
				pair := hash.Pairs[k]
				values = append(values, pair.Value)
			}

			return &obj.Array{Elements: values}
		},
	},
	"contains": {
		Fn: func(args ...obj.Object) obj.Object {
			if len(args) != 2 {
				return newError("wrong number of arguments. got=%d, want=2", len(args))
			}

			hash, ok := args[0].(*obj.Hash)
			if !ok {
				return newError("argument to `contains` not supported. got=%s", args[0].Type())
			}

			key, ok := args[1].(obj.Hashable)
			if !ok {
				return newError("unusable as hash key: %s", args[1].Type())
			}

			if _, ok := hash.Pairs[key.HashKey()]; ok {
				return TRUE
			}
			return FALSE
		},
	},
	"random": {
		Fn: func(args ...obj.Object) obj.Object {
			if len(args) == 0 {
				return &obj.Float{Value: rand.Float64()}
			}
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=0 or 1", len(args))
			}

			switch arg := args[0].(type) {
			case *obj.Integer:
				if arg.Value <= 0 {
					return newError("argument to `random` must be positive. got=%d", arg.Value)
				}
				return &obj.Integer{Value: rand.Int63n(arg.Value)}
			case *obj.Float:
				if arg.Value <= 0 {
					return newError("argument to `random` must be positive. got=%f", arg.Value)
				}
				return &obj.Float{Value: rand.Float64() * arg.Value}
			default:
				return newError("argument to `random` not supported. got=%s", args[0].Type())
			}
		}, 
	},
	"matrix": {
		Fn: func(args ...obj.Object) obj.Object {
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

			val, ok := args[2].(*obj.Array)
			if !ok {
				return newError("third argument to `matrix` not an array. got=%s", args[2].Type())
			}

			if rows < 1 || cols < 1 {
				return newError("matrix dimensions must be positive. got=%d, %d", rows, cols)
			}

			if len(val.Elements) != rows*cols {
				return newError("array length does not match matrix dimensions. got=%d, want=%d", len(val.Elements), rows*cols)
			}

			data := make([][]obj.Object, rows)
			for i := 0; i < rows; i++ {
				row := make([]obj.Object, cols)
				for j := 0; j < cols; j++ {
					e := val.Elements[i*cols+j]
					switch e.(type) {
					case *obj.Integer, *obj.Float:
					default:
						return newError("matrix elements must be integers or floats. got=%s", e.Type())
					}
					row[j] = e
				}
				data[i] = row
			}

			return &obj.Matrix{Rows: rows, Cols: cols, Data: data}
		},
	},
	"make" : {
		Fn: func(args ...obj.Object) obj.Object {
			if len(args) < 1 || len(args) > 2 {
				return newError("wrong number of arguments. got=%d, want=1 or 2", len(args))
			}

			var size int64
			switch args[0].(type) {
			case *obj.Integer:
				size = args[0].(*obj.Integer).Value
			case *obj.Float:
				fVal := args[0].(*obj.Float).Value
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
			
			fill := obj.Object(NULL)
			if len(args) == 2 {
				fill = args[1]
			}
			
			elements := make([]obj.Object, size)
			for i := range elements {
				elements[i] = fill
			}
			
			return &obj.Array{Elements: elements}
		},
	},
	"color": {
		Fn: func(args ...obj.Object) obj.Object {
			if len(args) != 2 {
				return newError("wrong number of arguments. got=%d, want=2", len(args))
			}

			switch arg := args[0].(type) {
			case *obj.String:
				switch arg.Value {
				case "RED":
					return &obj.String{Value: fmt.Sprintf("\033[31m%s\033[0m", args[1].Inspect())}
				case "GREEN":
					return &obj.String{Value: fmt.Sprintf("\033[32m%s\033[0m", args[1].Inspect())}
				case "YELLOW":
					return &obj.String{Value: fmt.Sprintf("\033[33m%s\033[0m", args[1].Inspect())}
				case "BLUE":
					return &obj.String{Value: fmt.Sprintf("\033[34m%s\033[0m", args[1].Inspect())}
				case "MAGENTA":
					return &obj.String{Value: fmt.Sprintf("\033[35m%s\033[0m", args[1].Inspect())}
				case "CYAN":
					return &obj.String{Value: fmt.Sprintf("\033[36m%s\033[0m", args[1].Inspect())}
				default:
					return newError("color not supported. got=%s", args[0].Type())
				}
			}

			return newError("first argument to `color` not supported. got=%s", args[0].Type())
		},
	},
	"sleep": {
		Fn: func(args ...obj.Object) obj.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}

			switch arg := args[0].(type) {
			case *obj.Integer:
				time.Sleep(time.Duration(arg.Value) * time.Millisecond)
				return NULL
			default:
				return newError("argument to `sleep` not supported. got=%s", args[0].Type())
			}
		},
	},
	"exp": {
		Fn: func(args ...obj.Object) obj.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}

			switch arg := args[0].(type) {
			case *obj.Integer:
				return &obj.Float{Value: math.Exp(float64(arg.Value))}
			case *obj.Float:
				return &obj.Float{Value: math.Exp(arg.Value)}
			default:
				return newError("argument to `exp` not supported. got=%s", args[0].Type())
			}
		},
	},
	"pi" : {
		Fn: func(args ...obj.Object) obj.Object {
			return &obj.Float{Value: math.Pi}
		},
	},
}

var stdinReader *bufio.Reader

func getStdinReader() *bufio.Reader {
	if stdinReader == nil {
		stdinReader = bufio.NewReader(os.Stdin)
	}
	return stdinReader
}

func toInt(o obj.Object) (int, bool) {
	switch v := o.(type) {
	case *obj.Integer:
		return int(v.Value), true
	case *obj.Float:
		return int(v.Value), true
	}
	return 0, false
}

func objectToValue(object obj.Object) any {
	switch val := object.(type) {
	case *obj.Integer:
		return val.Value
	case *obj.Float:
		return val.Value
	case *obj.String:
		return val.Value
	case *obj.Boolean:
		return val.Value
	default:
		return val.Inspect()
	}

}
