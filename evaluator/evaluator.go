package evaluator

import (
	"fmt"

	"github.com/theawakener0/zod/ast"
	obj "github.com/theawakener0/zod/object"
)

var (
	NULL  = &obj.Null{}
	TRUE  = &obj.Boolean{Value: true}
	FALSE = &obj.Boolean{Value: false}
)

func Eval(node ast.Node, env *obj.Enviroment) obj.Object {
	switch n := node.(type) {
	case *ast.Program:
		return evalProgram(n, env)

	case *ast.ExpressionStatement:
		return Eval(n.Expression, env)

	case *ast.IntegerLiteral:
		return &obj.Integer{Value: n.Value}
	
	case *ast.Boolean:
		return nattiveBoolToBooleanObject(n.Value)

	case *ast.NullLiteral:
		return NULL
	
	case *ast.PrefixExpression:
		if n.Opt == "++" || n.Opt == "--" {
			return evalIncrementDecrement(n.Opt, n.Right, env, false)
		}
		right := Eval(n.Right, env)
		if isError(right) {
			return right
		}
		return evalPrefixExpression(n.Opt, right)

	case *ast.PostfixExpression:
		if n.Opt == "++" || n.Opt == "--" {
			return evalIncrementDecrement(n.Opt, n.Left, env, true)
		}
		left := Eval(n.Left, env)
		if isError(left) {
			return left
		}
		return newError("unknown postfix operator: %s%s", left.Type(), n.Opt)
	
	case *ast.InfixExpression:
		if n.Opt == "&&" {
			left := Eval(n.Left, env)
			if isError(left) {
				return left
			}
			if !isTruthy(left) {
				return left
			}
			return Eval(n.Right, env)
		}
		if n.Opt == "||" {
			left := Eval(n.Left, env)
			if isError(left) {
				return left
			}
			if isTruthy(left) {
				return left
			}
			return Eval(n.Right, env)
		}

		left := Eval(n.Left, env)
		if isError(left) {
			return left
		}

		right := Eval(n.Right, env)
		if isError(right) {
			return right
		}

		return evalInfixExpression(n.Opt, left, right)
	
	case *ast.BlockStatement:
		return evalBlockStatement(n, env)
	
	case *ast.IfExpression:
		return evalIfExpression(n, env)

	case *ast.ReturnStatement:
		val := Eval(n.ReturnValue, env)
		if isError(val) {
			return val
		}
		return &obj.ReturnValue{Value: val}

	case *ast.BreakStatement:
		return &obj.Break{}

	case *ast.ContinueStatement:
		return &obj.Continue{}

	case *ast.LetStatement:
		val := Eval(n.Value, env)
		if isError(val) {
			return val
		}
		env.Set(n.Name.Value, val)

	case *ast.AssignStatement:
		val := Eval(n.Value, env)
		if isError(val) {
			return val
		}

		if idx, ok := n.Left.(*ast.IndexExpression); ok {
			return evalIndexAssignment(idx, n.Token.Literal, val, env)
		}

		ident := n.Left.(*ast.Identifier)

		switch n.Token.Literal {
		case ":=":
			env.Set(ident.Value, val)
		case "=":
			if _, ok := env.Get(ident.Value); !ok {
				return newError("identifier not found: %s", ident.Value)
			}
			env.Set(ident.Value, val)
		case "+=":
			curr := Eval(ident, env)
			if isError(curr) {
				return curr
			}
			result := evalInfixExpression("+", curr, val)
			if isError(result) {
				return result
			}
			env.Set(ident.Value, result)
		case "-=":
			curr := Eval(ident, env)
			if isError(curr) {
				return curr
			}
			result := evalInfixExpression("-", curr, val)
			if isError(result) {
				return result
			}
			env.Set(ident.Value, result)
		case "*=":
			curr := Eval(ident, env)
			if isError(curr) {
				return curr
			}
			result := evalInfixExpression("*", curr, val)
			if isError(result) {
				return result
			}
			env.Set(ident.Value, result)
		case "/=":
			curr := Eval(ident, env)
			if isError(curr) {
				return curr
			}
			result := evalInfixExpression("/", curr, val)
			if isError(result) {
				return result
			}
			env.Set(ident.Value, result)
		}

	case *ast.Identifier:
		return evalIdentifier(n, env)

	case *ast.FunctionLiteral:
		params := n.Parameters
		body := n.Body
		return &obj.Function{Parameters: params, Env: env, Body: body}

	case *ast.CallExpression:
		fn := Eval(n.Function, env)
		if isError(fn) {
			return fn
		}
		args := evalExpressions(n.Arguments, env)
		if len(args) == 1 && isError(args[0]) {
			return args[0]
		}

		return applyFunction(fn, args)

	case *ast.StringLiteral:
		return &obj.String{Value: n.Value}

	case *ast.ForExpression:
		return evalForExpression(n, env)

	case *ast.LoopExpression:
		return evalLoopExpression(n, env)
	case *ast.ArrayLiteral:
		elements := evalExpressions(n.Elements, env)
		if len(elements) == 1 && isError(elements[0]) {
			return elements[0]
		}
		return &obj.Array{Elements: elements}
	case *ast.IndexExpression:
		left := Eval(n.Left, env)
		if isError(left) {
			return left
		}

		index := Eval(n.Index, env)
		if isError(index) {
			return index
		}

		return evalIndexExpression(left, index)
	case *ast.HashLiteral:
		return evalHashLiteral(n, env)
	}

	return nil
}

