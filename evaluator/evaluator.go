package evaluator

import (
	"fmt"
	"math"

	"github.com/theawakener0/Zod/ast"
	obj "github.com/theawakener0/Zod/object"
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

	case *ast.FloatLiteral:
		return &obj.Float{Value: n.Value}
	
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
			if !env.Assign(ident.Value, val) {
				return newError("identifier not found: %s", ident.Value)
			}
		case "+=":
			curr := Eval(ident, env)
			if isError(curr) {
				return curr
			}
			result := evalInfixExpression("+", curr, val)
			if isError(result) {
				return result
			}
			if !env.Assign(ident.Value, result) {
				return newError("identifier not found: %s", ident.Value)
			}
		case "-=":
			curr := Eval(ident, env)
			if isError(curr) {
				return curr
			}
			result := evalInfixExpression("-", curr, val)
			if isError(result) {
				return result
			}
			if !env.Assign(ident.Value, result) {
				return newError("identifier not found: %s", ident.Value)
			}
		case "*=":
			curr := Eval(ident, env)
			if isError(curr) {
				return curr
			}
			result := evalInfixExpression("*", curr, val)
			if isError(result) {
				return result
			}
			if !env.Assign(ident.Value, result) {
				return newError("identifier not found: %s", ident.Value)
			}
		case "/=":
			curr := Eval(ident, env)
			if isError(curr) {
				return curr
			}
			result := evalInfixExpression("/", curr, val)
			if isError(result) {
				return result
			}
			if !env.Assign(ident.Value, result) {
				return newError("identifier not found: %s", ident.Value)
			}
		}

	case *ast.Identifier:
		return evalIdentifier(n, env)

	case *ast.FunctionLiteral:
		params := n.Parameters
		body := n.Body
		return &obj.Function{Parameters: params, Env: env, Body: body}

	case *ast.CallExpression:
		if ident, ok := n.Function.(*ast.Identifier); ok && ident.Value == "try" {
			return evalTryExpression(n, env)
		}
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
	switch right.Type() {
	case obj.INTEGER_OBJ:
		value := right.(*obj.Integer).Value
		return &obj.Integer{Value: -value}
	case obj.FLOAT_OBJ:
		value := right.(*obj.Float).Value
		return &obj.Float{Value: -value}
	default:
		return newError("unknown prefix operator: -%s", right.Type())
	}
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
	case left.Type() == obj.FLOAT_OBJ || right.Type() == obj.FLOAT_OBJ:
		return evalFloatInfixExpression(operator, left, right)
	case left.Type() == obj.MATRIX_OBJ || right.Type() == obj.MATRIX_OBJ:
		return evalMatrixInfixExpression(operator, left, right)
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

func evalMatrixInfixExpression(operator string, left, right obj.Object) obj.Object {
	leftM, leftIsM := left.(*obj.Matrix)
	rightM, rightIsM := right.(*obj.Matrix)

	switch operator {
	case "+", "-", "*", "/":
		if leftIsM && rightIsM {
			return evalMatrixMatrixInfix(operator, leftM, rightM)
		}
		if leftIsM {
			return evalMatrixScalarInfix(operator, leftM, right)
		}
		return evalScalarMatrixInfix(operator, left, rightM)
	default:
		return newError("unknown infix operator: %s %s %s", left.Type(), operator, right.Type())
	}
}

func evalMatrixMatrixInfix(operator string, a, b *obj.Matrix) obj.Object {
	switch operator {
	case "+", "-":
		if a.Rows != b.Rows || a.Cols != b.Cols {
			return newError("matrix dimension mismatch for %s: %dx%d vs %dx%d", operator, a.Rows, a.Cols, b.Rows, b.Cols)
		}
		allInt := matrixIsAllInteger(a) && matrixIsAllInteger(b)
		data := make([][]obj.Object, a.Rows)
		for i := 0; i < a.Rows; i++ {
			row := make([]obj.Object, a.Cols)
			for j := 0; j < a.Cols; j++ {
				av, _ := numericValue(a.Data[i][j])
				bv, _ := numericValue(b.Data[i][j])
				var v float64
				if operator == "+" {
					v = av + bv
				} else {
					v = av - bv
				}
				row[j] = resultValue(v, allInt)
			}
			data[i] = row
		}
		return &obj.Matrix{Rows: a.Rows, Cols: a.Cols, Data: data}
	case "*":
		if a.Cols != b.Rows {
			return newError("matrix dimension mismatch for multiplication: %dx%d vs %dx%d", a.Rows, a.Cols, b.Rows, b.Cols)
		}
		allInt := matrixIsAllInteger(a) && matrixIsAllInteger(b)
		data := make([][]obj.Object, a.Rows)
		for i := 0; i < a.Rows; i++ {
			row := make([]obj.Object, b.Cols)
			for j := 0; j < b.Cols; j++ {
				var sum float64
				for k := 0; k < a.Cols; k++ {
					av, _ := numericValue(a.Data[i][k])
					bv, _ := numericValue(b.Data[k][j])
					sum += av * bv
				}
				row[j] = resultValue(sum, allInt)
			}
			data[i] = row
		}
		return &obj.Matrix{Rows: a.Rows, Cols: b.Cols, Data: data}
	default:
		return newError("unknown infix operator: %s %s %s", a.Type(), operator, b.Type())
	}
}

