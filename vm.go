package main

import "fmt"

const (
	STACK_MAX = 256
)

type VM struct {
	Chunk *Chunk
	Stack []Value
	Ip    int
	Sp    int
}

func newVM(chunk *Chunk) *VM {
	vm := &VM{Ip: 0, Sp: 0}
	vm.Chunk = chunk
	vm.Stack = make([]Value, STACK_MAX)
	return vm
}

func (vm *VM) run() error {
	instrLen := len(vm.Chunk.Code)

	for vm.Ip != instrLen {
		instruction := vm.Chunk.Code[vm.Ip]
		// vm.Chunk.disassembleInstruction(vm.Ip)
		vm.Ip += 1
		switch instruction {
		case byte(OP_RETURN):
			fmt.Printf("%f\n", vm.pop())
		case byte(OP_CONSTANT):
			index := vm.Chunk.ReadUint16(vm.Chunk.Code[vm.Ip:])
			constant := vm.readConstant(int(index))
			vm.push(constant)
			vm.Ip += 2
		case byte(OP_NEGATE):
			vm.push(-vm.pop())
		case byte(OP_ADD):
			vm.executeBinary("+")
		case byte(OP_SUBTRACT):
			vm.executeBinary("-")
		case byte(OP_MULTIPLY):
			vm.executeBinary("*")
		case byte(OP_DIVIDE):
			vm.executeBinary("/")
		}
	}

	return nil
}

func (vm *VM) executeBinary(op string) {
	right := vm.pop()
	left := vm.pop()

	switch op {
	case "+":
		vm.push(left + right)
	case "-":
		vm.push(left - right)
	case "*":
		vm.push(left * right)
	case "/":
		vm.push(left / right)

	}
}

func (vm *VM) push(val Value) {
	vm.Stack[vm.Sp] = val
	vm.Sp += 1
}

func (vm *VM) pop() Value {
	vm.Sp -= 1
	return vm.Stack[vm.Sp]
}

func (vm *VM) readConstant(index int) Value {
	value := vm.Chunk.Constants[index]
	return value
}
