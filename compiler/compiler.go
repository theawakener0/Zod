package compiler

import (
	"github.com/theawakener0/Zod/ast"
	"github.com/theawakener0/Zod/code"
	obj "github.com/theawakener0/Zod/object"
)

type Compiler struct {
	instructions 	code.Instructions
	constant 		[]obj.Object
}

type Bytecode struct {
	Instructions 	code.Instructions
	Constant 		[]obj.Object
}

func New() *Compiler {
	 return &Compiler{
		instructions: code.Instructions{},
		constant: []obj.Object{},
	}
}

func (c *Compiler) Compile (node ast.Node) *Bytecode {
	return nil
}

func (c *Compiler) Bytecode() *Bytecode {
	return &Bytecode{
		Instructions: c.instructions,
		Constant: c.constant,
	}
}


