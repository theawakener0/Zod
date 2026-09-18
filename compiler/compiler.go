package compiler

import (
	"fmt"

	"github.com/theawakener0/Zod/ast"
	"github.com/theawakener0/Zod/code"
	obj "github.com/theawakener0/Zod/object"
)

type EmittedInstruction struct {
	Opcode 		code.Opcode
	Position 	int
}

type CompilationScope struct {
	instructions 			code.Instructions
	lastInstruction 		EmittedInstruction
	previousInstruction 	EmittedInstruction
}

type Compiler struct {
	instructions 		code.Instructions
	constant 			[]obj.Object

	lastInstruction 	EmittedInstruction
	previousInstruction EmittedInstruction
	symbolTable 		*SymbolTable

	scopes 				[]CompilationScope
	scopesIndex			int
}

type Bytecode struct {
	Instructions 	code.Instructions
	Constant 		[]obj.Object
}

func New() *Compiler {
	mainScope := CompilationScope{
		instructions: code.Instructions{},
		lastInstruction: EmittedInstruction{},
		previousInstruction: EmittedInstruction{},
	}
	symbolTable := NewSymbolTable()

	for i, v := range obj.Builtins {
		symbolTable.DefineBuiltin(i, v.Name)
	}

	return &Compiler{
		constant: []obj.Object{},
		symbolTable: symbolTable,
		scopes: []CompilationScope{mainScope},
		scopesIndex: 0,
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

		if c.lastInstructionIs(code.OpPop) {
			c.removeLastPop()
		}

		jumpPos := c.emit(code.OpJump, 9999)

		afterConsequencePos := len(c.currentInstruction())
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

			if c.lastInstructionIs(code.OpPop) {
				c.removeLastPop()
			}

		}

		afterAlternativePos := len(c.currentInstruction())
		c.changeOperand(jumpPos, afterAlternativePos)
	case *ast.BlockStatement:
		for _, s := range node.Statements {
			err := c.Compile(s)
			if err != nil {
				return err
			}
		}
	case *ast.FunctionLiteral:
		c.enterScope()

		if node.Name != "" {
			c.symbolTable.DefineFunctionName(node.Name)
		}

		for _, param := range node.Parameters {
			c.symbolTable.Define(param.Value)
		}

		err := c.Compile(node.Body)
		if err != nil {
			return err
		}

		if c.lastInstructionIs(code.OpPop) {
			c.replaceLastPopWithReturn()
		}

		if !c.lastInstructionIs(code.OpReturnValue) {
			c.emit(code.OpReturn)
		}

		freeSymbols := c.symbolTable.FreeSymbols
		numLocals := c.symbolTable.numDefinitions
		instructions := c.leaveScope()

		for _, s := range freeSymbols {
			c.loadSymbol(s)
		}

		compiledFn := &obj.CompiledFunction{Instructions: instructions, NumLocals: numLocals, NumParams: len(node.Parameters)}
		fnIndex := c.addConstant(compiledFn)
		c.emit(code.OpClosure, fnIndex, len(freeSymbols))
	case *ast.ReturnStatement:
		err := c.Compile(node.ReturnValue)
		if err != nil {
			return err
		}
		
		c.emit(code.OpReturnValue)
	case *ast.CallExpression:
		err := c.Compile(node.Function)
		if err != nil {
			return err
		}

		for _, arg := range node.Arguments {
			err := c.Compile(arg)
			if err != nil {
				return err
			}
		}

		c.emit(code.OpCall, len(node.Arguments))
	case *ast.LetStatement:
		symbol := c.symbolTable.Define(node.Name.Value)
		err := c.Compile(node.Value)
		if err != nil {
			return err
		}
		if symbol.Scope == GlobalScope {
			c.emit(code.OpSetGlobal, symbol.Index)
		} else {
			c.emit(code.OpSetLocal, symbol.Index)
		}
	case *ast.AssignStatement:
		ident, ok := node.Left.(*ast.Identifier)
		if !ok {
			return fmt.Errorf("only identifier assignment supported, got %T", node.Left)
		}
		switch node.Token.Literal {
		case ":=":
			symbol := c.symbolTable.DefineIfNotExists(ident.Value)
			err := c.Compile(node.Value)
			if err != nil {
				return err
			}
			if symbol.Scope == GlobalScope {
				c.emit(code.OpSetGlobal, symbol.Index)
			} else {
				c.emit(code.OpSetLocal, symbol.Index)
			}
		case "=":
			symbol, found := c.symbolTable.Resolve(ident.Value)
			if !found {
				return fmt.Errorf("identifier not found: %s", ident.Value)
			}
			err := c.Compile(node.Value)
			if err != nil {
				return err
			}
			if symbol.Scope == GlobalScope {
				c.emit(code.OpSetGlobal, symbol.Index)
			} else {
				c.emit(code.OpSetLocal, symbol.Index)
			}
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
		if node.Opt == "++" || node.Opt == "--" {
			ident, ok := node.Right.(*ast.Identifier)
			if !ok {
				return fmt.Errorf("++/-- requires identifier")
			}
			return c.compileIncrementDecrement(ident, node.Opt, false)
		}

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
	case *ast.PostfixExpression:
		if node.Opt == "++" || node.Opt == "--" {
			ident, ok := node.Left.(*ast.Identifier)
			if !ok {
				return fmt.Errorf("++/-- requires identifier")
			}
			return c.compileIncrementDecrement(ident, node.Opt, true)
		}
		return fmt.Errorf("unknown postfix operator %s", node.Opt)
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
		for _, pairs := range node.Pairs {
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

		c.loadSymbol(symbol)
	case *ast.NullLiteral:
		c.emit(code.OpNull)
	}
	return nil
}

func (c *Compiler) Bytecode() *Bytecode {
	return &Bytecode{
		Instructions: c.currentInstruction(),
		Constant: c.constant,
	}
}

func (c *Compiler) addConstant(obj obj.Object) int {
	c.constant = append(c.constant, obj)
	return len(c.constant) - 1
}

func (c *Compiler) addInstruction(ins []byte) int {
	posNewInstruction := len(c.currentInstruction())
	updatedInstructions := append(c.currentInstruction(), ins...)

	c.scopes[c.scopesIndex].instructions = updatedInstructions
	return posNewInstruction
}

func (c *Compiler) emit(op code.Opcode, operands ...int) int {
	ins := code.Make(op, operands...)
	pos := c.addInstruction(ins)

	c.setLastInstruction(op, pos)

	return pos
}

func (c *Compiler) setLastInstruction(op code.Opcode, pos int) {
	previous := c.scopes[c.scopesIndex].lastInstruction
	last := EmittedInstruction{Opcode: op, Position: pos}

	c.scopes[c.scopesIndex].previousInstruction = previous
	c.scopes[c.scopesIndex].lastInstruction = last
}

func (c *Compiler) lastInstructionIs(op code.Opcode) bool {
	if len(c.currentInstruction()) == 0 {
		return false
	}

	return c.scopes[c.scopesIndex].lastInstruction.Opcode == op
}

func (c *Compiler) removeLastPop() {
	last := c.scopes[c.scopesIndex].lastInstruction
	previous := c.scopes[c.scopesIndex].previousInstruction

	old := c.currentInstruction()
	new := old[:last.Position]

	c.scopes[c.scopesIndex].instructions = new
	c.scopes[c.scopesIndex].lastInstruction = previous
}

func (c *Compiler) replaceInstruction(pos int, newInstruction []byte) {
	ins := c.currentInstruction()

	for i := range len(newInstruction) {
		ins[pos+i] = newInstruction[i]
	}
}

func (c *Compiler) replaceLastPopWithReturn() {
	lasrPos := c.scopes[c.scopesIndex].lastInstruction.Position
	c.replaceInstruction(lasrPos, code.Make(code.OpReturnValue))

	c.scopes[c.scopesIndex].lastInstruction.Opcode = code.OpReturnValue
}

func (c *Compiler) changeOperand(opPos int, operand int) {
	ins := c.currentInstruction()

	op := code.Opcode(ins[opPos])
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

	if c.lastInstructionIs(code.OpPop) {
		c.removeLastPop()
	}

	jumpPos := c.emit(code.OpJump, 9999)

	afterConsequencePos := len(c.currentInstruction())
	c.changeOperand(jumpNotTruthyPos, afterConsequencePos)

	if node.Alternative == nil {
		c.emit(code.OpNull)
	} else {
		err := c.Compile(node.Alternative)
		if err != nil {
			return err
		}

		if c.lastInstructionIs(code.OpPop) {
			c.removeLastPop()
		}

	}

	afterAlternativePos := len(c.currentInstruction())
	c.changeOperand(jumpPos, afterAlternativePos)

	return nil
}

func (c *Compiler) currentInstruction() code.Instructions {
	return c.scopes[c.scopesIndex].instructions
}

func (c *Compiler) enterScope() {
	scope := CompilationScope{
		instructions: code.Instructions{},
		lastInstruction: EmittedInstruction{},
		previousInstruction: EmittedInstruction{},
	}
	c.scopes = append(c.scopes, scope)
	c.scopesIndex++

	c.symbolTable = NewEnclosedSymbolTable(c.symbolTable)
}

func (c *Compiler) leaveScope() code.Instructions {
	instructions := c.currentInstruction()

	c.scopes = c.scopes[:len(c.scopes)-1]
	c.scopesIndex--

	c.symbolTable = c.symbolTable.Outer

	return instructions
}

func (c *Compiler) loadSymbol(s Symbol) {
	switch s.Scope {
	case GlobalScope:
		c.emit(code.OpGetGlobal, s.Index)
	case LocalScope:
		c.emit(code.OpGetLocal, s.Index)
	case BuiltinScope:
		c.emit(code.OpGetBuiltin, s.Index)
	case FreeScope:
		c.emit(code.OpGetFree, s.Index)
	case FunctionScope:
		c.emit(code.OpCurrentClosure)
	}
}

func (c *Compiler) compileIncrementDecrement(ident *ast.Identifier, op string, isPostfix bool) error {
	symbol, ok := c.symbolTable.Resolve(ident.Value)
	if !ok {
		return fmt.Errorf("undefined variable %s", ident.Value)
	}

	c.loadSymbol(symbol)

	if isPostfix {
		c.emit(code.OpDup)
	}

	one := c.addConstant(&obj.Integer{Value: 1})
	c.emit(code.OpConstant, one)
	if op == "++" {
		c.emit(code.OpAdd)
	} else {
		c.emit(code.OpSub)
	}

	switch symbol.Scope {
	case GlobalScope:
		c.emit(code.OpSetGlobal, symbol.Index)
	case LocalScope:
		c.emit(code.OpSetLocal, symbol.Index)
	default:
		return fmt.Errorf("cannot ++/-- %s", ident.Value)
	}

	return nil
}

func NewWithState(s *SymbolTable, constants []obj.Object) *Compiler {
	compiler := New()
	compiler.symbolTable = s
	compiler.constant = constants

	return compiler
}


