package vm

import (
	"encoding/binary"
	"fmt"
	"math"

	"github.com/theawakener0/Zod/code"
	"github.com/theawakener0/Zod/compiler"
	obj "github.com/theawakener0/Zod/object"
)

const StackSize = 2048
const GlobalsSize = 65536
const MaxFrames = 1024

var True = &obj.Boolean{Value: true}
var False = &obj.Boolean{Value: false}
var Null = &obj.Null{}

type VM struct {
	constant []obj.Object

	stack []obj.Object
	sp    int

	globals []obj.Object

	frames     []*Frame
	frameIndex int
}

func New(bytecode *compiler.Bytecode) *VM {
	mainFn := &obj.CompiledFunction{Instructions: bytecode.Instructions}
	mainClosure := &obj.Closure{Fn: mainFn}
	mainFrame := NewFrame(mainClosure, 0)

	frames := make([]*Frame, MaxFrames)
	frames[0] = mainFrame

	return &VM{
		constant: bytecode.Constant,

		stack: make([]obj.Object, StackSize),
		sp:    0,

		globals: make([]obj.Object, GlobalsSize),

		frames:     frames,
		frameIndex: 1,
	}
}

func (vm *VM) StackTop() obj.Object {
	if vm.sp == 0 {
		return nil
	}
	return vm.stack[vm.sp-1]
}

func (vm *VM) LastPoppedStackElem() obj.Object {
	return vm.stack[vm.sp]
}

