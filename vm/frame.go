package vm

import (
	"github.com/theawakener0/Zod/code"
	obj "github.com/theawakener0/Zod/object"
)

type Frame struct {
	cl				*obj.Closure
	ip 				int
	basePointer 	int
	globals []obj.Object
	instructions code.Instructions
}

func NewFrame(cl *obj.Closure, basePointer int, globals []obj.Object) *Frame {
	return &Frame{
		cl: cl,
		ip: -1,
		basePointer: basePointer,
		globals: globals,
		instructions: cl.Fn.Instructions,
	}
}

func (f *Frame) Instructions() code.Instructions {
	return f.instructions
}
