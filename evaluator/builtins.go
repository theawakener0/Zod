package evaluator

import (
	obj "github.com/theawakener0/Zod/object"
)

func GetBuiltinByName(name string) *obj.Builtin {
	return obj.GetBuiltinByName(name)
}
