package compiler

import (
	"fmt"

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

func (c *Compiler) Compile (node ast.Node) error {
	switch node := node.(type) {
	case *ast.Program:
		for _, s := range node.Statements {
			err := c.Compile(s)
			if err != nil {
				return err
			}
		}
	case *ast.ExpressionStatement:
		err := c.Compile(node.Expression)
		if err != nil {
			return err
		}
		c.emit(code.OpPop)
	case *ast.InfixExpression:
		if node.Opt == "<" {
			err0 := c.Compile(node.Right)
			if err0 != nil {
				return err0
			}

			err1 := c.Compile(node.Left)
			if err1 != nil {
				return err1
			}

			c.emit(code.OpGreaterThan)
			return nil
		}
		if node.Opt == "<=" {
			err0 := c.Compile(node.Right)
			if err0 != nil {
				return err0
			}

			err1 := c.Compile(node.Left)
			if err1 != nil {
				return err1
			}

			c.emit(code.OpGreaterThanEqual)
			return nil
		}
		err0 := c.Compile(node.Left)
		if err0 != nil {
			return err0
		}

		err1 := c.Compile(node.Right)
		if err1 != nil {
			return err1
		}

		switch node.Opt {
		case "+":
			c.emit(code.OpAdd)
		case "-":
			c.emit(code.OpSub)
		case "*":
			c.emit(code.OpMul)
		case "/":
			c.emit(code.OpDiv)
		case ">":
			c.emit(code.OpGreaterThan)
		case "==":
			c.emit(code.OpEqual)
		case "!=":
			c.emit(code.OpNotEqual)
		case ">=":
			c.emit(code.OpGreaterThanEqual)
		default:
			return fmt.Errorf("unknown operator %s", node.Opt)
		}
	case *ast.PrefixExpression:
		err := c.Compile(node.Right)
		if err != nil {
			return err
		}

		switch node.Opt {
		case "!":
			c.emit(code.OpBang)
		case "-":
			c.emit(code.OpMinus)
		default:
			return fmt.Errorf("unknown operator %s", node.Opt)
		}
	case *ast.IntegerLiteral:
		integer := &obj.Integer{Value: node.Value}
		c.emit(code.OpConstant, c.addConstant(integer))
	case *ast.FloatLiteral:
		float := &obj.Float{Value: node.Value}
		c.emit(code.OpConstant, c.addConstant(float))
	case *ast.Boolean:
		if node.Value {
			c.emit(code.OpTrue)
		} else {
			c.emit(code.OpFalse)
		}
	}
	return nil
}

func (c *Compiler) Bytecode() *Bytecode {
	return &Bytecode{
		Instructions: c.instructions,
		Constant: c.constant,
	}
}

func (c *Compiler) addConstant(obj obj.Object) int {
	c.constant = append(c.constant, obj)
	return len(c.constant) - 1
}

func (c *Compiler) addInstruction(ins []byte) int {
	posNewInstruction := len(c.instructions)
	c.instructions = append(c.instructions, ins...)
	return posNewInstruction
}

func (c *Compiler) emit(op code.Opcode, operands ...int) int {
	ins := code.Make(op, operands...)
	pos := c.addInstruction(ins)
	return pos
}