func (vm *VM) Run() error {
	var (
		ip  int
		ins code.Instructions
		op  code.Opcode
	)

	for vm.currentFrame().ip < len(vm.currentFrame().Instructions())-1 {
		vm.currentFrame().ip++

		ip = vm.currentFrame().ip
		ins = vm.currentFrame().Instructions()
		op = code.Opcode(ins[ip])

		switch op {
		case code.OpConstant:
			constIndex := binary.BigEndian.Uint16(ins[ip+1:])
			vm.currentFrame().ip += 2

			err := vm.push(vm.constant[constIndex])
			if err != nil {
				return err
			}
		case code.OpAdd, code.OpSub, code.OpMul, code.OpDiv:
			right := vm.pop()
			left := vm.pop()

			result := executeBinaryOperation(op, left, right)

			err := vm.push(result)
			if err != nil {
				return err
			}
		case code.OpPop:
			vm.pop()
		case code.OpTrue:
			err := vm.push(True)
			if err != nil {
				return err
			}
		case code.OpFalse:
			err := vm.push(False)
			if err != nil {
				return err
			}
		case code.OpEqual, code.OpNotEqual, code.OpGreaterThan, code.OpGreaterThanEqual:
			err := vm.executeComparison(op)
			if err != nil {
				return err
			}
		case code.OpBang:
			err := vm.executeBangOperator()
			if err != nil {
				return err
			}
		case code.OpMinus:
			err := vm.executeMinusOperator()
			if err != nil {
				return err
			}
		case code.OpJump:
			pos := int(binary.BigEndian.Uint16(ins[ip+1:]))
			vm.currentFrame().ip = pos - 1
		case code.OpJumpNotTruthy:
			pos := int(binary.BigEndian.Uint16(ins[ip+1:]))
			vm.currentFrame().ip += 2

			condition := vm.pop()
			if !isTruthy(condition) {
				vm.currentFrame().ip = pos - 1
			}
		case code.OpNull:
			err := vm.push(Null)
			if err != nil {
				return err
			}
		case code.OpSetGlobal:
			globalIndex := binary.BigEndian.Uint16(ins[ip+1:])
			vm.currentFrame().ip += 2

			vm.globals[globalIndex] = vm.pop()
		case code.OpSetLocal:
			localIndex := int(ins[ip+1])
			vm.currentFrame().ip += 1

			frame := vm.currentFrame()

			v := vm.pop()
			if cell, ok := vm.stack[frame.basePointer+int(localIndex)].(*obj.Cell); ok && cell != nil {
				cell.Value = v
			} else {
				vm.stack[frame.basePointer+int(localIndex)] = v
			}
		case code.OpDefineLocal:
			localIndex := int(ins[ip+1])
			vm.currentFrame().ip += 1

			frame := vm.currentFrame()

			v := vm.pop()
			if v == nil {
				v = Null
			}
			vm.stack[frame.basePointer+int(localIndex)] = obj.NewCell(v)
		case code.OpSetFree:
			freeIndex := code.ReadUnit8(ins[ip+1:])
			vm.currentFrame().ip += 1

			currentClosure := vm.currentFrame().cl
			if int(freeIndex) < 0 || int(freeIndex) >= len(currentClosure.Free) {
				return fmt.Errorf("free variable index out of range: %d", freeIndex)
			}
			v := vm.pop()
			if cell, ok := currentClosure.Free[freeIndex].(*obj.Cell); ok && cell != nil {
				cell.Value = v
			} else {
				currentClosure.Free[freeIndex] = v
			}
		case code.OpSetIndex:
			kind := int(code.ReadUnit8(ins[ip+1:]))
			vm.currentFrame().ip += 1

			if kind >= 10 {
				err := vm.executeDoubleIndexAssign(kind - 10)
				if err != nil {
					return err
				}
			} else {
				err := vm.executeSingleIndexAssign(kind)
				if err != nil {
					return err
				}
			}
		case code.OpTry:
			err := vm.executeTry()
			if err != nil {
				return err
			}
		case code.OpGetGlobal:
			globalIndex := binary.BigEndian.Uint16(ins[ip+1:])
			vm.currentFrame().ip += 2

			val := vm.globals[globalIndex]
			if val == nil {
				val = Null
			}
			err := vm.push(val)
			if err != nil {
				return err
			}
		case code.OpGetLocal:
			localIndex := int(ins[ip+1])
			vm.currentFrame().ip += 1

			frame := vm.currentFrame()

			err := vm.push(obj.Deref(vm.stack[frame.basePointer+int(localIndex)]))
			if err != nil {
				return err
			}
		case code.OpGetLocalCell:
			localIndex := int(ins[ip+1])
			vm.currentFrame().ip += 1

			frame := vm.currentFrame()

			slot := vm.stack[frame.basePointer+int(localIndex)]
			if cell, ok := slot.(*obj.Cell); ok && cell != nil {
				err := vm.push(cell)
				if err != nil {
					return err
				}
			} else if slot == nil {
				cell := obj.NewCell(Null)
				vm.stack[frame.basePointer+int(localIndex)] = cell
				err := vm.push(cell)
				if err != nil {
					return err
				}
			} else {
				cell := obj.NewCell(slot)
				vm.stack[frame.basePointer+int(localIndex)] = cell
				err := vm.push(cell)
				if err != nil {
					return err
				}
			}
		case code.OpArray:
			numElements := int(binary.BigEndian.Uint16(ins[ip+1:]))
			vm.currentFrame().ip += 2

			array := vm.buildArray(vm.sp-numElements, vm.sp)
			vm.sp -= numElements

			err := vm.push(array)
			if err != nil {
				return err
			}
		case code.OpHash:
			numElements := int(binary.BigEndian.Uint16(ins[ip+1:]))
			vm.currentFrame().ip += 2

			hash, errObj := vm.buildHash(vm.sp-numElements, vm.sp)
			vm.sp -= numElements

			if errObj != nil {
				err := vm.push(errObj)
				if err != nil {
					return err
				}
			} else {
				err := vm.push(hash)
				if err != nil {
					return err
				}
			}
		case code.OpIndex:
			index := vm.pop()
			left := vm.pop()

			err := vm.executeIndexExpression(left, index)
			if err != nil {
				return err
			}
		case code.OpCall:
			numArgs := code.ReadUnit8(ins[ip+1:])
			vm.currentFrame().ip += 1

			err := vm.executeCall(int(numArgs))
			if err != nil {
				return err
			}
		case code.OpReturnValue:
			returnValue := vm.pop()

			frame := vm.popFrame()
			vm.sp = frame.basePointer - 1

			err := vm.push(returnValue)
			if err != nil {
				return err
			}
		case code.OpReturn:
			frame := vm.popFrame()
			vm.sp = frame.basePointer - 1

			err := vm.push(Null)
			if err != nil {
				return err
			}
		case code.OpGetBuiltin:
			builtinIndex := code.ReadUnit8(ins[ip+1:])
			vm.currentFrame().ip += 1

			definition := obj.Builtins[builtinIndex]

			err := vm.push(definition.Builtin)
			if err != nil {
				return err
			}
		case code.OpClosure:
			constIndex := binary.BigEndian.Uint16(ins[ip+1:])
			numFree := code.ReadUnit8(ins[ip+3:])
			vm.currentFrame().ip += 3

			err := vm.pushClosure(int(constIndex), int(numFree))
			if err != nil {
				return err
			}
		case code.OpGetFree:
			freeIndex := code.ReadUnit8(ins[ip+1:])
			vm.currentFrame().ip += 1

			currentClosure := vm.currentFrame().cl

			err := vm.push(obj.Deref(currentClosure.Free[freeIndex]))
			if err != nil {
				return err
			}
		case code.OpGetFreeCell:
			freeIndex := code.ReadUnit8(ins[ip+1:])
			vm.currentFrame().ip += 1

			currentClosure := vm.currentFrame().cl
			if int(freeIndex) < 0 || int(freeIndex) >= len(currentClosure.Free) {
				return fmt.Errorf("free variable index out of range: %d", freeIndex)
			}

			slot := currentClosure.Free[freeIndex]
			if cell, ok := slot.(*obj.Cell); ok && cell != nil {
				err := vm.push(cell)
				if err != nil {
					return err
				}
			} else if slot == nil {
				cell := obj.NewCell(Null)
				currentClosure.Free[freeIndex] = cell
				err := vm.push(cell)
				if err != nil {
					return err
				}
			} else {
				cell := obj.NewCell(slot)
				currentClosure.Free[freeIndex] = cell
				err := vm.push(cell)
				if err != nil {
					return err
				}
			}
		case code.OpCurrentClosure:
			currentClosure := vm.currentFrame().cl

			err := vm.push(currentClosure)
			if err != nil {
				return err
			}
		case code.OpDup:
			if vm.sp == 0 {
				return fmt.Errorf("stack underflow on OpDup")
			}

			err := vm.push(vm.stack[vm.sp-1])
			if err != nil {
				return err
			}
		}

	}

	return nil
}

func (vm *VM) push(o obj.Object) error {
	if vm.sp >= StackSize {
		return fmt.Errorf("stack overflow :(")
	}

	vm.stack[vm.sp] = o
	vm.sp++

	return nil
}

