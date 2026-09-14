package vm

import (
	"github.com/theawakener0/Zod/code"
	obj "github.com/theawakener0/Zod/object"
)

type Frame struct {
	fn 				*obj.CompiledFunction
	ip 				int
	basePointer 	int
}

func NewFrame(fn *obj.CompiledFunction, basePointer int) *Frame {
	return &Frame{
		fn: fn,
		ip: -1,
		basePointer: basePointer,
	}
}

func (f *Frame) Instructions() code.Instructions {
	return f.fn.Instructions
}