func evalStatements(stmts []ast.Statement, env *obj.Enviroment) obj.Object {
	var result obj.Object

	for _, stmt := range stmts {
		result = Eval(stmt, env)

		if returnValue, ok := result.(*obj.ReturnValue); ok {
			return returnValue.Value
		}
	}

	return result
}

func nattiveBoolToBooleanObject(input bool) *obj.Boolean {
	if input {
		return TRUE
	}

	return FALSE
}

func evalBangOperatorExpression(right obj.Object) obj.Object {
	switch right {
	case TRUE:
		return FALSE
	case FALSE:
		return TRUE
	case NULL:
		return TRUE
	default:
		return FALSE
	}
}

func evalMinusPrefixOperatorExpression(right obj.Object) obj.Object {
	if right.Type() != obj.INTEGER_OBJ {
		return newError("unknown prefix operator: -%s", right.Type())
	}

	value := right.(*obj.Integer).Value
	return &obj.Integer{Value: -value}
}

func evalPrefixExpression(operator string, right obj.Object) obj.Object {
	switch operator {
	case "!":
		return evalBangOperatorExpression(right)
	case "-":
		return evalMinusPrefixOperatorExpression(right)
	default:
		return newError("unknown prefix operator: %s%s", operator, right.Type())
	}
}

func evalIntegerInfixExpression(operator string, left, right obj.Object) obj.Object {
	leftValue := left.(*obj.Integer).Value
	rightValue := right.(*obj.Integer).Value

	switch operator {
	case "+":
		return &obj.Integer{Value: leftValue + rightValue}
	case "-":
		return &obj.Integer{Value: leftValue - rightValue}
	case "*":
		return &obj.Integer{Value: leftValue * rightValue}
	case "/":
		if rightValue == 0 {
			return newError("division by zero")
		}
		return &obj.Integer{Value: leftValue / rightValue}
	case "<":
		return nattiveBoolToBooleanObject(leftValue < rightValue)
	case ">":
		return nattiveBoolToBooleanObject(leftValue > rightValue)
	case "==":
		return nattiveBoolToBooleanObject(leftValue == rightValue)
	case "!=":
		return nattiveBoolToBooleanObject(leftValue != rightValue)
	case "<=":
		return nattiveBoolToBooleanObject(leftValue <= rightValue)
	case ">=":
		return nattiveBoolToBooleanObject(leftValue >= rightValue)
	default:
		return newError("unknown infix operator: %s %s %s", left.Type(), operator, right.Type())
	}
}

func evalInfixExpression(operator string, left, right obj.Object) obj.Object {
	switch {
	case left.Type() == obj.INTEGER_OBJ && right.Type() == obj.INTEGER_OBJ:
		return evalIntegerInfixExpression(operator, left, right)
	case left.Type() == obj.STRING_OBJ && right.Type() == obj.STRING_OBJ:
		return evalStringInfixExpression(operator, left, right)
	case operator == "==":
		return nattiveBoolToBooleanObject(left == right)
	case operator == "!=":
		return nattiveBoolToBooleanObject(left != right)
	default:
		return newError("unknown infix operator: %s %s %s", left.Type(), operator, right.Type())
	}
}

func isTruthy(object obj.Object) bool {
	switch object {
	case TRUE:
		return true
	case FALSE:
		return false
	case NULL:
		return false
	default:
		return true
	}
}

func evalIfExpression(ie *ast.IfExpression, env *obj.Enviroment) obj.Object {
	condition := Eval(ie.Condition, env)
	if isError(condition) {
		return condition
	}

	if isTruthy(condition) {
		return Eval(ie.Consequence, env)
	}

	if ie.ElseIf != nil {
		return evalIfExpression(ie.ElseIf, env)
	}

	if ie.Alternative != nil {
		return Eval(ie.Alternative, env)
	}

	return NULL
}