func (vm *VM) pop() obj.Object {
	o := vm.stack[vm.sp-1]
	vm.sp--
	return o
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
		if len(m.Data[i]) != m.Cols {
			return false
		}
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

func opToString(op code.Opcode) string {
	switch op {
	case code.OpAdd:
		return "+"
	case code.OpSub:
		return "-"
	case code.OpMul:
		return "*"
	case code.OpDiv:
		return "/"
	default:
		return "?"
	}
}

func executeBinaryOperation(op code.Opcode, left, right obj.Object) obj.Object {
	if left == nil {
		left = Null
	}
	if right == nil {
		right = Null
	}
	if errObj, ok := left.(*obj.Error); ok {
		return errObj
	}
	if errObj, ok := right.(*obj.Error); ok {
		return errObj
	}
	if left.Type() == obj.MATRIX_OBJ || right.Type() == obj.MATRIX_OBJ {
		return evalMatrixInfix(opToString(op), left, right)
	}
	if left.Type() == obj.INTEGER_OBJ && right.Type() == obj.INTEGER_OBJ {
		leftVal := left.(*obj.Integer).Value
		rightVal := right.(*obj.Integer).Value

		switch op {
		case code.OpAdd:
			return &obj.Integer{Value: leftVal + rightVal}
		case code.OpSub:
			return &obj.Integer{Value: leftVal - rightVal}
		case code.OpMul:
			return &obj.Integer{Value: leftVal * rightVal}
		case code.OpDiv:
			if rightVal == 0 {
				return &obj.Error{Message: "division by zero"}
			}
			return &obj.Integer{Value: leftVal / rightVal}
		}
	}
	if left.Type() == obj.STRING_OBJ && right.Type() == obj.STRING_OBJ {
		leftVal := left.(*obj.String).Value
		rightVal := right.(*obj.String).Value

		if op == code.OpAdd {
			return &obj.String{Value: leftVal + rightVal}
		}
		return &obj.Error{Message: fmt.Sprintf("unknown infix operator: %s %s %s", left.Type(), opToString(op), right.Type())}
	}

	leftVal, leftOk := numericValue(left)
	rightVal, rightOk := numericValue(right)
	if !leftOk || !rightOk {
		return &obj.Error{Message: fmt.Sprintf("unknown infix operator: %s %s %s", left.Type(), opToString(op), right.Type())}
	}

	switch op {
	case code.OpAdd:
		return &obj.Float{Value: leftVal + rightVal}
	case code.OpSub:
		return &obj.Float{Value: leftVal - rightVal}
	case code.OpMul:
		return &obj.Float{Value: leftVal * rightVal}
	case code.OpDiv:
		if rightVal == 0 {
			return &obj.Error{Message: "division by zero"}
		}
		return &obj.Float{Value: leftVal / rightVal}
	}

	return &obj.Error{Message: fmt.Sprintf("unknown operation: %d", op)}
}

func evalMatrixInfix(operator string, left, right obj.Object) obj.Object {
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
	case "==", "!=":
		if leftIsM && rightIsM {
			eq := leftM.Rows == rightM.Rows && leftM.Cols == rightM.Cols
			if eq {
				for i := 0; i < leftM.Rows; i++ {
					for j := 0; j < leftM.Cols; j++ {
						av, _ := numericValue(leftM.Data[i][j])
						bv, _ := numericValue(rightM.Data[i][j])
						if av != bv {
							eq = false
							break
						}
					}
					if !eq {
						break
					}
				}
			}
			if operator == "==" {
				return nativeBoolToBooleanObj(eq)
			}
			return nativeBoolToBooleanObj(!eq)
		}
		if operator == "==" {
			return False
		}
		return True
	default:
		return &obj.Error{Message: fmt.Sprintf("unknown infix operator: %s %s %s", left.Type(), operator, right.Type())}
	}
}