func evalMatrixScalarInfix(operator string, m *obj.Matrix, scalar obj.Object) obj.Object {
	if _, ok := numericValue(scalar); !ok {
		return newError("matrix arithmetic requires numeric operand, got %s", scalar.Type())
	}
	allInt := matrixIsAllInteger(m) && scalarIsInteger(scalar)
	data := make([][]obj.Object, m.Rows)
	for i := 0; i < m.Rows; i++ {
		row := make([]obj.Object, m.Cols)
		for j := 0; j < m.Cols; j++ {
			mv, _ := numericValue(m.Data[i][j])
			sv, _ := numericValue(scalar)
			var v float64
			switch operator {
			case "+":
				v = mv + sv
			case "-":
				v = mv - sv
			case "*":
				v = mv * sv
			case "/":
				if sv == 0 {
					return newError("division by zero")
				}
				v = mv / sv
			}
			row[j] = resultValue(v, allInt)
		}
		data[i] = row
	}
	return &obj.Matrix{Rows: m.Rows, Cols: m.Cols, Data: data}
}

func evalScalarMatrixInfix(operator string, scalar obj.Object, m *obj.Matrix) obj.Object {
	if _, ok := numericValue(scalar); !ok {
		return newError("matrix arithmetic requires numeric operand, got %s", scalar.Type())
	}
	allInt := scalarIsInteger(scalar) && matrixIsAllInteger(m)
	data := make([][]obj.Object, m.Rows)
	for i := 0; i < m.Rows; i++ {
		row := make([]obj.Object, m.Cols)
		for j := 0; j < m.Cols; j++ {
			mv, _ := numericValue(m.Data[i][j])
			sv, _ := numericValue(scalar)
			var v float64
			switch operator {
			case "+":
				v = sv + mv
			case "-":
				v = sv - mv
			case "*":
				v = sv * mv
			case "/":
				if mv == 0 {
					return newError("division by zero")
				}
				v = sv / mv
			}
			row[j] = resultValue(v, allInt)
		}
		data[i] = row
	}
	return &obj.Matrix{Rows: m.Rows, Cols: m.Cols, Data: data}
}

func numericValue(o obj.Object) (float64, bool) {
	switch v := o.(type) {
	case *obj.Integer:
		return float64(v.Value), true
	case *obj.Float:
		return v.Value, true
	}
	return 0, false
}

func resultValue(v float64, isInt bool) obj.Object {
	if isInt && v == math.Trunc(v) {
		return &obj.Integer{Value: int64(v)}
	}
	return &obj.Float{Value: v}
}

func matrixIsAllInteger(m *obj.Matrix) bool {
	for i := 0; i < m.Rows; i++ {
		for j := 0; j < m.Cols; j++ {
			if _, ok := m.Data[i][j].(*obj.Integer); !ok {
				return false
			}
		}
	}
	return true
}

func scalarIsInteger(o obj.Object) bool {
	_, ok := o.(*obj.Integer)
	return ok
}

func evalFloatInfixExpression(operator string, left, right obj.Object) obj.Object {
	var leftValue, rightValue float64

	switch l := left.(type) {
	case *obj.Float:
		leftValue = l.Value
	case *obj.Integer:
		leftValue = float64(l.Value)
	}

	switch r := right.(type) {
	case *obj.Float:
		rightValue = r.Value
	case *obj.Integer:
		rightValue = float64(r.Value)
	}

	switch operator {
	case "+":
		return &obj.Float{Value: leftValue + rightValue}
	case "-":
		return &obj.Float{Value: leftValue - rightValue}
	case "*":
		return &obj.Float{Value: leftValue * rightValue}
	case "/":
		if rightValue == 0 {
			return newError("division by zero")
		}
		return &obj.Float{Value: leftValue / rightValue}
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
	loopEnv := obj.NewEnclosedEnviroment(env)

	if fe.Init != nil {
		result := Eval(fe.Init, loopEnv)
		if isError(result) {
			return result
		}
	}

	for {
		var condition obj.Object
		if fe.Condition != nil {
			condition = Eval(fe.Condition, loopEnv)
		} else {
			condition = TRUE
		}

		if isError(condition) {
			return condition
		}

		if !isTruthy(condition) {
			break
		}

		bodyEnv := obj.NewEnclosedEnviroment(loopEnv)
		result := Eval(fe.Body, bodyEnv)
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
			result := Eval(fe.Update, loopEnv)
			if isError(result) {
				return result
			}
		}
	}

	return NULL
}

