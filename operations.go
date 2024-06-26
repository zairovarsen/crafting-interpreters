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
	OP_POP
	OP_DEFINE_GLOBAL
	OP_DEFINE_LOCAL
	OP_SET_GLOBAL
	OP_SET_LOCAL
	OP_GET_BUILTIN
	OP_GET_GLOBAL
	OP_GET_UPVALUE
	OP_GET_LOCAL
	OP_JUMP_IF_FALSE
	OP_JUMP
	OP_LOOP
	OP_CALL
	OP_FUNCTION
	OP_CLOSURE
	OP_CLASS
	OP_SET_PROPERTY
	OP_GET_PROPERTY
	OP_METHOD
)

type Definition struct {
	Name          string
	OperandWidths []int
}

var definitions = map[OpCode]*Definition{
	OP_CONSTANT:      {"OP_CONSTANT", []int{2}},
	OP_NEGATE:        {"OP_NEGATE", []int{}},
	OP_RETURN:        {"OP_RETURN", []int{}},
	OP_ADD:           {"OP_ADD", []int{}},
	OP_SUBTRACT:      {"OP_SUBTRACT", []int{}},
	OP_MULTIPLY:      {"OP_MULTIPLY", []int{}},
	OP_DIVIDE:        {"OP_DIVIDE", []int{}},
	OP_TRUE:          {"OP_TRUE", []int{}},
	OP_FALSE:         {"OP_FALSE", []int{}},
	OP_NIL:           {"OP_NIL", []int{}},
	OP_LESS:          {"OP_LESS", []int{}},
	OP_GREATER:       {"OP_GREATER", []int{}},
	OP_EQUAL:         {"OP_EQUAL", []int{}},
	OP_NOT:           {"OP_NOT", []int{}},
	OP_POP:           {"OP_POP", []int{}},
	OP_DEFINE_GLOBAL: {"OP_DEFINE_GLOBAL", []int{2}},
	OP_DEFINE_LOCAL:  {"OP_DEFINE_LOCAL", []int{1}},
	OP_GET_GLOBAL:    {"OP_GET_GLOBAL", []int{2}},
	OP_GET_LOCAL:     {"OP_GET_LOCAL", []int{1}},
	OP_GET_BUILTIN:   {"OP_GET_BUILTIN", []int{1}},
	OP_SET_GLOBAL:    {"OP_SET_GLOBAL", []int{2}},
	OP_SET_LOCAL:     {"OP_SET_LOCAL", []int{1}},
	OP_JUMP_IF_FALSE: {"OP_JUMP_IF_FALSE", []int{2}},
	OP_JUMP:          {"OP_JUMP", []int{2}},
	OP_LOOP:          {"OP_LOOP", []int{2}},
	OP_CALL:          {"OP_CALL", []int{1}},
	OP_FUNCTION:      {"OP_FUNCTION", []int{2}},
	OP_CLOSURE:       {"OP_CLOSURE", []int{2, 1}},
	OP_GET_UPVALUE:   {"OP_GET_UPVALUE", []int{1}},
	OP_CLASS:         {"OP_CLASS", []int{2}},
	OP_SET_PROPERTY:  {"OP_SET_PROPERTY", []int{1}},
	OP_GET_PROPERTY:  {"OP_GET_PROPERTY", []int{1}},
	OP_METHOD:        {"OP_METHOD", []int{2}},
}

func Lookup(opcode byte) (*Definition, error) {
	if def, ok := definitions[OpCode(opcode)]; ok {
		return def, nil
	}

	return nil, fmt.Errorf("Invalid opcode: %d\n", opcode)
}
