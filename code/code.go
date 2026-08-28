package code

import (
	"encoding/binary"
	"fmt"
)

type Instructions []byte

type Opcode byte

const (
	OpConstant Opcode = iota
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
