package main

import "fmt"

const (
	STACK_MAX = 256
)

type VM struct {
	Instructions Instructions
	Constants    []Object
	Stack        []Object
	Ip           int
	Sp           int
}

func NewVM(bytecode *ByteCode) *VM {
	vm := &VM{Ip: 0, Sp: 0}
	vm.Instructions = bytecode.Code
	vm.Constants = bytecode.Constants
	vm.Stack = make([]Object, STACK_MAX)
	return vm
}

func (vm *VM) run() error {
	instrLen := len(vm.Instructions)

	for vm.Ip != instrLen {
		instruction := OpCode(vm.Instructions[vm.Ip])
		// vm.Chunk.disassembleInstruction(vm.Ip)
		vm.Ip += 1
		switch instruction {
		case OP_RETURN:
			result := vm.pop()
			fmt.Println(result.Inspect())
		case OP_CONSTANT:
			index := ReadUint16(vm.Instructions[vm.Ip:])
			vm.Ip += 2
			constant := vm.readConstant(int(index))

			err := vm.push(constant)
			if err != nil {
				return err
			}
		case OP_NEGATE:
			err := vm.negate()
			if err != nil {
				return err
			}
		case OP_ADD:
			err := vm.executeBinary("+")
			if err != nil {
				return err
			}
		case OP_SUBTRACT:
			err := vm.executeBinary("-")
			if err != nil {
				return err
			}
		case OP_MULTIPLY:
			err := vm.executeBinary("*")
			if err != nil {
				return err
			}
		case OP_DIVIDE:
			err := vm.executeBinary("/")
			if err != nil {
				return err
			}
		case OP_GREATER:
			err := vm.executeBinary(">")
			if err != nil {
				return err
			}
		case OP_LESS:
			err := vm.executeBinary("<")
			if err != nil {
				return err
			}
		case OP_NIL:
			err := vm.push(Null)
			if err != nil {
				return err
			}
		case OP_TRUE:
			err := vm.push(True)
			if err != nil {
				return err
			}
		case OP_FALSE:
			err := vm.push(False)
			if err != nil {
				return err
			}
		case OP_EQUAL:
			err := vm.equal()
			if err != nil {
				return err
			}
		case OP_NOT:
			err := vm.not()
			if err != nil {
				return err
			}
		}

	}

	return nil
}

func (vm *VM) equal() error {
	left := vm.pop()
	right := vm.pop()

	if left.Type() != right.Type() {
		return fmt.Errorf("cannot compare two values of different types: %s %s", left.Type(), right.Type())
	}

	var result bool

	switch left.Type() {
	case BooleanObj:
		leftValue := left.(*BooleanObject)
		rightValue := right.(*BooleanObject)
		result = leftValue.Value == rightValue.Value
	case NillObj:
		result = true
	case FloatObj:
		leftValue := left.(*FloatObject)
		rightValue := right.(*FloatObject)
		result = leftValue.Value == rightValue.Value
	default:
		result = false
	}

	return vm.push(&BooleanObject{Value: result})
}

func (vm *VM) not() error {
	operand := vm.pop()

	switch operand {
	case True:
		return vm.push(False)
	case False:
		return vm.push(True)
	case Null:
		return vm.push(True)
	default:
		return vm.push(False)
	}
}

func (vm *VM) negate() error {
	operand := vm.pop()

	if operand.Type() != FloatObj {
		return fmt.Errorf("invalid operand for unary operation: %s", operand.Type())
	}

	operandValue := operand.(*FloatObject)

	return vm.push(&FloatObject{Value: -operandValue.Value})
}

func (vm *VM) executeBinary(op string) error {
	right := vm.pop()
	left := vm.pop()

	leftType := left.Type()
	rightType := right.Type()

	switch {
	case leftType == FloatObj && rightType == FloatObj:
		return vm.executeBinaryOperation(left, op, right)
	default:
		return fmt.Errorf("unsupported types for binary operation: %s %s", leftType, rightType)
	}
}

func (vm *VM) executeBinaryOperation(left Object, op string, right Object) error {
	if op == ">" || op == "<" {
		return vm.executeBinaryComparisonOperations(left, op, right)
	}
	return vm.executeBinaryArithmeticOperations(left, op, right)
}

func (vm *VM) executeBinaryComparisonOperations(left Object, op string, right Object) error {
	leftValue := left.(*FloatObject)
	rightValue := right.(*FloatObject)

	var result bool

	switch op {
	case ">":
		result = leftValue.Value > rightValue.Value
	case "<":
		result = leftValue.Value < rightValue.Value
	default:
		return fmt.Errorf("unknown operator: %s", op)
	}

	if result {
		return vm.push(True)
	}
	return vm.push(False)
}

func (vm *VM) executeBinaryArithmeticOperations(left Object, op string, right Object) error {
	leftValue := left.(*FloatObject)
	rightValue := right.(*FloatObject)

	var result float64

	switch op {
	case "+":
		result = leftValue.Value + rightValue.Value
	case "-":
		result = leftValue.Value - rightValue.Value
	case "*":
		result = leftValue.Value * rightValue.Value
	case "/":
		result = leftValue.Value / rightValue.Value
	default:
		return fmt.Errorf("unknown operator: %s", op)
	}

	return vm.push(&FloatObject{Value: result})
}

func (vm *VM) push(value Object) error {
	if vm.Sp >= STACK_MAX {
		return fmt.Errorf("stack overflow")
	}
	vm.Stack[vm.Sp] = value
	vm.Sp += 1
	return nil
}

func (vm *VM) pop() Object {
	vm.Sp -= 1
	return vm.Stack[vm.Sp]
}

func (vm *VM) readConstant(index int) Object {
	value := vm.Constants[index]
	return value
}
