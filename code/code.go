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
	OpCall
	OpReturnValue
	OpReturn
	OpGetLocal
	OpSetLocal
	OpGetBuiltin
	OpClosure
	OpGetFree
	OpCurrentClosure
	OpDup
	OpSetFree
	OpSetIndex
	OpTry
	OpGetLocalCell
	OpGetFreeCell
	OpDefineLocal
	OpImport
	OpGetProp
	OpMod
	OpMatrixCell
)

type Definition struct {
	Name          string
	OperandWidths []int
}

var definition = map[Opcode]*Definition{
	OpConstant: {
		Name:          "OpConstant",
		OperandWidths: []int{2},
	},
	OpAdd: {
		Name:          "OpAdd",
		OperandWidths: []int{},
	},
	OpPop: {
		Name:          "OpPop",
		OperandWidths: []int{},
	},
	OpSub: {
		Name:          "OpSub",
		OperandWidths: []int{},
	},
	OpMul: {
		Name:          "OpMul",
		OperandWidths: []int{},
	},
	OpDiv: {
		Name:          "OpDiv",
		OperandWidths: []int{},
	},
	OpTrue: {
		Name:          "OpTrue",
		OperandWidths: []int{},
	},
	OpFalse: {
		Name:          "OpFalse",
		OperandWidths: []int{},
	},
	OpEqual: {
		Name:          "OpEqual",
		OperandWidths: []int{},
	},
	OpNotEqual: {
		Name:          "OpNotEqual",
		OperandWidths: []int{},
	},
	OpGreaterThan: {
		Name:          "OpGreaterThan",
		OperandWidths: []int{},
	},
	OpGreaterThanEqual: {
		Name:          "OpGreaterThanEqual",
		OperandWidths: []int{},
	},
	OpMinus: {
		Name:          "OpMinus",
		OperandWidths: []int{},
	},
	OpBang: {
		Name:          "OpBang",
		OperandWidths: []int{},
	},
	OpJumpNotTruthy: {
		Name:          "OpJumpNotTruthy",
		OperandWidths: []int{2},
	},
	OpJump: {
		Name:          "OpJump",
		OperandWidths: []int{2},
	},
	OpNull: {
		Name:          "OpNull",
		OperandWidths: []int{},
	},
	OpGetGlobal: {
		Name:          "OpGetGlobal",
		OperandWidths: []int{2},
	},
	OpSetGlobal: {
		Name:          "OpSetGlobal",
		OperandWidths: []int{2},
	},
	OpArray: {
		Name:          "OpArray",
		OperandWidths: []int{2},
	},
	OpHash: {
		Name:          "OpHash",
		OperandWidths: []int{2},
	},
	OpIndex: {
		Name:          "OpIndex",
		OperandWidths: []int{},
	},
	OpCall: {
		Name:          "OpCall",
		OperandWidths: []int{1},
	},
	OpReturnValue: {
		Name:          "OpReturnValue",
		OperandWidths: []int{},
	},
	OpReturn: {
		Name:          "OpReturn",
		OperandWidths: []int{},
	},
	OpGetLocal: {
		Name:          "OpGetLocal",
		OperandWidths: []int{1},
	},
	OpSetLocal: {
		Name:          "OpSetLocal",
		OperandWidths: []int{1},
	},
	OpGetBuiltin: {
		Name:          "OpGetBuiltin",
		OperandWidths: []int{1},
	},
	OpClosure: {
		Name:          "OpClosure",
		OperandWidths: []int{2, 1},
	},
	OpGetFree: {
		Name:          "OpGetFree",
		OperandWidths: []int{1},
	},
	OpCurrentClosure: {
		Name:          "OpCurrentClosure",
		OperandWidths: []int{},
	},
	OpDup: {
		Name:          "OpDup",
		OperandWidths: []int{},
	},
	OpSetFree: {
		Name:          "OpSetFree",
		OperandWidths: []int{1},
	},
	OpSetIndex: {
		Name:          "OpSetIndex",
		OperandWidths: []int{1},
	},
	OpTry: {
		Name:          "OpTry",
		OperandWidths: []int{},
	},
	OpGetLocalCell: {
		Name:          "OpGetLocalCell",
		OperandWidths: []int{1},
	},
	OpGetFreeCell: {
		Name:          "OpGetFreeCell",
		OperandWidths: []int{1},
	},
	OpDefineLocal: {
		Name:          "OpDefineLocal",
		OperandWidths: []int{1},
	},
	OpImport: {
		Name:          "OpImport",
		OperandWidths: []int{2},
	},
	OpGetProp: {
		Name:          "OpGetProp",
		OperandWidths: []int{2},
	},
	OpMod: {
		Name:          "OpMod",
		OperandWidths: []int{},
	},
	OpMatrixCell: {
		Name:          "OpMatrixCell",
		OperandWidths: []int{},
	},
}

var definitionCache [256]*Definition

func init() {
	for op, def := range definition {
		definitionCache[op] = def
	}
}

func Lookup(op byte) (*Definition, error) {
	if def := definitionCache[op]; def != nil {
		return def, nil
	}
	return nil, fmt.Errorf("opcode %d undifined", op)
}

func Make(op Opcode, operands ...int) []byte {
	def, ok := definition[op]
	if !ok {
		return []byte{}
	}

	widths := def.OperandWidths
	instructionLen := 1
	for _, w := range widths {
		instructionLen += w
	}

	instruction := make([]byte, instructionLen)
	instruction[0] = byte(op)

	offset := 1
	for i, operand := range operands {
		width := widths[i]
		if width == 2 {
			instruction[offset] = byte(operand >> 8)
			instruction[offset+1] = byte(operand)
		} else {
			instruction[offset] = byte(operand)
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
			i++
			continue
		}

		operands, read := ReadOperands(def, ins[i+1:])

		fmt.Fprintf(&out, "%04d %s\n", i, ins.fmtInstructions(def, operands))

		i += 1 + read
	}

	return out.String()
}

func ReadOperands(def *Definition, ins Instructions) ([]int, int) {
	if len(def.OperandWidths) == 0 {
		return nil, 0
	}
	operands := make([]int, len(def.OperandWidths))
	offset := 0

	for i, width := range def.OperandWidths {
		switch width {
		case 2:
			operands[i] = int(binary.BigEndian.Uint16(ins[offset:]))
		case 1:
			operands[i] = int(ReadUnit8(ins[offset:]))
		}

		offset += width
	}

	return operands, offset
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
	case 2:
		return fmt.Sprintf("%s %d %d", def.Name, operands[0], operands[1])
	}

	return fmt.Sprintf("ERROR: unhandled operandCount for %s\n", def.Name)
}

func ReadUnit8(ins Instructions) uint8 { return uint8(ins[0]) }
