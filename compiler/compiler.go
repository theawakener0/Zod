package compiler

import (
	"fmt"
	"sort"

	"github.com/theawakener0/Zod/ast"
	"github.com/theawakener0/Zod/code"
	obj "github.com/theawakener0/Zod/object"
)

type EmittedInstruction struct {
	Opcode 		code.Opcode
	Position 	int
}

type Compiler struct {
	instructions 		code.Instructions
	constant 			[]obj.Object

	lastInstruction 	EmittedInstruction
	previousInstruction EmittedInstruction
	symbolTable 		*SymbolTable
}

type Bytecode struct {
	Instructions 	code.Instructions
	Constant 		[]obj.Object
}

func New() *Compiler {
	 return &Compiler{
		instructions: code.Instructions{},
		constant: []obj.Object{},
		lastInstruction: EmittedInstruction{},
		previousInstruction: EmittedInstruction{},
		symbolTable: NewSymbolTable(),
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
	case *ast.IfExpression:
		err0 := c.Compile(node.Condition)
		if err0 != nil {
			return err0
		}
		
		jumpNotTruthyPos := c.emit(code.OpJumpNotTruthy, 9999)

		err1 := c.Compile(node.Consequence)
		if err1 != nil {
			return err1
		}

		if c.lastInstructionIsPop() {
			c.removeLastPop()
		}

		jumpPos := c.emit(code.OpJump, 9999)

		afterConsequencePos := len(c.instructions)
		c.changeOperand(jumpNotTruthyPos, afterConsequencePos)

		if node.ElseIf != nil {
			err := c.compileIfExpression(node.ElseIf)
			if err != nil {
				return err
			}
		} else if (node.Alternative == nil) {
			c.emit(code.OpNull)
		} else {
			err := c.Compile(node.Alternative)
			if err != nil {
				return err
			}

			if c.lastInstructionIsPop() {
				c.removeLastPop()
			}

		}

		afterAlternativePos := len(c.instructions)
		c.changeOperand(jumpPos, afterAlternativePos)
	case *ast.BlockStatement:
		for _, s := range node.Statements {
			err := c.Compile(s)
			if err != nil {
				return err
			}
		}
	case *ast.LetStatement:
		err := c.Compile(node.Value)
		if err != nil {
			return err
		}
		symbol := c.symbolTable.Define(node.Name.Value)
		c.emit(code.OpSetGlobal, symbol.Index)
	case *ast.AssignStatement:
		err := c.Compile(node.Value)
		if err != nil {
			return err
		}
		ident, ok := node.Left.(*ast.Identifier)
		if !ok {
			return fmt.Errorf("only identifier assignment supported, got %T", node.Left)
		}
		switch node.Token.Literal {
		case ":=":
			symbol := c.symbolTable.DefineIfNotExists(ident.Value)
			c.emit(code.OpSetGlobal, symbol.Index)
		case "=":
			symbol, found := c.symbolTable.Resolve(ident.Value)
			if !found {
				return fmt.Errorf("identifier not found: %s", ident.Value)
			}
			c.emit(code.OpSetGlobal, symbol.Index)
		default:
			return fmt.Errorf("unknown assignment operator %s", node.Token.Literal)
		}
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
	case *ast.StringLiteral:
		str := &obj.String{Value: node.Value}
		c.emit(code.OpConstant, c.addConstant(str))
	case *ast.ArrayLiteral:
		for _, ele := range node.Elements {
			err := c.Compile(ele)
			if err != nil {
				return err
			}
		}
		c.emit(code.OpArray, len(node.Elements))
	case *ast.HashLiteral:
		pairs := make([]ast.HashLiteralPair, len(node.Pairs))
		copy(pairs, node.Pairs)

		sort.Slice(pairs, func(i, j int) bool {
			return pairs[i].Key.String() < pairs[j].Key.String()
		})

		for _, pairs := range pairs {
			err0 := c.Compile(pairs.Key)
			if err0 != nil {
				return err0
			}

			err1 := c.Compile(pairs.Value)
			if err1 != nil {
				return err1
			}
		}
		
		c.emit(code.OpHash, len(node.Pairs)*2)
	case *ast.IndexExpression:
		err0 := c.Compile(node.Left)
		if err0 != nil {
			return err0
		}

		err1 := c.Compile(node.Index)
		if err1 != nil {
			return err1
		}

		c.emit(code.OpIndex)
	case *ast.Boolean:
		if node.Value {
			c.emit(code.OpTrue)
		} else {
			c.emit(code.OpFalse)
		}
	case *ast.Identifier:
		symbol, ok := c.symbolTable.Resolve(node.Value)
		if !ok {
			return fmt.Errorf("undefined variable %s", node.Value)
		}
		c.emit(code.OpGetGlobal, symbol.Index)
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

	c.setLastInstruction(op, pos)

	return pos
}

func (c *Compiler) setLastInstruction(op code.Opcode, pos int) {
	previous := c.lastInstruction
	last := EmittedInstruction{Opcode: op, Position: pos}

	c.previousInstruction = previous
	c.lastInstruction = last
}

func (c *Compiler) lastInstructionIsPop() bool {
	return c.lastInstruction.Opcode == code.OpPop
}

func (c *Compiler) removeLastPop() {
	c.instructions = c.instructions[:c.lastInstruction.Position]
	c.lastInstruction = c.previousInstruction
}

func (c *Compiler) replaceInstruction(pos int, newInstruction []byte) {
	for i := range len(newInstruction) {
		c.instructions[pos+i] = newInstruction[i]
	}
}

func (c *Compiler) changeOperand(opPos int, operand int) {
	op := code.Opcode(c.instructions[opPos])
	newInstruction := code.Make(op, operand)

	c.replaceInstruction(opPos, newInstruction)
}

func (c *Compiler) compileIfExpression(node *ast.IfExpression) error {
	err0 := c.Compile(node.Condition)
	if err0 != nil {
		return err0
	}
	
	jumpNotTruthyPos := c.emit(code.OpJumpNotTruthy, 9999)

	err1 := c.Compile(node.Consequence)
	if err1 != nil {
		return err1
	}

	if c.lastInstructionIsPop() {
		c.removeLastPop()
	}

	jumpPos := c.emit(code.OpJump, 9999)

	afterConsequencePos := len(c.instructions)
	c.changeOperand(jumpNotTruthyPos, afterConsequencePos)

	if node.Alternative == nil {
		c.emit(code.OpNull)
	} else {
		err := c.Compile(node.Alternative)
		if err != nil {
			return err
		}

		if c.lastInstructionIsPop() {
			c.removeLastPop()
		}

	}

	afterAlternativePos := len(c.instructions)
	c.changeOperand(jumpPos, afterAlternativePos)

	return nil
}

func NewWithState(s *SymbolTable, constants []obj.Object) *Compiler {
	compiler := New()
	compiler.symbolTable = s
	compiler.constant = constants

	return compiler
}