func evalMatrixMatrixInfix(operator string, a, b *obj.Matrix) obj.Object {
	switch operator {
	case "+", "-":
		if a.Rows != b.Rows || a.Cols != b.Cols {
			return &obj.Error{Message: fmt.Sprintf("matrix dimension mismatch for %s: %dx%d vs %dx%d", operator, a.Rows, a.Cols, b.Rows, b.Cols)}
		}
		allInt := matrixIsAllInteger(a) && matrixIsAllInteger(b)
		data := make([][]obj.Object, a.Rows)
		for i := 0; i < a.Rows; i++ {
			row := make([]obj.Object, a.Cols)
			for j := 0; j < a.Cols; j++ {
				av, ok := numericValue(a.Data[i][j])
				if !ok {
					return &obj.Error{Message: fmt.Sprintf("matrix element not numeric, got %s", a.Data[i][j].Type())}
				}
				bv, ok := numericValue(b.Data[i][j])
				if !ok {
					return &obj.Error{Message: fmt.Sprintf("matrix element not numeric, got %s", b.Data[i][j].Type())}
				}
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
			return &obj.Error{Message: fmt.Sprintf("matrix dimension mismatch for multiplication: %dx%d vs %dx%d", a.Rows, a.Cols, b.Rows, b.Cols)}
		}
		allInt := matrixIsAllInteger(a) && matrixIsAllInteger(b)
		data := make([][]obj.Object, a.Rows)
		for i := 0; i < a.Rows; i++ {
			row := make([]obj.Object, b.Cols)
			for j := 0; j < b.Cols; j++ {
				var sum float64
				for k := 0; k < a.Cols; k++ {
					av, ok := numericValue(a.Data[i][k])
					if !ok {
						return &obj.Error{Message: fmt.Sprintf("matrix element not numeric, got %s", a.Data[i][k].Type())}
					}
					bv, ok := numericValue(b.Data[k][j])
					if !ok {
						return &obj.Error{Message: fmt.Sprintf("matrix element not numeric, got %s", b.Data[k][j].Type())}
					}
					sum += av * bv
				}
				row[j] = resultValue(sum, allInt)
			}
			data[i] = row
		}
		return &obj.Matrix{Rows: a.Rows, Cols: b.Cols, Data: data}
	case "/":
		return &obj.Error{Message: "matrix division not supported: use scalar division"}
	default:
		return &obj.Error{Message: fmt.Sprintf("unknown infix operator: %s %s %s", a.Type(), operator, b.Type())}
	}
}

func evalMatrixScalarInfix(operator string, m *obj.Matrix, scalar obj.Object) obj.Object {
	if _, ok := numericValue(scalar); !ok {
		return &obj.Error{Message: fmt.Sprintf("matrix arithmetic requires numeric operand, got %s", scalar.Type())}
	}
	allInt := matrixIsAllInteger(m) && scalarIsInteger(scalar)
	data := make([][]obj.Object, m.Rows)
	for i := 0; i < m.Rows; i++ {
		row := make([]obj.Object, m.Cols)
		for j := 0; j < m.Cols; j++ {
			mv, ok := numericValue(m.Data[i][j])
			if !ok {
				return &obj.Error{Message: fmt.Sprintf("matrix element not numeric, got %s", m.Data[i][j].Type())}
			}
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
					return &obj.Error{Message: "division by zero"}
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
		return &obj.Error{Message: fmt.Sprintf("matrix arithmetic requires numeric operand, got %s", scalar.Type())}
	}
	allInt := scalarIsInteger(scalar) && matrixIsAllInteger(m)
	data := make([][]obj.Object, m.Rows)
	for i := 0; i < m.Rows; i++ {
		row := make([]obj.Object, m.Cols)
		for j := 0; j < m.Cols; j++ {
			mv, ok := numericValue(m.Data[i][j])
			if !ok {
				return &obj.Error{Message: fmt.Sprintf("matrix element not numeric, got %s", m.Data[i][j].Type())}
			}
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
					return &obj.Error{Message: "division by zero"}
				}
				v = sv / mv
			}
			row[j] = resultValue(v, allInt)
		}
		data[i] = row
	}
	return &obj.Matrix{Rows: m.Rows, Cols: m.Cols, Data: data}
}

func nativeBoolToBooleanObj(input bool) *obj.Boolean {
	if input {
		return True
	}
	return False
}

func (vm *VM) executeBinaryComparison(op code.Opcode, left, right obj.Object) error {
	if left == nil {
		left = Null
	}
	if right == nil {
		right = Null
	}
	if errObj, ok := left.(*obj.Error); ok {
		return vm.push(errObj)
	}
	if errObj, ok := right.(*obj.Error); ok {
		return vm.push(errObj)
	}
	if left.Type() == obj.INTEGER_OBJ && right.Type() == obj.INTEGER_OBJ {
		leftVal := left.(*obj.Integer).Value
		rightVal := right.(*obj.Integer).Value

		switch op {
		case code.OpEqual:
			return vm.push(nativeBoolToBooleanObj(rightVal == leftVal))
		case code.OpNotEqual:
			return vm.push(nativeBoolToBooleanObj(rightVal != leftVal))
		case code.OpGreaterThan:
			return vm.push(nativeBoolToBooleanObj(leftVal > rightVal))
		case code.OpGreaterThanEqual:
			return vm.push(nativeBoolToBooleanObj(leftVal >= rightVal))
		}
	}
	if left.Type() == obj.STRING_OBJ && right.Type() == obj.STRING_OBJ {
		leftVal := left.(*obj.String).Value
		rightVal := right.(*obj.String).Value
		switch op {
		case code.OpEqual:
			return vm.push(nativeBoolToBooleanObj(leftVal == rightVal))
		case code.OpNotEqual:
			return vm.push(nativeBoolToBooleanObj(leftVal != rightVal))
		default:
			return vm.push(&obj.Error{Message: fmt.Sprintf("unknown operator: %d (%s %s)", op, left.Type(), right.Type())})
		}
	}
	if left.Type() == obj.MATRIX_OBJ || right.Type() == obj.MATRIX_OBJ {
		opStr := "=="
		switch op {
		case code.OpEqual:
			opStr = "=="
		case code.OpNotEqual:
			opStr = "!="
		default:
			return vm.push(&obj.Error{Message: fmt.Sprintf("unknown operator: %d (%s %s)", op, left.Type(), right.Type())})
		}
		return vm.push(evalMatrixInfix(opStr, left, right))
	}

	leftVal, leftOk := numericValue(left)
	rightVal, rightOk := numericValue(right)
	if !leftOk || !rightOk {
		if op == code.OpEqual || op == code.OpNotEqual {
			eq := left == right
			if left.Type() == obj.BOOLEAN_OBJ && right.Type() == obj.BOOLEAN_OBJ {
				eq = left.(*obj.Boolean).Value == right.(*obj.Boolean).Value
			} else if left.Type() == obj.NULL_OBJ && right.Type() == obj.NULL_OBJ {
				eq = true
			}
			if op == code.OpEqual {
				return vm.push(nativeBoolToBooleanObj(eq))
			}
			return vm.push(nativeBoolToBooleanObj(!eq))
		}
		return vm.push(&obj.Error{Message: fmt.Sprintf("unknown operator: %d (%s %s)", op, left.Type(), right.Type())})
	}

	switch op {
	case code.OpEqual:
		return vm.push(nativeBoolToBooleanObj(rightVal == leftVal))
	case code.OpNotEqual:
		return vm.push(nativeBoolToBooleanObj(rightVal != leftVal))
	case code.OpGreaterThan:
		return vm.push(nativeBoolToBooleanObj(leftVal > rightVal))
	case code.OpGreaterThanEqual:
		return vm.push(nativeBoolToBooleanObj(leftVal >= rightVal))
	}

	return vm.push(&obj.Error{Message: fmt.Sprintf("unknown operation: %d", op)})
}

