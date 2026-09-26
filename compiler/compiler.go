package compiler

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/theawakener0/Zod/ast"
	"github.com/theawakener0/Zod/code"
	"github.com/theawakener0/Zod/lexer"
	obj "github.com/theawakener0/Zod/object"
	"github.com/theawakener0/Zod/parser"
)

type EmittedInstruction struct {
	Opcode   code.Opcode
	Position int
}

type CompilationScope struct {
	instructions        code.Instructions
	lastInstruction     EmittedInstruction
	previousInstruction EmittedInstruction
}

type loopContext struct {
	breakJumps    []int
	continueJumps []int
}

type Compiler struct {
	instructions code.Instructions
	constant     []obj.Object

	lastInstruction     EmittedInstruction
	previousInstruction EmittedInstruction
	symbolTable         *SymbolTable

	scopes      []CompilationScope
	scopesIndex int

	loops []loopContext

	BaseDir    string
	exports    map[string]int
	allSymbols map[string]int
}

type Bytecode struct {
	Instructions code.Instructions
	Constant     []obj.Object
	Exports      map[string]int
	AllSymbols   map[string]int
}

func isTopLevel(c *Compiler) bool { return c.scopesIndex == 0 && c.symbolTable.Outer == nil }

func New() *Compiler {
	mainScope := CompilationScope{
		instructions:        code.Instructions{},
		lastInstruction:     EmittedInstruction{},
		previousInstruction: EmittedInstruction{},
	}
	symbolTable := NewSymbolTable()

	for i, v := range obj.Builtins {
		symbolTable.DefineBuiltin(i, v.Name)
	}

	return &Compiler{
		constant:    make([]obj.Object, 0, 32),
		symbolTable: symbolTable,
		scopes:      []CompilationScope{mainScope},
		scopesIndex: 0,
		exports:     map[string]int{},
		allSymbols:  map[string]int{},
	}
}

