package evaluator

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	obj "github.com/theawakener0/zod/object"
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

			fmt.Printf(format.Value, valArgs...)

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

			reader := bufio.NewReader(os.Stdin)
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
			for k, v := range hash.Pairs {
				newPairs[k] = v
			}
			newPairs[key.HashKey()] = obj.HashPair{Key: args[1], Value: args[2]}

			return &obj.Hash{Pairs: newPairs}
		},
	},
	"remove": {
		Fn: func(args ...obj.Object) obj.Object {
			if len(args) != 2 {
				return newError("wrong number of arguments. got=%d, want=2", len(args))
			}

			hash, ok := args[0].(*obj.Hash)
			if !ok {
				return newError("argument to `delete` not supported. got=%s", args[0].Type())
			}

			key, ok := args[1].(obj.Hashable)
			if !ok {
				return newError("unusable as hash key: %s", args[1].Type())
			}

			newPairs := make(map[obj.HashKey]obj.HashPair, len(hash.Pairs))
			for k, v := range hash.Pairs {
				newPairs[k] = v
			}
			delete(newPairs, key.HashKey())

			return &obj.Hash{Pairs: newPairs}
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
			for _, pair := range hash.Pairs {
				keys = append(keys, pair.Key)
			}

			return &obj.Array{Elements: keys}
		},
	},
	"values": {
		Fn: func(args ...obj.Object) obj.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}

			hash, ok := args[0].(*obj.Hash)
			if !ok {
				return newError("argument to `values` not supported. got=%s", args[0].Type())
			}

			values := make([]obj.Object, 0, len(hash.Pairs))
			for _, pair := range hash.Pairs {
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
}

func objectToValue(object obj.Object) any {
	switch val := object.(type) {
	case *obj.Integer:
		return val.Value
	case *obj.String:
		return val.Value
	case *obj.Boolean:
		return val.Value
	default:
		return val.Inspect()
	}

}
