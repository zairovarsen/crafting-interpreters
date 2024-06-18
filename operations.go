package main

import "fmt"

type OpCode byte
type Instructions []byte

const (
	OP_RETURN OpCode = iota
	OP_CONSTANT
	OP_NEGATE
	OP_ADD
	OP_SUBTRACT
	OP_MULTIPLY
	OP_DIVIDE
	OP_NIL
	OP_TRUE
	OP_FALSE
	OP_EQUAL
	OP_GREATER
	OP_LESS
	OP_NOT
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
	OP_TRUE:     {"OP_TRUE", []int{}},
	OP_FALSE:    {"OP_FALSE", []int{}},
	OP_NIL:      {"OP_NIL", []int{}},
	OP_LESS:     {"OP_LESS", []int{}},
	OP_GREATER:  {"OP_GREATER", []int{}},
	OP_EQUAL:    {"OP_EQUAL", []int{}},
	OP_NOT:      {"OP_NOT", []int{}},
}

func Lookup(opcode byte) (*Definition, error) {
	if def, ok := definitions[OpCode(opcode)]; ok {
		return def, nil
	}

	return nil, fmt.Errorf("Invalid opcode: %d\n", opcode)
}