func evalLoopExpression(le *ast.LoopExpression, env *obj.Enviroment) obj.Object {
	for {
		bodyEnv := obj.NewEnclosedEnviroment(env)
		result := Eval(le.Body, bodyEnv)
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

		switch val.Type() {
	case obj.INTEGER_OBJ:
		intVal := val.(*obj.Integer).Value
		oldObj := &obj.Integer{Value: intVal}
		if opt == "++" {
			intVal++
		} else {
			intVal--
		}

		newObj := &obj.Integer{Value: intVal}
		env.Assign(ident.Value, newObj)

		if isPostfix {
			return oldObj
		}
		return newObj
	case obj.FLOAT_OBJ:
		floatVal := val.(*obj.Float).Value
		oldObj := &obj.Float{Value: floatVal}
		if opt == "++" {
			floatVal++
		} else {
			floatVal--
		}

		newObj := &obj.Float{Value: floatVal}
		env.Assign(ident.Value, newObj)

		if isPostfix {
			return oldObj
		}
		return newObj
	default:
		return newError("%s requires integer or float, got %s", opt, val.Type())
	}
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
	blockEnv := obj.NewEnclosedEnviroment(env)

	for _, stmt := range b.Statements {
		result = Eval(stmt, blockEnv)

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

func evalTryExpression(ce *ast.CallExpression, env *obj.Enviroment) obj.Object {
	if len(ce.Arguments) != 1 {
		return newError("wrong number of arguments. got=%d, want=1", len(ce.Arguments))
	}

	result := Eval(ce.Arguments[0], env)
	if isError(result) {
		msg := result.Inspect()
		if err, ok := result.(*obj.Error); ok {
			msg = err.Message
		}
		return &obj.Array{Elements: []obj.Object{FALSE, &obj.String{Value: msg}}}
	}
	return &obj.Array{Elements: []obj.Object{TRUE, result}}
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
	case left.Type() == obj.MATRIX_OBJ && index.Type() == obj.INTEGER_OBJ:
		return evalMatrixIndexExpression(left, index)
	case left.Type() == obj.HASH_OBJ:
		return evalHashIndexExpression(left, index)
	default:
		return newError("index operator not supported: %s", left.Type())
	}
}

func evalMatrixIndexExpression(matrix, index obj.Object) obj.Object {
	matrixObj := matrix.(*obj.Matrix)
	idx := index.(*obj.Integer).Value

	if idx < 0 || idx >= int64(matrixObj.Rows) {
		return NULL
	}

	return &obj.Array{Elements: matrixObj.Data[idx]}
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

		hashKey := key.HashKey()

		switch opt {
		case "=":
			if _, exists := hash.Pairs[hashKey]; !exists {
				hash.Order = append(hash.Order, hashKey)
			}
			hash.Pairs[hashKey] = obj.HashPair{Key: index, Value: val}
		case "+=", "-=", "*=", "/=":
			pair, ok := hash.Pairs[hashKey]
			if !ok {
				return newError("key not found: %s", index.Inspect())
			}
			result := evalInfixExpression(string(opt[0]), pair.Value, val)
			if isError(result) {
				return result
			}
			hash.Pairs[hashKey] = obj.HashPair{Key: index, Value: result}
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
	order := make([]obj.HashKey, 0, len(h.Pairs))

	for _, p := range h.Pairs {
		key := Eval(p.Key, env)
		if isError(key) {
			return key
		}

		hashKey, ok := key.(obj.Hashable)
		if !ok {
			return newError("hash key must be string, got %s", key.Type())
		}

		val := Eval(p.Value, env)
		if isError(val) {
			return val
		}

		hashed := hashKey.HashKey()

		if _, exists := pairs[hashed]; !exists {
			order = append(order, hashed)
		}
		pairs[hashed] = obj.HashPair{Key: key, Value: val}
	}
	
	return &obj.Hash{Pairs: pairs, Order: order}
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
