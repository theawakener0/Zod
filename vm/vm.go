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
			leftVal := left.(*obj.Integer).Value
			rightVal := right.(*obj.Integer).Value

			result := leftVal + rightVal
			err := vm.push(&obj.Integer{Value:  result})
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