func (vm *VM) executeComparison(op code.Opcode) error {
	right := vm.pop()
	left := vm.pop()

	if left == nil {
		left = Null
	}
	if right == nil {
		right = Null
	}
	if errObj, ok := left.(*obj.Error); ok {
		return vm.push(errObj)
	}
	if errObj, ok := right.(*obj.Error); ok {
		return vm.push(errObj)
	}

	if (left.Type() == obj.INTEGER_OBJ || left.Type() == obj.FLOAT_OBJ) && (right.Type() == obj.INTEGER_OBJ || right.Type() == obj.FLOAT_OBJ) {
		return vm.executeBinaryComparison(op, left, right)
	}
	if left.Type() == obj.STRING_OBJ && right.Type() == obj.STRING_OBJ {
		return vm.executeBinaryComparison(op, left, right)
	}
	if left.Type() == obj.MATRIX_OBJ || right.Type() == obj.MATRIX_OBJ {
		return vm.executeBinaryComparison(op, left, right)
	}

	switch op {
	case code.OpEqual:
		if left.Type() == obj.BOOLEAN_OBJ && right.Type() == obj.BOOLEAN_OBJ {
			return vm.push(nativeBoolToBooleanObj(left.(*obj.Boolean).Value == right.(*obj.Boolean).Value))
		}
		if left.Type() == obj.NULL_OBJ && right.Type() == obj.NULL_OBJ {
			return vm.push(nativeBoolToBooleanObj(true))
		}
		return vm.push(nativeBoolToBooleanObj(false))
	case code.OpNotEqual:
		if left.Type() == obj.BOOLEAN_OBJ && right.Type() == obj.BOOLEAN_OBJ {
			return vm.push(nativeBoolToBooleanObj(left.(*obj.Boolean).Value != right.(*obj.Boolean).Value))
		}
		if left.Type() == obj.NULL_OBJ && right.Type() == obj.NULL_OBJ {
			return vm.push(nativeBoolToBooleanObj(false))
		}
		if (left.Type() == obj.NULL_OBJ) != (right.Type() == obj.NULL_OBJ) {
			return vm.push(nativeBoolToBooleanObj(true))
		}
		return vm.push(nativeBoolToBooleanObj(left != right))
	default:
		return vm.push(&obj.Error{Message: fmt.Sprintf("unknown operator: %d (%s %s)", op, left.Type(), right.Type())})
	}
}

func (vm *VM) executeBangOperator() error {
	operand := vm.pop()
	if operand == nil {
		operand = Null
	}
	if errObj, ok := operand.(*obj.Error); ok {
		return vm.push(errObj)
	}

	switch operand {
	case True:
		return vm.push(False)
	case False:
		return vm.push(True)
	case Null:
		return vm.push(True)
	default:
		if b, ok := operand.(*obj.Boolean); ok {
			if b.Value {
				return vm.push(False)
			}
			return vm.push(True)
		}
		if _, ok := operand.(*obj.Null); ok {
			return vm.push(True)
		}
		return vm.push(False)
	}
}

func (vm *VM) executeMinusOperator() error {
	operand := vm.pop()
	if operand == nil {
		operand = Null
	}
	if errObj, ok := operand.(*obj.Error); ok {
		return vm.push(errObj)
	}

	switch operand := operand.(type) {
	case *obj.Integer:
		return vm.push(&obj.Integer{Value: -operand.Value})
	case *obj.Float:
		return vm.push(&obj.Float{Value: -operand.Value})
	}

	return vm.push(&obj.Error{Message: fmt.Sprintf("unknown prefix operator: -%s", operand.Type())})
}

func (vm *VM) buildArray(startIndex, endIndex int) obj.Object {
	elements := make([]obj.Object, endIndex-startIndex)

	for i := startIndex; i < endIndex; i++ {
		elements[i-startIndex] = vm.stack[i]
		if elements[i-startIndex] == nil {
			elements[i-startIndex] = Null
		}
	}

	return &obj.Array{Elements: elements}
}

func (vm *VM) buildHash(startIndex, endIndex int) (obj.Object, obj.Object) {
	hashedPairs := make(map[obj.HashKey]obj.HashPair)
	order := make([]obj.HashKey, 0, (endIndex-startIndex)/2)

	for i := startIndex; i < endIndex; i += 2 {
		key := vm.stack[i]
		value := vm.stack[i+1]
		if key == nil {
			key = Null
		}
		if value == nil {
			value = Null
		}
		if errObj, ok := key.(*obj.Error); ok {
			return nil, errObj
		}
		if errObj, ok := value.(*obj.Error); ok {
			return nil, errObj
		}

		pair := obj.HashPair{Key: key, Value: value}

		hashKey, ok := key.(obj.Hashable)
		if !ok {
			return nil, &obj.Error{Message: fmt.Sprintf("unusable as hash key: %s", key.Type())}
		}

		hashed := hashKey.HashKey()
		if _, exists := hashedPairs[hashed]; !exists {
			order = append(order, hashed)
		}
		hashedPairs[hashed] = pair
	}

	return &obj.Hash{Pairs: hashedPairs, Order: order}, nil
}