func evalForExpression(fe *ast.ForExpression, env *obj.Enviroment) obj.Object {
	if fe.Init != nil {
		result := Eval(fe.Init, env)
		if isError(result) {
			return result
		}
	}

	for {
		var condition obj.Object
		if fe.Condition != nil {
			condition = Eval(fe.Condition, env)
		} else {
			condition = TRUE
		}

		if isError(condition) {
			return condition
		}

		if !isTruthy(condition) {
			break
		}

		result := Eval(fe.Body, env)
		if result != nil {
			switch result.Type() {
			case obj.RETURN_VALUE_OBJ, obj.ERROR_OBJ:
				return result
			case obj.BREAK_OBJ:
				return NULL
			case obj.CONTINUE_OBJ:
			}
		}

		if fe.Update != nil {
			result := Eval(fe.Update, env)
			if isError(result) {
				return result
			}
		}
	}

	return NULL
}

func evalLoopExpression(le *ast.LoopExpression, env *obj.Enviroment) obj.Object {
	for {
		result := Eval(le.Body, env)
		if result != nil {
			switch result.Type() {
			case obj.RETURN_VALUE_OBJ, obj.ERROR_OBJ:
				return result
			case obj.BREAK_OBJ:
				return NULL
			case obj.CONTINUE_OBJ:
				continue
			}
		}
	}
}

func evalIncrementDecrement(opt string, right ast.Expression, env *obj.Enviroment, isPostfix bool) obj.Object {
	ident, ok := right.(*ast.Identifier)
	if !ok {
		return newError("%s requires identifier, got %T", opt, right)
	}

	val, ok := env.Get(ident.Value)
	if !ok {
		return newError("identifier not found: %s", ident.Value)
	}

	if val.Type() != obj.INTEGER_OBJ {
		return newError("%s requires integer, got %s", opt, val.Type())
	}

	intVal := val.(*obj.Integer).Value
	oldObj := &obj.Integer{Value: intVal}
	if opt == "++" {
		intVal++
	} else {
		intVal--
	}

	newObj := &obj.Integer{Value: intVal}
	env.Set(ident.Value, newObj)

	if isPostfix {
		return oldObj
	}
	return newObj
}

func evalProgram(p *ast.Program, env *obj.Enviroment) obj.Object {
	var result obj.Object

	for _, stmt := range p.Statements {
		result = Eval(stmt, env)
		
		switch result := result.(type) {
		case *obj.ReturnValue:
			return result.Value
		case *obj.Error:
			return result
		case *obj.Break:
			return newError("break used outside of loop")
		case *obj.Continue:
			return newError("continue used outside of loop")
		}
	}

	return result
}

func evalBlockStatement(b *ast.BlockStatement, env *obj.Enviroment) obj.Object {
	var result obj.Object

	for _, stmt := range b.Statements {
		result = Eval(stmt, env)

		if result != nil {
			rt := result.Type()
			if rt == obj.RETURN_VALUE_OBJ || rt == obj.ERROR_OBJ ||
				rt == obj.BREAK_OBJ || rt == obj.CONTINUE_OBJ {
				return result
			}
		}
	}

	return result
}

func newError(format string, args ...any) *obj.Error {
	return &obj.Error{Message: fmt.Sprintf(format, args...)}
}

func isError(object obj.Object) bool {
	if object != nil {
		return object.Type() == obj.ERROR_OBJ
	}
	return false
}

func evalIdentifier(node *ast.Identifier, env *obj.Enviroment) obj.Object {
	if val, ok := env.Get(node.Value); ok {
		return val
	}

	if builtin, ok := builtins[node.Value]; ok {
		return builtin
	}

	return newError("identifier not found: %s", node.Value)
}

func evalExpressions(exps []ast.Expression, env *obj.Enviroment) []obj.Object {
	var result []obj.Object

	for _, exp := range exps {
		eval := Eval(exp, env)
		if isError(eval) {
			return []obj.Object{eval}
		}
		result = append(result, eval)
	}

	return result
}

func unwrapReturnValue(object obj.Object) obj.Object {
	if returnValue, ok := object.(*obj.ReturnValue); ok {
		return returnValue.Value
	}

	return object
}

func extendFunctonEnv(fn *obj.Function, args []obj.Object) *obj.Enviroment {
	env := obj.NewEnclosedEnviroment(fn.Env)

	for id, arg := range fn.Parameters {
		env.Set(arg.Value, args[id])
	}

	return env
}