func (c *Compiler) Compile(node ast.Node) error {
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

		c.ensureBlockValue()

		jumpPos := c.emit(code.OpJump, 9999)

		afterConsequencePos := len(c.currentInstruction())
		c.changeOperand(jumpNotTruthyPos, afterConsequencePos)

		if node.ElseIf != nil {
			err := c.compileIfExpression(node.ElseIf)
			if err != nil {
				return err
			}
		} else if node.Alternative == nil {
			c.emit(code.OpNull)
		} else {
			err := c.Compile(node.Alternative)
			if err != nil {
				return err
			}

			c.ensureBlockValue()

		}

		afterAlternativePos := len(c.currentInstruction())
		c.changeOperand(jumpPos, afterAlternativePos)
	case *ast.BlockStatement:
		c.enterBlockScope()
		for _, s := range node.Statements {
			err := c.Compile(s)
			if err != nil {
				c.leaveBlockScope()
				return err
			}
		}
		c.leaveBlockScope()
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
			switch s.Scope {
			case LocalScope:
				c.emit(code.OpGetLocalCell, s.Index)
			case FreeScope:
				c.emit(code.OpGetFreeCell, s.Index)
			case GlobalScope:
				c.emit(code.OpGetGlobal, s.Index)
			case BuiltinScope:
				c.emit(code.OpGetBuiltin, s.Index)
			case FunctionScope:
				c.emit(code.OpCurrentClosure)
			}
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
	case *ast.BreakStatement:
		if len(c.loops) == 0 {
			return fmt.Errorf("break used outside of loop")
		}
		pos := c.emit(code.OpJump, 9999)
		c.loops[len(c.loops)-1].breakJumps = append(c.loops[len(c.loops)-1].breakJumps, pos)
	case *ast.ContinueStatement:
		if len(c.loops) == 0 {
			return fmt.Errorf("continue used outside of loop")
		}
		pos := c.emit(code.OpJump, 9999)
		c.loops[len(c.loops)-1].continueJumps = append(c.loops[len(c.loops)-1].continueJumps, pos)
	case *ast.CallExpression:
		if ident, ok := node.Function.(*ast.Identifier); ok && ident.Value == "try" {
			if len(node.Arguments) != 1 {
				tryErr := &obj.Error{Message: fmt.Sprintf("wrong number of arguments. got=%d, want=1", len(node.Arguments))}
				c.emit(code.OpConstant, c.addConstant(tryErr))
				c.emit(code.OpTry)
				return nil
			}
			err := c.Compile(node.Arguments[0])
			if err != nil {
				return err
			}
			c.emit(code.OpTry)
			return nil
		}
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
		c.emitDefine(symbol)
		if c.allSymbols == nil {
			c.allSymbols = map[string]int{}
		}
		if isTopLevel(c) {
			c.allSymbols[node.Name.Value] = symbol.Index
			if node.Public {
				if c.exports == nil {
					c.exports = map[string]int{}
				}
				c.exports[node.Name.Value] = symbol.Index
			}
		}
	case *ast.AssignStatement:
		if err := c.compileAssignStatement(node); err != nil {
			return err
		}
		if isTopLevel(c) && node.Token.Literal == ":=" {
			if ident, ok := node.Left.(*ast.Identifier); ok {
				if sym, ok := c.symbolTable.Resolve(ident.Value); ok {
					if c.allSymbols == nil {
						c.allSymbols = map[string]int{}
					}
					c.allSymbols[ident.Value] = sym.Index
					if node.Public {
						if c.exports == nil {
							c.exports = map[string]int{}
						}
						c.exports[ident.Value] = sym.Index
					}
				}
			}
		}
		return nil
	case *ast.InfixExpression:
		if node.Opt == "&&" {
			return c.compileLogicalAnd(node)
		}
		if node.Opt == "||" {
			return c.compileLogicalOr(node)
		}
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
		case "%":
			c.emit(code.OpMod)
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
		if inner, ok := node.Left.(*ast.IndexExpression); ok {
			if err := c.Compile(inner.Left); err != nil {
				return err
			}
			if err := c.Compile(inner.Index); err != nil {
				return err
			}
			if err := c.Compile(node.Index); err != nil {
				return err
			}
			c.emit(code.OpMatrixCell)
			return nil
		}
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
			return fmt.Errorf("identifier not found: %s", node.Value)
		}

		c.loadSymbol(symbol)
	case *ast.NullLiteral:
		c.emit(code.OpNull)
	case *ast.ForExpression:
		return c.compileForExpression(node)
	case *ast.LoopExpression:
		return c.compileLoopExpression(node)
	case *ast.ImportStatement:
		return c.compileImportStatement(node)
	case *ast.PropertyExpression:
		return c.compilePropertyExpression(node)
	}
	return nil
}

func (c *Compiler) Bytecode() *Bytecode {
	exports := map[string]int{}
	for k, v := range c.exports {
		exports[k] = v
	}
	allSymbols := map[string]int{}
	for k, v := range c.allSymbols {
		allSymbols[k] = v
	}
	return &Bytecode{
		Instructions: c.currentInstruction(),
		Constant:     c.constant,
		Exports:      exports,
		AllSymbols:   allSymbols,
	}
}

func (c *Compiler) addConstant(obj obj.Object) int {
	c.constant = append(c.constant, obj)
	return len(c.constant) - 1
}