func (vm *VM) executeArrayIndex(array, index obj.Object) error {
	if array == nil {
		array = Null
	}
	if index == nil {
		index = Null
	}
	if errObj, ok := array.(*obj.Error); ok {
		return vm.push(errObj)
	}
	if errObj, ok := index.(*obj.Error); ok {
		return vm.push(errObj)
	}
	arrayObj := array.(*obj.Array)
	i, ok := index.(*obj.Integer)
	if !ok {
		return vm.push(&obj.Error{Message: fmt.Sprintf("index operator requires integer index, got %s", index.Type())})
	}
	max0 := int64(len(arrayObj.Elements) - 1)

	if i.Value < 0 || i.Value > max0 {
		return vm.push(Null)
	}

	return vm.push(arrayObj.Elements[i.Value])
}

func (vm *VM) executeMatrixIndex(matrix, index obj.Object) error {
	if errObj, ok := matrix.(*obj.Error); ok {
		return vm.push(errObj)
	}
	if errObj, ok := index.(*obj.Error); ok {
		return vm.push(errObj)
	}
	matrixObj := matrix.(*obj.Matrix)
	i, ok := index.(*obj.Integer)
	if !ok {
		return vm.push(&obj.Error{Message: fmt.Sprintf("index operator requires integer index, got %s", index.Type())})
	}
	if i.Value < 0 || i.Value >= int64(matrixObj.Rows) {
		return vm.push(Null)
	}
	row := matrixObj.Data[i.Value]
	copied := make([]obj.Object, len(row))
	copy(copied, row)
	return vm.push(&obj.Array{Elements: copied})
}

func (vm *VM) executeHashIndex(hash, index obj.Object) error {
	if errObj, ok := hash.(*obj.Error); ok {
		return vm.push(errObj)
	}
	if errObj, ok := index.(*obj.Error); ok {
		return vm.push(errObj)
	}
	hashObj := hash.(*obj.Hash)

	key, ok := index.(obj.Hashable)
	if !ok {
		return vm.push(&obj.Error{Message: fmt.Sprintf("unusable as hash key: %s", index.Type())})
	}

	pair, ok := hashObj.Pairs[key.HashKey()]
	if !ok {
		return vm.push(Null)
	}

	return vm.push(pair.Value)

}

func (vm *VM) executeIndexExpression(left, index obj.Object) error {
	if left == nil {
		left = Null
	}
	if index == nil {
		index = Null
	}
	if errObj, ok := left.(*obj.Error); ok {
		return vm.push(errObj)
	}
	if errObj, ok := index.(*obj.Error); ok {
		return vm.push(errObj)
	}
	switch {
	case left.Type() == obj.ARRAY_OBJ && index.Type() == obj.INTEGER_OBJ:
		return vm.executeArrayIndex(left, index)
	case left.Type() == obj.MATRIX_OBJ && index.Type() == obj.INTEGER_OBJ:
		return vm.executeMatrixIndex(left, index)
	case left.Type() == obj.ARRAY_OBJ:
		return vm.push(&obj.Error{Message: fmt.Sprintf("index operator requires integer index, got %s", index.Type())})
	case left.Type() == obj.MATRIX_OBJ:
		return vm.push(&obj.Error{Message: fmt.Sprintf("index operator requires integer index, got %s", index.Type())})
	case left.Type() == obj.HASH_OBJ:
		return vm.executeHashIndex(left, index)
	default:
		return vm.push(&obj.Error{Message: fmt.Sprintf("index operator not supported: %s", left.Type())})
	}
}

func (vm *VM) executeSingleIndexAssign(kind int) error {
	val := vm.pop()
	index := vm.pop()
	left := vm.pop()
	if left == nil {
		left = Null
	}
	if index == nil {
		index = Null
	}
	if val == nil {
		val = Null
	}
	if errObj, ok := left.(*obj.Error); ok {
		vm.stack[vm.sp] = errObj
		return nil
	}
	if errObj, ok := index.(*obj.Error); ok {
		vm.stack[vm.sp] = errObj
		return nil
	}
	if errObj, ok := val.(*obj.Error); ok {
		vm.stack[vm.sp] = errObj
		return nil
	}

	result := vm.singleIndexAssign(left, index, val, kind)
	vm.stack[vm.sp] = result
	return nil
}

