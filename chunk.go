package main

import (
	"encoding/binary"
	"fmt"
)

type OpCode byte
type Instructions []byte
type Value float64

const (
	OP_RETURN OpCode = iota
	OP_CONSTANT
	OP_NEGATE
	OP_ADD
	OP_SUBTRACT
	OP_MULTIPLY
	OP_DIVIDE
)

type Definition struct {
	Name          string
	OperandWidths []int
}

var definitions = map[OpCode]*Definition{
	OP_CONSTANT: {"OP_CONSTANT", []int{2}},
	OP_NEGATE:   {"OP_NEGATE", []int{}},
	OP_RETURN:   {"OP_RETURN", []int{}},
	OP_ADD:      {"OP_ADD", []int{}},
	OP_SUBTRACT: {"OP_SUBTRACT", []int{}},
	OP_MULTIPLY: {"OP_MULTIPLY", []int{}},
	OP_DIVIDE:   {"OP_DIVIDE", []int{}},
}

type LineInfo struct {
	Line  int
	Count int
}

type Chunk struct {
	Code      Instructions
	Constants []Value
	LineInfo  []LineInfo
}

func newChunk() *Chunk {
	return &Chunk{Code: make([]byte, 0), Constants: make([]Value, 0), LineInfo: make([]LineInfo, 0)}
}

func (c *Chunk) Lookup(opcode byte) (*Definition, error) {
	if def, ok := definitions[OpCode(opcode)]; ok {
		return def, nil
	}

	return nil, fmt.Errorf("Invalid opcode: %d\n", opcode)
}

func (c *Chunk) ReadUint16(ins Instructions) uint16 {
	return binary.BigEndian.Uint16(ins)
}

func (c *Chunk) ReadUint8(ins Instructions) uint8 {
	return uint8(ins[0])
}

func (c *Chunk) WriteChunk(opcode OpCode, line int, operands ...int) {
	if len(c.LineInfo) > 0 && c.LineInfo[len(c.LineInfo)-1].Line == line {
		c.LineInfo[len(c.LineInfo)-1].Count++
	} else {
		c.LineInfo = append(c.LineInfo, LineInfo{Line: line, Count: 1})
	}
	c.Code = append(c.Code, byte(opcode))
	definition := definitions[opcode]

	for i, o := range operands {
		operandWidth := definition.OperandWidths[i]
		switch operandWidth {
		case 2:
			c.Code = binary.BigEndian.AppendUint16(c.Code, uint16(o))
		case 1:
			c.Code = append(c.Code, byte(o))
		}
	}
}

func (c *Chunk) GetLine(insIndex int) int {
	accumulatedCount := 0

	for _, info := range c.LineInfo {
		accumulatedCount += info.Count
		if insIndex < accumulatedCount {
			return info.Line
		}
	}

	return -1 // In case of an invalid instruction index
}

func (c *Chunk) WriteValue(value Value) uint16 {
	// max index is 1 << 16
	c.Constants = append(c.Constants, value)
	return uint16((len(c.Constants) - 1))
}

func (c *Chunk) DisassembleChunks() {
	var offset int

	for offset < len(c.Code) {
		offset = c.disassembleInstruction(offset)
	}
}

func (c *Chunk) disassembleInstruction(offset int) int {
	definition, err := c.Lookup(c.Code[offset])
	newOffset := offset + 1

	if err != nil {
		fmt.Println(err)
		return newOffset
	}
	fmt.Printf("%04d %4d %s", offset, c.GetLine(offset), definition.Name)

	for _, w := range definition.OperandWidths {
		switch w {
		case 2:
			fmt.Printf(" %f", c.Constants[c.ReadUint16(c.Code[newOffset:])])
		case 1:
			fmt.Printf(" %f", c.Constants[c.ReadUint8(c.Code[newOffset:])])
		}
		newOffset += w
	}
	fmt.Printf("\n")

	return newOffset
}