func applyFunction(fn obj.Object, args []obj.Object) obj.Object {
	switch function := fn.(type) {
	case *obj.Function:
		if len(args) != len(function.Parameters) {
			return newError("wrong number of arguments. got=%d, want=%d", len(args), len(function.Parameters))
		}
		extendedEnv := extendFunctonEnv(function, args)
		eval := Eval(function.Body, extendedEnv)
		return unwrapReturnValue(eval)
	case *obj.Builtin:
		return function.Fn(args...)
	}

	return newError("not a function: %s", fn.Type())
}

func evalStringInfixExpression(operator string, left, right obj.Object) obj.Object {
	leftValue := left.(*obj.String).Value
	rightValue := right.(*obj.String).Value

	switch operator {
	case "+":
		return &obj.String{Value: leftValue + rightValue}
	case "==":
		return nattiveBoolToBooleanObject(leftValue == rightValue)
	case "!=":
		return nattiveBoolToBooleanObject(leftValue != rightValue)
	default:
		return newError("unknown infix operator: %s %s %s", left.Type(), operator, right.Type())
	}
}

func evalIndexExpression(left, index obj.Object) obj.Object {
	switch {
	case left.Type() == obj.ARRAY_OBJ && index.Type() == obj.INTEGER_OBJ:
		return evalArrayIndexExpression(left, index)
	case left.Type() == obj.HASH_OBJ:
		return evalHashIndexExpression(left, index)
	default:
		return newError("index operator not supported: %s", left.Type())
	}
}

func evalArrayIndexExpression(array, index obj.Object) obj.Object {
	arrayObj := array.(*obj.Array)
	idx := index.(*obj.Integer).Value
	max := int64(len(arrayObj.Elements) - 1)

	if idx < 0 || idx > max {
		return NULL
	}

	return arrayObj.Elements[idx]
}

func evalIndexAssignment(idx *ast.IndexExpression, opt string, val obj.Object, env *obj.Enviroment) obj.Object {
	left := Eval(idx.Left, env)
	if isError(left) {
		return left
	}

	index := Eval(idx.Index, env)
	if isError(index) {
		return index
	}

	if hash, ok := left.(*obj.Hash); ok {
		key, ok := index.(obj.Hashable)
		if !ok {
			return newError("unusable as hash key: %s", index.Type())
		}

		switch opt {
		case "=":
			hash.Pairs[key.HashKey()] = obj.HashPair{Key: index, Value: val}
		case "+=", "-=", "*=", "/=":
			pair, ok := hash.Pairs[key.HashKey()]
			if !ok {
				return newError("key not found: %s", index.Inspect())
			}
			result := evalInfixExpression(string(opt[0]), pair.Value, val)
			if isError(result) {
				return result
			}
			hash.Pairs[key.HashKey()] = obj.HashPair{Key: index, Value: result}
			return nil
		default:
			return newError("unknown assignment operator: %s", opt)
		}

		return nil
	}

	array, ok := left.(*obj.Array)
	if !ok {
		return newError("index assignment requires array or hash, got %s", left.Type())
	}

	i, ok := index.(*obj.Integer)
	if !ok {
		return newError("index assignment requires integer index, got %s", index.Type())
	}

	if i.Value < 0 || i.Value >= int64(len(array.Elements)) {
		return newError("index out of range: %d", i.Value)
	}

	switch opt {
	case "=":
		array.Elements[i.Value] = val
	case "+=", "-=", "*=", "/=":
		curr := array.Elements[i.Value]
		result := evalInfixExpression(string(opt[0]), curr, val)
		if isError(result) {
			return result
		}
		array.Elements[i.Value] = result
		return nil
	default:
		return newError("unknown assignment operator: %s", opt)
	}

	return nil
}

func evalHashLiteral(h *ast.HashLiteral, env *obj.Enviroment) obj.Object {
	pairs := make(map[obj.HashKey]obj.HashPair)

	for k, v := range h.Pairs {
		key := Eval(k, env)
		if isError(key) {
			return key
		}

		hashKey, ok := key.(obj.Hashable)
		if !ok {
			return newError("hash key must be string, got %s", key.Type())
		}

		val := Eval(v, env)
		if isError(val) {
			return val
		}

		hashed := hashKey.HashKey()

		pairs[hashed] = obj.HashPair{Key: key, Value: val}
	}
	
	return &obj.Hash{Pairs: pairs}
}

func evalHashIndexExpression(hash, index obj.Object) obj.Object {
	hashObj := hash.(*obj.Hash)

	key, ok := index.(obj.Hashable)
	if !ok {
		return newError("unusable as hash key: %s", index.Type())
	}

	pair, ok := hashObj.Pairs[key.HashKey()]
	if !ok {
		return NULL
	}

	return pair.Value
}