func (vm *VM) singleIndexAssign(left, index, val obj.Object, kind int) obj.Object {
	opStr := ""
	switch kind {
	case 1:
		opStr = "+"
	case 2:
		opStr = "-"
	case 3:
		opStr = "*"
	case 4:
		opStr = "/"
	}
	if hash, ok := left.(*obj.Hash); ok {
		key, ok := index.(obj.Hashable)
		if !ok {
			return &obj.Error{Message: fmt.Sprintf("unusable as hash key: %s", index.Type())}
		}
		hashKey := key.HashKey()
		if kind == 0 {
			if _, exists := hash.Pairs[hashKey]; !exists {
				hash.Order = append(hash.Order, hashKey)
			}
			hash.Pairs[hashKey] = obj.HashPair{Key: index, Value: val}
			return Null
		}
		pair, ok := hash.Pairs[hashKey]
		if !ok {
			return &obj.Error{Message: fmt.Sprintf("key not found: %s", index.Inspect())}
		}
		res := executeBinaryOperation(kindToOpcode(kind), pair.Value, val)
		if _, ok := res.(*obj.Error); ok {
			_ = opStr
			return res
		}
		hash.Pairs[hashKey] = obj.HashPair{Key: index, Value: res}
		return Null
	}
	if mat, ok := left.(*obj.Matrix); ok {
		i, ok := index.(*obj.Integer)
		if !ok {
			return &obj.Error{Message: fmt.Sprintf("index assignment requires integer index, got %s", index.Type())}
		}
		if i.Value < 0 || i.Value >= int64(mat.Rows) {
			return &obj.Error{Message: fmt.Sprintf("index out of range: %d", i.Value)}
		}
		if kind != 0 {
			return &obj.Error{Message: fmt.Sprintf("matrix row assignment only supports =, got %s", kindToOpString(kind))}
		}
		arr, ok := val.(*obj.Array)
		if !ok {
			return &obj.Error{Message: fmt.Sprintf("matrix row assignment requires array, got %s", val.Type())}
		}
		if len(arr.Elements) != mat.Cols {
			return &obj.Error{Message: fmt.Sprintf("matrix row assignment dimension mismatch: got %d, want %d", len(arr.Elements), mat.Cols)}
		}
		for _, e := range arr.Elements {
			if _, ok := e.(*obj.Integer); !ok {
				if _, ok := e.(*obj.Float); !ok {
					return &obj.Error{Message: fmt.Sprintf("matrix row elements must be numeric, got %s", e.Type())}
				}
			}
		}
		newRow := make([]obj.Object, mat.Cols)
		copy(newRow, arr.Elements)
		mat.Data[i.Value] = newRow
		return Null
	}
	array, ok := left.(*obj.Array)
	if !ok {
		return &obj.Error{Message: fmt.Sprintf("index assignment requires array or hash, got %s", left.Type())}
	}
	i, ok := index.(*obj.Integer)
	if !ok {
		return &obj.Error{Message: fmt.Sprintf("index assignment requires integer index, got %s", index.Type())}
	}
	if i.Value < 0 || i.Value >= int64(len(array.Elements)) {
		return &obj.Error{Message: fmt.Sprintf("index out of range: %d", i.Value)}
	}
	if kind == 0 {
		array.Elements[i.Value] = val
		return Null
	}
	res := executeBinaryOperation(kindToOpcode(kind), array.Elements[i.Value], val)
	if _, ok := res.(*obj.Error); ok {
		return res
	}
	array.Elements[i.Value] = res
	return Null
}

func (vm *VM) executeDoubleIndexAssign(kind int) error {
	val := vm.pop()
	colIdxObj := vm.pop()
	rowIdxObj := vm.pop()
	matObj := vm.pop()
	if matObj == nil {
		matObj = Null
	}
	if rowIdxObj == nil {
		rowIdxObj = Null
	}
	if colIdxObj == nil {
		colIdxObj = Null
	}
	if val == nil {
		val = Null
	}
	for _, o := range []obj.Object{matObj, rowIdxObj, colIdxObj, val} {
		if errObj, ok := o.(*obj.Error); ok {
			vm.stack[vm.sp] = errObj
			return nil
		}
	}
	result := vm.doubleIndexAssign(matObj, rowIdxObj, colIdxObj, val, kind)
	vm.stack[vm.sp] = result
	return nil
}

func (vm *VM) doubleIndexAssign(matObj, rowIdxObj, colIdxObj, val obj.Object, kind int) obj.Object {
	rowIdx, ok := rowIdxObj.(*obj.Integer)
	if !ok {
		return &obj.Error{Message: fmt.Sprintf("index operator requires integer index, got %s", rowIdxObj.Type())}
	}
	colIdx, ok := colIdxObj.(*obj.Integer)
	if !ok {
		return &obj.Error{Message: fmt.Sprintf("index operator requires integer index, got %s", colIdxObj.Type())}
	}
	if mat, ok := matObj.(*obj.Matrix); ok {
		if rowIdx.Value < 0 || rowIdx.Value >= int64(mat.Rows) {
			return &obj.Error{Message: fmt.Sprintf("index out of range: %d", rowIdx.Value)}
		}
		if colIdx.Value < 0 || colIdx.Value >= int64(mat.Cols) {
			return &obj.Error{Message: fmt.Sprintf("index out of range: %d", colIdx.Value)}
		}
		if _, ok := val.(*obj.Integer); !ok {
			if _, ok := val.(*obj.Float); !ok {
				if kind == 0 {
					return &obj.Error{Message: fmt.Sprintf("matrix element must be numeric, got %s", val.Type())}
				}
			}
		}
		if kind == 0 {
			mat.Data[rowIdx.Value][colIdx.Value] = val
			return Null
		}
		curr := mat.Data[rowIdx.Value][colIdx.Value]
		res := executeBinaryOperation(kindToOpcode(kind), curr, val)
		if _, ok := res.(*obj.Error); ok {
			return res
		}
		if _, ok := res.(*obj.Integer); !ok {
			if _, ok := res.(*obj.Float); !ok {
				return &obj.Error{Message: fmt.Sprintf("matrix element must be numeric, got %s", res.Type())}
			}
		}
		mat.Data[rowIdx.Value][colIdx.Value] = res
		return Null
	}
	if arr, ok := matObj.(*obj.Array); ok {
		if rowIdx.Value < 0 || rowIdx.Value >= int64(len(arr.Elements)) {
			return &obj.Error{Message: fmt.Sprintf("index out of range: %d", rowIdx.Value)}
		}
		row, ok := arr.Elements[rowIdx.Value].(*obj.Array)
		if !ok {
			return &obj.Error{Message: fmt.Sprintf("index assignment requires array or hash, got %s", arr.Elements[rowIdx.Value].Type())}
		}
		if colIdx.Value < 0 || colIdx.Value >= int64(len(row.Elements)) {
			return &obj.Error{Message: fmt.Sprintf("index out of range: %d", colIdx.Value)}
		}
		if kind == 0 {
			row.Elements[colIdx.Value] = val
			return Null
		}
		res := executeBinaryOperation(kindToOpcode(kind), row.Elements[colIdx.Value], val)
		if _, ok := res.(*obj.Error); ok {
			return res
		}
		row.Elements[colIdx.Value] = res
		return Null
	}
	return &obj.Error{Message: fmt.Sprintf("index assignment requires array or hash, got %s", matObj.Type())}
}

