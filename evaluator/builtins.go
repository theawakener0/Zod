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
