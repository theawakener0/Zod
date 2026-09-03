package code

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

type Instructions []byte

type Opcode byte

const (
	OpConstant Opcode = iota
	OpAdd
	OpPop
	OpSub
	OpMul
	OpDiv
	OpTrue
	OpFalse
	OpEqual
	OpNotEqual
	OpGreaterThan
	OpGreaterThanEqual
	OpMinus
	OpBang
	OpJumpNotTruthy
	OpJump
	OpNull
	OpGetGlobal
	OpSetGlobal
	OpArray
	OpHash
	OpIndex
)

type Definition struct {
	Name 			string
	OperandWidths 	[]int
}

var definition = map[Opcode]*Definition {
	OpConstant: {
		Name: "OpConstant",
		OperandWidths: []int{2},
	},
	OpAdd: {
		Name: "OpAdd",
		OperandWidths: []int{},
	},
	OpPop: {
		Name: "OpPop",
		OperandWidths: []int{},
	},
	OpSub: {
		Name: "OpSub",
		OperandWidths: []int{},
	},
	OpMul: {
		Name: "OpMul",
		OperandWidths: []int{},
	},
	OpDiv: {
		Name: "OpDiv",
		OperandWidths: []int{},
	},
	OpTrue: {
		Name: "OpTrue",
		OperandWidths: []int{},
	},
	OpFalse: {
		Name: "OpFalse",
		OperandWidths: []int{},
	},
	OpEqual: {
		Name: "OpEqual",
		OperandWidths: []int{},
	},
	OpNotEqual: {
		Name: "OpNotEqual",
		OperandWidths: []int{},
	},
	OpGreaterThan: {
		Name: "OpGreaterThan",
		OperandWidths: []int{},
	},
	OpGreaterThanEqual: {
		Name: "OpGreaterThanEqual",
		OperandWidths: []int{},
	},
	OpMinus: {
		Name: "OpMinus",
		OperandWidths: []int{},
	},
	OpBang: {
		Name: "OpBang",
		OperandWidths: []int{},
	},
	OpJumpNotTruthy: {
		Name: "OpJumpNotTruthy",
		OperandWidths: []int{2},
	},
	OpJump: {
		Name: "OpJump",
		OperandWidths: []int{2},
	},
	OpNull: {
		Name: "OpNull",
		OperandWidths: []int{},
	},
	OpGetGlobal: {
		Name: "OpGetGlobal",
		OperandWidths: []int{2},
	},
	OpSetGlobal: {
		Name: "OpSetGlobal",
		OperandWidths: []int{2},
	},
	OpArray: {
		Name: "OpArray",
		OperandWidths: []int{2},
	},
	OpHash: {
		Name: "OpHash",
		OperandWidths: []int{2},
	},
	OpIndex: {
		Name: "OpIndex",
		OperandWidths: []int{},
	},
}

func Lookup(op byte) (*Definition, error) {
	def, ok := definition[Opcode(op)]
	if !ok {
		return nil, fmt.Errorf("opcode %d undifined", op)
	}

	return def, nil
}

func Make(op Opcode, operands ...int) []byte {
	def, ok := definition[op]
	if !ok {
		return []byte{}
	}

	instructionLen := 1
	for _, w := range def.OperandWidths {
		instructionLen += w
	}

	instruction := make([]byte, instructionLen)
	instruction[0] = byte(op)

	offset := 1
	for i, w := range operands {
		width := def.OperandWidths[i]
		switch width {
		case 2:
			binary.BigEndian.PutUint16(instruction[offset:], uint16(w))
		}
		offset += width
	}

	return instruction
}

func (ins Instructions) String() string {
	var out bytes.Buffer

	i := 0
	for i < len(ins) {
		def, err := Lookup(ins[i])
		if err != nil {
			fmt.Fprintf(&out, "ERROR: %s\n", err)
			continue
		}

		operands, read := ReadOperands(def, ins[i+1:])
		
		fmt.Fprintf(&out, "%04d %s\n", i, ins.fmtInstructions(def, operands))

		i += 1 + read
	}

	return out.String()
}

func ReadOperands(def *Definition, ins Instructions) ([]int, int) {
	operands := make([]int, len(def.OperandWidths))
	offset := 0
	
	for i, width := range def.OperandWidths {
		switch width {
		case 2:
			operands[i] = int(binary.BigEndian.Uint16(ins[offset:]))
		}

		offset += width
	}

	return  operands, offset
}

func (ins Instructions) fmtInstructions(def *Definition, operands []int) string {
	operandCount := len(def.OperandWidths)

	if len(operands) != operandCount {
		return fmt.Sprintf("ERROR: operand len %d does not match defined %d\n", len(operands), operandCount)
	}

	switch operandCount {
	case 0: 
		return def.Name
	case 1:
		return fmt.Sprintf("%s %d", def.Name, operands[0])
	}

	return fmt.Sprintf("ERROR: unhandled operandCount for %s\n", def.Name)
}