func (c *Compiler) addInstruction(ins []byte) int {
	cur := c.scopes[c.scopesIndex].instructions
	posNewInstruction := len(cur)
	c.scopes[c.scopesIndex].instructions = append(cur, ins...)
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

func (c *Compiler) ensureBlockValue() {
	if c.lastInstructionIs(code.OpPop) {
		c.removeLastPop()
		return
	}
	if c.lastInstructionIs(code.OpReturnValue) || c.lastInstructionIs(code.OpReturn) {
		return
	}
	c.emit(code.OpNull)
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

	c.ensureBlockValue()

	jumpPos := c.emit(code.OpJump, 9999)

	afterConsequencePos := len(c.currentInstruction())
	c.changeOperand(jumpNotTruthyPos, afterConsequencePos)

	if node.ElseIf != nil {
		err := c.compileIfExpression(node.ElseIf)
		if err != nil {
			return err
		}
	} else if node.Alternative == nil {
		c.emit(code.OpNull)
	} else {
		err := c.Compile(node.Alternative)
		if err != nil {
			return err
		}

		c.ensureBlockValue()

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
		instructions:        code.Instructions{},
		lastInstruction:     EmittedInstruction{},
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

func (c *Compiler) enterBlockScope() {
	block := NewBlockSymbolTable(c.symbolTable)
	c.symbolTable = block
}

func (c *Compiler) leaveBlockScope() {
	block := c.symbolTable
	outer := block.Outer
	if outer != nil {
		if block.numDefinitions > outer.numDefinitions {
			outer.numDefinitions = block.numDefinitions
		}
		c.symbolTable = outer
	}
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

func (c *Compiler) emitStore(s Symbol) {
	switch s.Scope {
	case GlobalScope:
		c.emit(code.OpSetGlobal, s.Index)
	case LocalScope:
		c.emit(code.OpSetLocal, s.Index)
	case FreeScope:
		c.emit(code.OpSetFree, s.Index)
	default:
		c.emit(code.OpSetGlobal, s.Index)
	}
}

func (c *Compiler) emitDefine(s Symbol) {
	switch s.Scope {
	case GlobalScope:
		c.emit(code.OpSetGlobal, s.Index)
	default:
		c.emit(code.OpDefineLocal, s.Index)
	}
}

func assignKind(op string) (int, error) {
	switch op {
	case "=":
		return 0, nil
	case "+=":
		return 1, nil
	case "-=":
		return 2, nil
	case "*=":
		return 3, nil
	case "/=":
		return 4, nil
	case "%=":
		return 5, nil
	default:
		return 0, fmt.Errorf("unknown assignment operator %s", op)
	}
}

func (c *Compiler) compileAssignStatement(node *ast.AssignStatement) error {
	if idx, ok := node.Left.(*ast.IndexExpression); ok {
		kind, err := assignKind(node.Token.Literal)
		if err != nil {
			return err
		}
		return c.compileIndexAssign(idx, node.Value, kind)
	}

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
		c.emitDefine(symbol)
		return nil
	case "=", "+=", "-=", "*=", "/=", "%=":
		if node.Token.Literal == "=" {
			symbol, found := c.symbolTable.Resolve(ident.Value)
			if !found {
				return fmt.Errorf("identifier not found: %s", ident.Value)
			}
			err := c.Compile(node.Value)
			if err != nil {
				return err
			}
			c.emitStore(symbol)
			return nil
		}
		symbol, found := c.symbolTable.Resolve(ident.Value)
		if !found {
			return fmt.Errorf("identifier not found: %s", ident.Value)
		}
		c.loadSymbol(symbol)
		err := c.Compile(node.Value)
		if err != nil {
			return err
		}
		switch node.Token.Literal {
		case "+=":
			c.emit(code.OpAdd)
		case "-=":
			c.emit(code.OpSub)
		case "*=":
			c.emit(code.OpMul)
		case "/=":
			c.emit(code.OpDiv)
		case "%=":
			c.emit(code.OpMod)
		}
		c.emitStore(symbol)
		return nil
	default:
		return fmt.Errorf("unknown assignment operator %s", node.Token.Literal)
	}
}

func (c *Compiler) compileIndexAssign(idx *ast.IndexExpression, value ast.Expression, kind int) error {
	if inner, ok := idx.Left.(*ast.IndexExpression); ok {
		if err := c.Compile(inner.Left); err != nil {
			return err
		}
		if err := c.Compile(inner.Index); err != nil {
			return err
		}
		if err := c.Compile(idx.Index); err != nil {
			return err
		}
		if err := c.Compile(value); err != nil {
			return err
		}
		c.emit(code.OpSetIndex, 10+kind)
		return nil
	}
	if err := c.Compile(idx.Left); err != nil {
		return err
	}
	if err := c.Compile(idx.Index); err != nil {
		return err
	}
	if err := c.Compile(value); err != nil {
		return err
	}
	c.emit(code.OpSetIndex, kind)
	return nil
}

func (c *Compiler) compileLogicalAnd(node *ast.InfixExpression) error {
	if err := c.Compile(node.Left); err != nil {
		return err
	}
	c.emit(code.OpDup)
	jumpPos := c.emit(code.OpJumpNotTruthy, 9999)
	c.emit(code.OpPop)
	if err := c.Compile(node.Right); err != nil {
		return err
	}
	c.changeOperand(jumpPos, len(c.currentInstruction()))
	return nil
}

func (c *Compiler) compileLogicalOr(node *ast.InfixExpression) error {
	if err := c.Compile(node.Left); err != nil {
		return err
	}
	c.emit(code.OpDup)
	jumpNotTruthyPos := c.emit(code.OpJumpNotTruthy, 9999)
	jumpEndPos := c.emit(code.OpJump, 9999)
	rightPos := len(c.currentInstruction())
	c.changeOperand(jumpNotTruthyPos, rightPos)
	c.emit(code.OpPop)
	if err := c.Compile(node.Right); err != nil {
		return err
	}
	c.changeOperand(jumpEndPos, len(c.currentInstruction()))
	return nil
}

func (c *Compiler) compileForExpression(node *ast.ForExpression) error {
	c.enterBlockScope()
	c.loops = append(c.loops, loopContext{})

	before := make(map[string]struct{}, len(c.symbolTable.store))
	for name := range c.symbolTable.store {
		before[name] = struct{}{}
	}

	if node.Init != nil {
		if err := c.Compile(node.Init); err != nil {
			return err
		}
	}

	type shadowedVar struct {
		name string
		idx  int
	}
	var shadowed []shadowedVar
	for name, sym := range c.symbolTable.store {
		if _, ok := before[name]; ok {
			continue
		}
		if sym.Scope != LocalScope {
			continue
		}
		shadowed = append(shadowed, shadowedVar{name: name, idx: sym.Index})
	}
	sort.Slice(shadowed, func(a, b int) bool { return shadowed[a].idx < shadowed[b].idx })

	condPos := len(c.currentInstruction())
	var endJump int = -1
	if node.Condition != nil {
		if err := c.Compile(node.Condition); err != nil {
			return err
		}
		endJump = c.emit(code.OpJumpNotTruthy, 9999)
	}

	if node.Body == nil {
		return fmt.Errorf("for missing body")
	}
	c.enterBlockScope()
	for _, sv := range shadowed {
		shadow := c.symbolTable.Define(sv.name)
		c.emit(code.OpGetLocal, sv.idx)
		c.emit(code.OpDefineLocal, shadow.Index)
	}
	for _, s := range node.Body.Statements {
		if err := c.Compile(s); err != nil {
			c.leaveBlockScope()
			return err
		}
	}
	c.leaveBlockScope()

	updatePos := len(c.currentInstruction())
	if node.Update != nil {
		if err := c.Compile(node.Update); err != nil {
			return err
		}
	}

	c.emit(code.OpJump, condPos)

	endPos := len(c.currentInstruction())
	if endJump != -1 {
		c.changeOperand(endJump, endPos)
	}
	ctx := c.loops[len(c.loops)-1]
	for _, pos := range ctx.breakJumps {
		c.changeOperand(pos, endPos)
	}
	for _, pos := range ctx.continueJumps {
		c.changeOperand(pos, updatePos)
	}
	c.loops = c.loops[:len(c.loops)-1]
	c.leaveBlockScope()

	c.emit(code.OpNull)
	return nil
}

func (c *Compiler) compileLoopExpression(node *ast.LoopExpression) error {
	c.enterBlockScope()
	c.loops = append(c.loops, loopContext{})

	startPos := len(c.currentInstruction())
	if node.Body == nil {
		return fmt.Errorf("loop missing body")
	}
	if err := c.Compile(node.Body); err != nil {
		return err
	}
	c.emit(code.OpJump, startPos)

	endPos := len(c.currentInstruction())
	ctx := c.loops[len(c.loops)-1]
	for _, pos := range ctx.breakJumps {
		c.changeOperand(pos, endPos)
	}
	for _, pos := range ctx.continueJumps {
		c.changeOperand(pos, startPos)
	}
	c.loops = c.loops[:len(c.loops)-1]
	c.leaveBlockScope()

	c.emit(code.OpNull)
	return nil
}

func (c *Compiler) compileIncrementDecrement(ident *ast.Identifier, op string, isPostfix bool) error {
	symbol, ok := c.symbolTable.Resolve(ident.Value)
	if !ok {
		return fmt.Errorf("identifier not found: %s", ident.Value)
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

	if !isPostfix {
		c.emit(code.OpDup)
	}

	switch symbol.Scope {
	case GlobalScope, LocalScope, FreeScope:
		c.emitStore(symbol)
		return nil
	default:
		return fmt.Errorf("cannot ++/-- %s", ident.Value)
	}
}

func NewWithState(s *SymbolTable, constants []obj.Object) *Compiler {
	compiler := New()
	compiler.symbolTable = s
	compiler.constant = constants

	return compiler
}

func statExists(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}

func stdlibPath() string {
	if p := os.Getenv("ZOD_STDLIB"); p != "" {
		return p
	}
	if exe, err := os.Executable(); err == nil {
		return filepath.Join(filepath.Dir(exe), "stdlib")
	}
	return "stdlib"
}

func stdlibCandidateDirs(baseDir string) []string {
	var dirs []string
	if p := os.Getenv("ZOD_STDLIB"); p != "" {
		dirs = append(dirs, p)
	}
	if exe, err := os.Executable(); err == nil {
		dirs = append(dirs, filepath.Join(filepath.Dir(exe), "stdlib"))
	}
	if cwd, err := os.Getwd(); err == nil {
		dirs = append(dirs, filepath.Join(cwd, "stdlib"))
	}
	start := baseDir
	if start == "" {
		if cwd, err := os.Getwd(); err == nil {
			start = cwd
		}
	}
	if start != "" {
		if abs, err := filepath.Abs(start); err == nil {
			start = abs
		}
		seen := make(map[string]bool, len(dirs)+5)
		for _, d := range dirs {
			if abs, err := filepath.Abs(d); err == nil {
				seen[abs] = true
			} else {
				seen[d] = true
			}
		}
		cur := start
		for i := 0; i < 5; i++ {
			cand := filepath.Join(cur, "stdlib")
			absCand := cand
			if abs, err := filepath.Abs(cand); err == nil {
				absCand = abs
			}
			if !seen[absCand] {
				dirs = append(dirs, cand)
				seen[absCand] = true
			}
			parent := filepath.Dir(cur)
			if parent == cur {
				break
			}
			cur = parent
		}
	}
	return dirs
}

func resolveStdModulePath(rel, baseDir string) (string, bool) {
	if rel == "" {
		return "", false
	}
	var variants []string
	if strings.HasSuffix(rel, ".zd") {
		variants = []string{rel}
	} else {
		variants = []string{rel, rel + ".zd"}
	}
	for _, dir := range stdlibCandidateDirs(baseDir) {
		for _, v := range variants {
			cand := filepath.Join(dir, v)
			if statExists(cand) {
				if abs, err := filepath.Abs(cand); err == nil {
					return abs, true
				}
				return cand, true
			}
		}
	}
	return "", false
}

func ResolveModulePath(baseDir, path string) (string, error) {
	if strings.HasPrefix(path, "std/") {
		rel := strings.TrimPrefix(path, "std/")
		if abs, ok := resolveStdModulePath(rel, baseDir); ok {
			return abs, nil
		}
		return "", fmt.Errorf("module not found")
	}
	if filepath.IsAbs(path) {
		if statExists(path) {
			return path, nil
		}
		if !strings.HasSuffix(path, ".zd") {
			if statExists(path + ".zd") {
				return path + ".zd", nil
			}
		}
	} else {
		if baseDir != "" {
			cand := filepath.Join(baseDir, path)
			if statExists(cand) {
				return cand, nil
			}
			if !strings.HasSuffix(path, ".zd") {
				candExt := filepath.Join(baseDir, path+".zd")
				if statExists(candExt) {
					return candExt, nil
				}
			}
		}
		if abs, err := filepath.Abs(path); err == nil {
			if statExists(abs) {
				return abs, nil
			}
		}
		if !strings.HasSuffix(path, ".zd") {
			if abs, err := filepath.Abs(path + ".zd"); err == nil {
				if statExists(abs) {
					return abs, nil
				}
			}
		}
	}
	std := stdlibPath()
	cand := filepath.Join(std, path)
	if statExists(cand) {
		if abs, err := filepath.Abs(cand); err == nil {
			return abs, nil
		}
		return cand, nil
	}
	if !strings.HasSuffix(path, ".zd") {
		candExt := filepath.Join(std, path+".zd")
		if statExists(candExt) {
			if abs, err := filepath.Abs(candExt); err == nil {
				return abs, nil
			}
			return candExt, nil
		}
	}
	return "", fmt.Errorf("module not found")
}

func CollectModuleExports(absPath string) ([]string, []string, error) {
	data, err := os.ReadFile(absPath)
	if err != nil {
		return nil, nil, err
	}
	l := lexer.New(string(data))
	p := parser.New(l)
	prog := p.ParseProgram()
	if len(p.Errors()) > 0 {
		return nil, nil, fmt.Errorf("%s", strings.Join(p.Errors(), "; "))
	}
	var exports []string
	var all []string
	if prog == nil {
		return exports, all, nil
	}
	for _, s := range prog.Statements {
		switch node := s.(type) {
		case *ast.LetStatement:
			all = append(all, node.Name.Value)
			if node.Public {
				exports = append(exports, node.Name.Value)
			}
		case *ast.AssignStatement:
			if node.Token.Literal == ":=" {
				if ident, ok := node.Left.(*ast.Identifier); ok {
					all = append(all, ident.Value)
					if node.Public {
						exports = append(exports, ident.Value)
					}
				}
			}
		}
	}
	return exports, all, nil
}

func (c *Compiler) compileImportStatement(node *ast.ImportStatement) error {
	if !isTopLevel(c) {
		return fmt.Errorf("import only supported at top-level in VM")
	}
	rawPath := node.Path.Value
	absPath, resolveErr := ResolveModulePath(c.BaseDir, rawPath)
	specPath := rawPath
	if resolveErr == nil {
		specPath = absPath
	}
	if node.IsFrom && node.IsStar {
		var names []string
		if resolveErr == nil {
			if ex, _, cerr := CollectModuleExports(absPath); cerr == nil {
				names = ex
			}
		}
		specNames := []string{}
		targets := []int{}
		for _, n := range names {
			// Imports must shadow builtins (e.g. `keys`, `println`).
			// DefineIfNotExists would keep the BUILTIN symbol and the
			// VM would silently keep calling the old builtin.
			sym := c.symbolTable.Define(n)
			if c.allSymbols == nil {
				c.allSymbols = map[string]int{}
			}
			c.allSymbols[n] = sym.Index
			specNames = append(specNames, n)
			targets = append(targets, sym.Index)
		}
		spec := &obj.ImportSpec{Path: specPath, Kind: obj.ImportStar, Names: specNames, Targets: targets, AliasIndex: -1}
		idx := c.addConstant(spec)
		c.emit(code.OpImport, idx)
		c.emit(code.OpPop)
		return nil
	}
	if node.IsFrom {
		names := []string{}
		targets := []int{}
		for _, id := range node.Names {
			// See above: imports shadow builtins.
			sym := c.symbolTable.Define(id.Value)
			if c.allSymbols == nil {
				c.allSymbols = map[string]int{}
			}
			c.allSymbols[id.Value] = sym.Index
			names = append(names, id.Value)
			targets = append(targets, sym.Index)
		}
		spec := &obj.ImportSpec{Path: specPath, Kind: obj.ImportNamed, Names: names, Targets: targets, AliasIndex: -1}
		idx := c.addConstant(spec)
		c.emit(code.OpImport, idx)
		c.emit(code.OpPop)
		return nil
	}
	alias := ""
	if node.Alias != nil {
		alias = node.Alias.Value
	} else {
		base := filepath.Base(rawPath)
		alias = strings.TrimSuffix(base, ".zd")
		if alias == "" {
			alias = rawPath
		}
	}
	sym := c.symbolTable.Define(alias)
	if c.allSymbols == nil {
		c.allSymbols = map[string]int{}
	}
	c.allSymbols[alias] = sym.Index
	spec := &obj.ImportSpec{Path: specPath, Kind: obj.ImportModule, Names: []string{}, Targets: []int{}, AliasIndex: sym.Index, Alias: alias}
	idx := c.addConstant(spec)
	c.emit(code.OpImport, idx)
	c.emit(code.OpPop)
	return nil
}

func (c *Compiler) compilePropertyExpression(node *ast.PropertyExpression) error {
	if err := c.Compile(node.Object); err != nil {
		return err
	}
	propStr := &obj.String{Value: node.Property.Value}
	idx := c.addConstant(propStr)
	c.emit(code.OpGetProp, idx)
	return nil
}
