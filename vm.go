package main

import (
	"fmt"
	"strings"
)

const (
	// 256
	STACK_MAX   = 10
	MAX_GLOBALS = 1 << 16
)

type VM struct {
	Instructions Instructions
	Constants    []Object
	Stack        []Object
	Ip           int
	Sp           int
	Globals      []Object
}

func NewVM(bytecode *ByteCode) *VM {
	vm := &VM{Ip: 0, Sp: 0}
	vm.Instructions = bytecode.Code
	vm.Constants = bytecode.Constants
	vm.Stack = make([]Object, STACK_MAX)
	vm.Globals = make([]Object, MAX_GLOBALS)
	return vm
}

func (vm *VM) run() error {
	instrLen := len(vm.Instructions)

	for vm.Ip != instrLen {
		instruction := OpCode(vm.Instructions[vm.Ip])
		// vm.Chunk.disassembleInstruction(vm.Ip)

		definition := definitions[instruction]
		fmt.Printf("Instruction: %s, Sp: %d, Ip: %d, Stack before: %v\n", definition.Name, vm.Sp, vm.Ip, vm.printStack())
		vm.Ip += 1

		switch instruction {
		case OP_POP:
			vm.pop()
		case OP_RETURN:
			// result := vm.pop()
			// fmt.Println(result.Inspect())
			return nil
		case OP_GET_BUILTIN:
			builinIndex := ReadUint8(vm.Instructions[vm.Ip:])

			definition := Builtins[builinIndex]

			err := vm.push(definition.Builtin)
			if err != nil {
				return nil
			}
		case OP_GET_GLOBAL:
			index := ReadUint16(vm.Instructions[vm.Ip:])
			vm.Ip += 2
			value := vm.Globals[index]
			err := vm.push(value)
			if err != nil {
				return err
			}
		case OP_GET_LOCAL:
			index := ReadUint8(vm.Instructions[vm.Ip:])
			vm.Ip += 1
			err := vm.push(vm.Stack[index])
			if err != nil {
				return err
			}
		case OP_SET_LOCAL:
			index := ReadUint8(vm.Instructions[vm.Ip:])
			vm.Ip += 1
			value := vm.pop()
			vm.Stack[index] = value
		case OP_SET_GLOBAL:
			index := ReadUint16(vm.Instructions[vm.Ip:])
			vm.Ip += 2
			vm.Globals[index] = vm.pop()
		case OP_DEFINE_GLOBAL:
			index := ReadUint16(vm.Instructions[vm.Ip:])
			vm.Ip += 2
			vm.Globals[index] = vm.pop()
		case OP_DEFINE_LOCAL:
			index := ReadUint8(vm.Instructions[vm.Ip:])
			vm.Ip += 1
			vm.Stack[index] = vm.peek(0)
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
		case OP_JUMP_IF_FALSE:
			offset := ReadUint16(vm.Instructions[vm.Ip:])
			vm.Ip += 2
			if !vm.isTruthy(vm.peek(0)) {
				fmt.Printf("Jumping by offset %d because top of stack is falsey\n", offset)
				vm.Ip += int(offset)
				vm.pop()
			}
		case OP_JUMP:
			offset := ReadUint16(vm.Instructions[vm.Ip:])
			fmt.Printf("Unconditional jump by offset %d\n", offset)
			vm.Ip += int(offset)
		case OP_LOOP:
			offset := ReadUint16(vm.Instructions[vm.Ip:])
			fmt.Printf("Looping back by offset %d\n", offset)
			vm.Ip += 1
			vm.Ip -= int(offset)
		}

		fmt.Printf("Stack after: %v\n", vm.printStack())
	}

	return nil
}

func (vm *VM) readString(obj Object) (string, error) {
	if obj.Type() != StringObj {
		return "", fmt.Errorf("cannot define a variable whose name is not a string")
	}
	str := obj.(*StringObject)
	return str.Value, nil
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
	case StringObj:
		leftValue := left.(*StringObject)
		rightValue := right.(*StringObject)
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
	case leftType == StringObj && rightType == StringObj:
		return vm.executeStringOperation(left, op, right)
	default:
		return fmt.Errorf("unsupported types for binary operation: %s %s", leftType, rightType)
	}
}

func (vm *VM) executeStringOperation(left Object, op string, right Object) error {
	leftValue := left.(*StringObject)
	rightValue := right.(*StringObject)

	var result string

	switch op {
	case "+":
		result = leftValue.Value + rightValue.Value
	default:
		return fmt.Errorf("unknown operator: %s", op)

	}

	return vm.push(&StringObject{Value: result})
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

func (vm *VM) peek(index int) Object {
	return vm.Stack[vm.Sp-1-index]
}

func (vm *VM) pop() Object {
	vm.Sp -= 1
	return vm.Stack[vm.Sp]
}

func (vm *VM) printStack() string {
	var str strings.Builder

	str.WriteString("[")

	for i := 0; i < vm.Sp; i++ {
		obj := vm.Stack[i]
		if obj != nil {
			str.WriteString(obj.Inspect() + ", ")
			continue
		}
		break
	}

	str.WriteString("]")
	return str.String()
}

func (vm *VM) isTruthy(value Object) bool {
	switch v := value.(type) {
	case *StringObject:
		return len(v.Value) > 0
	case *FloatObject:
		return v.Value > 0
	case *BooleanObject:
		return v.Value
	case *NilObject:
		return false
	default:
		return true
	}
}

func (vm *VM) readConstant(index int) Object {
	value := vm.Constants[index]
	return value
}