func kindToOpcode(kind int) code.Opcode {
	switch kind {
	case 1:
		return code.OpAdd
	case 2:
		return code.OpSub
	case 3:
		return code.OpMul
	case 4:
		return code.OpDiv
	default:
		return code.OpAdd
	}
}

func kindToOpString(kind int) string {
	switch kind {
	case 1:
		return "+="
	case 2:
		return "-="
	case 3:
		return "*="
	case 4:
		return "/="
	default:
		return "="
	}
}

func (vm *VM) executeTry() error {
	val := vm.pop()
	if val == nil {
		val = Null
	}
	if errObj, ok := val.(*obj.Error); ok {
		arr := &obj.Array{Elements: []obj.Object{False, &obj.String{Value: errObj.Message}}}
		return vm.push(arr)
	}
	arr := &obj.Array{Elements: []obj.Object{True, val}}
	return vm.push(arr)
}

func (vm *VM) currentFrame() *Frame {
	return vm.frames[vm.frameIndex-1]
}

func (vm *VM) pushFrame(f *Frame) {
	vm.frames[vm.frameIndex] = f
	vm.frameIndex++
}

func (vm *VM) popFrame() *Frame {
	vm.frameIndex--
	return vm.frames[vm.frameIndex]
}

func (vm *VM) callFunction(cl *obj.Closure, numArgs int) error {
	if numArgs != cl.Fn.NumParams {
		msg := fmt.Sprintf("wrong number of arguments. got=%d, want=%d", numArgs, cl.Fn.NumParams)
		vm.sp -= numArgs + 1
		if vm.sp < 0 {
			vm.sp = 0
		}
		return vm.push(&obj.Error{Message: msg})
	}

	frame := NewFrame(cl, vm.sp-numArgs)
	vm.pushFrame(frame)

	vm.sp = frame.basePointer + cl.Fn.NumLocals
	for i := numArgs; i < cl.Fn.NumLocals; i++ {
		vm.stack[frame.basePointer+i] = nil
	}

	return nil
}

func (vm *VM) callBuiltin(builtin *obj.Builtin, numArgs int) error {
	args := vm.stack[vm.sp-numArgs : vm.sp]

	result := builtin.Fn(args...)
	vm.sp = vm.sp - numArgs - 1

	if result != nil {
		return vm.push(result)
	}
	return vm.push(Null)
}

func (vm *VM) executeCall(numArgs int) error {
	callee := vm.stack[vm.sp-1-numArgs]
	if callee == nil {
		callee = Null
	}
	if errObj, ok := callee.(*obj.Error); ok {
		vm.sp -= numArgs + 1
		if vm.sp < 0 {
			vm.sp = 0
		}
		return vm.push(errObj)
	}

	switch callee := callee.(type) {
	case *obj.Closure:
		return vm.callFunction(callee, numArgs)
	case *obj.Builtin:
		return vm.callBuiltin(callee, numArgs)
	default:
		msg := fmt.Sprintf("not a function: %s", callee.Type())
		vm.sp -= numArgs + 1
		if vm.sp < 0 {
			vm.sp = 0
		}
		return vm.push(&obj.Error{Message: msg})
	}
}

func (vm *VM) pushClosure(constIndex, numFree int) error {
	constant := vm.constant[constIndex]
	function, ok := constant.(*obj.CompiledFunction)
	if !ok {
		return fmt.Errorf("not a function: %+v", constant)
	}

	free := make([]obj.Object, numFree)
	for i := range numFree {
		entry := vm.stack[vm.sp-numFree+i]
		if cell, ok := entry.(*obj.Cell); ok && cell != nil {
			free[i] = cell
		} else {
			free[i] = entry
		}
	}
	vm.sp -= numFree

	closure := &obj.Closure{Fn: function, Free: free}
	return vm.push(closure)
}

func isTruthy(o obj.Object) bool {
	if o == nil {
		return false
	}
	if errObj, ok := o.(*obj.Error); ok && errObj != nil {
		return true
	}
	switch o := o.(type) {
	case *obj.Boolean:
		return o.Value
	case *obj.Null:
		return false
	default:
		return true
	}
}

func NewWithGlobalsStore(bytecode *compiler.Bytecode, s []obj.Object) *VM {
	vm := New(bytecode)
	vm.globals = s

	return vm
}
