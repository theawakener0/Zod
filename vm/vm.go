package vm

import (
	"encoding/binary"
	"fmt"

	"github.com/theawakener0/Zod/code"
	"github.com/theawakener0/Zod/compiler"
	obj "github.com/theawakener0/Zod/object"
)

const StackSize = 2048

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

func (vm *VM) Run() error {
	for ip := 0; ip < len(vm.instructions); ip++ {
		op := code.Opcode(vm.instructions[ip])

		switch op {
		case code.OpConstant:
			cosntIndex := binary.BigEndian.Uint16(vm.instructions[ip+1:])
			ip += 2

			err := vm.push(vm.constant[cosntIndex])
			if err != nil {
				return err
			}
		case code.OpAdd:
			right := vm.pop()
			left := vm.pop()

			result, err0 := executeBinaryOperation("+", left, right)
			if err0 != nil {
				return err0
			}

			err1 := vm.push(result)
			if err1 != nil {
				return err1
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

func executeBinaryOperation(op string, left, right obj.Object) (obj.Object, error) {
	if left.Type() == obj.INTEGER_OBJ && right.Type() == obj.INTEGER_OBJ {

		leftVal := left.(*obj.Integer).Value
		rightVal := right.(*obj.Integer).Value

		switch op {
		case "+":
			return &obj.Integer{Value: leftVal + rightVal}, nil
		}
	}

	leftVal, leftOk := numericValue(left)
	rightVal, rightOk := numericValue(right)
	if !leftOk || !rightOk {
		return nil, fmt.Errorf("unsupported type for %s: %s and %s", op, left.Type(), right.Type())
	}

	switch op {
	case "+":
		return &obj.Float{Value: leftVal + rightVal}, nil
	}

	return nil, fmt.Errorf("unknown operation: %s", op)
}


