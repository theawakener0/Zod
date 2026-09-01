package vm

import (
	"encoding/binary"
	"fmt"

	"github.com/theawakener0/Zod/code"
	"github.com/theawakener0/Zod/compiler"
	obj "github.com/theawakener0/Zod/object"
)

const StackSize = 2048

var True = &obj.Boolean{Value: true}
var False = &obj.Boolean{Value: false}

type VM struct {
	constant		[]obj.Object
	instructions	code.Instructions

	stack			[]obj.Object
	sp 				int
}

func New(bytecode *compiler.Bytecode) *VM {
	return &VM{
		instructions: bytecode.Instructions,
		constant: bytecode.Constant,

		stack: make([]obj.Object, StackSize),
		sp: 0,
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
	for ip := 0; ip < len(vm.instructions); ip++ {
		op := code.Opcode(vm.instructions[ip])

		switch op {
		case code.OpConstant:
			constIndex := binary.BigEndian.Uint16(vm.instructions[ip+1:])
			ip += 2

			err := vm.push(vm.constant[constIndex])
			if err != nil {
				return err
			}
		case code.OpAdd, code.OpSub, code.OpMul, code.OpDiv:
			right := vm.pop()
			left := vm.pop()

			result, err0 := executeBinaryOperation(op, left, right)
			if err0 != nil {
				return err0
			}

			err1 := vm.push(result)
			if err1 != nil {
				return err1
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

func executeBinaryOperation(op code.Opcode, left, right obj.Object) (obj.Object, error) {
	if left.Type() == obj.INTEGER_OBJ && right.Type() == obj.INTEGER_OBJ {

		leftVal := left.(*obj.Integer).Value
		rightVal := right.(*obj.Integer).Value

		switch op {
		case code.OpAdd:
			return &obj.Integer{Value: leftVal + rightVal}, nil
		case code.OpSub:
			return &obj.Integer{Value: leftVal - rightVal}, nil
		case code.OpMul:
			return &obj.Integer{Value: leftVal * rightVal}, nil
		case code.OpDiv:
			return &obj.Integer{Value: leftVal / rightVal}, nil
		}
	}

	leftVal, leftOk := numericValue(left)
	rightVal, rightOk := numericValue(right)
	if !leftOk || !rightOk {
		return nil, fmt.Errorf("unsupported type for %s: %s and %s", string(op), left.Type(), right.Type())
	}

	switch op {
	case code.OpAdd:
		return &obj.Float{Value: leftVal + rightVal}, nil
	case code.OpSub:
		return &obj.Float{Value: leftVal - rightVal}, nil
	case code.OpMul:
		return &obj.Float{Value: leftVal * rightVal}, nil
	case code.OpDiv:
		return &obj.Float{Value: leftVal / rightVal}, nil
	}

	return nil, fmt.Errorf("unknown operation: %s", string(op))
}

func nativeBoolToBooleanObj(input bool) *obj.Boolean {
	if input {
		return True
	}
	return False
}

func (vm *VM) executeBinaryComparison(op code.Opcode, left, right obj.Object) error {
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
		}
	}

	leftVal, leftOk := numericValue(left)
	rightVal, rightOk := numericValue(right)
	if !leftOk || !rightOk {
		return fmt.Errorf("unsupported type for %s: %s and %s", string(op), left.Type(), right.Type())
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

	return fmt.Errorf("unknown operation: %s", string(op))
}

func (vm *VM) executeComparison(op code.Opcode) error {
	right := vm.pop()
	left := vm.pop()

	if (left.Type() == obj.INTEGER_OBJ || left.Type() == obj.FLOAT_OBJ) && (right.Type() == obj.INTEGER_OBJ || right.Type() == obj.FLOAT_OBJ) {
		return vm.executeBinaryComparison(op, left, right)
	}

	switch op {
	case code.OpEqual:
		return vm.push(nativeBoolToBooleanObj(right == left))
	case code.OpNotEqual:
		return vm.push(nativeBoolToBooleanObj(right != left))
	default:
		return fmt.Errorf("unknown operator: %d (%s %s)", op, left.Type(), right.Type())
	}
}


func (vm *VM) executeBangOperator() error {
	operand := vm.pop()

	switch operand {
	case True:
		return vm.push(False)
	case False:
		return vm.push(True)
	default:
		return vm.push(False)
	}
}

func (vm *VM) executeMinusOperator() error {
	operand := vm.pop()

	if operand.Type() != obj.INTEGER_OBJ {
		return fmt.Errorf("unsupported type for negation: %s", operand.Type())
	}
	if operand.Type() == obj.FLOAT_OBJ {
		return fmt.Errorf("unsupported type for negation: %s", operand.Type())
	}

	switch operand := operand.(type) {
	case *obj.Integer:
		return vm.push(&obj.Integer{Value: -operand.Value})
	case *obj.Float:
		return vm.push(&obj.Float{Value: -operand.Value})
	}
	
	return fmt.Errorf("unsupported type for negation: %s", operand.Type())
}


