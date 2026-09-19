package vm

import (
	"github.com/theawakener0/Zod/code"
	obj "github.com/theawakener0/Zod/object"
)

type Frame struct {
	cl				*obj.Closure
	ip 				int
	basePointer 	int
}

func NewFrame(cl *obj.Closure, basePointer int) *Frame {
	return &Frame{
		cl: cl,
		ip: -1,
		basePointer: basePointer,
	}
}

func (f *Frame) Instructions() code.Instructions {
	return f.cl.Fn.Instructions
}
