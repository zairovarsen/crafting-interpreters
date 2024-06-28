package main

import (
	"fmt"
	"strings"
)

const (
	// 256
	STACK_MAX   = 10
	MAX_GLOBALS = 1 << 16
	FRAMES_MAX  = 64
)

type VM struct {
	Constants  []Object
	Stack      []Object
	Sp         int
	Globals    []Object
	Frames     []*CallFrame
	FrameCount int
	LineInfo   []LineInfo
}

type CallFrame struct {
	Closure     *Closure
	Ip          int
	BasePointer int // Points to the base of stack
}

func (cf *CallFrame) Instructions() Instructions {
	return cf.Closure.Function.Instructions
}

func NewVM(bytecode *ByteCode) *VM {
	main := &CompiledFunction{Instructions: bytecode.Code, Name: "main"}
	closure := &Closure{Function: main}
	mainFrame := &CallFrame{Closure: closure, Ip: 0, BasePointer: 0}

	vm := &VM{}
	vm.Constants = bytecode.Constants
	vm.Stack = make([]Object, STACK_MAX)
	vm.Globals = make([]Object, MAX_GLOBALS)

	vm.Frames = make([]*CallFrame, FRAMES_MAX)
	vm.pushFrame(mainFrame)
	return vm
}

func (vm *VM) GetLine(insIndex int) int {
	accumulatedCount := 0

	for _, info := range vm.LineInfo {
		accumulatedCount += info.Count
		if insIndex < accumulatedCount {
			return info.Line
		}
	}

	return -1 // In case of an invalid instruction index
}

func (vm *VM) prinStackTrace(err error) {

	fmt.Println(err)

	for i := vm.FrameCount - 1; i >= 0; i-- {
		frame := vm.Frames[i]
		function := frame.Closure.Function

		opcode := OpCode(function.Instructions[frame.Ip-1])
		definition, ok := definitions[opcode]

		if !ok {
			return
		}

		line := vm.GetLine(frame.Ip - 1)
		fmt.Printf("[Instruction %s], ", definition.Name)
		fmt.Printf("[Line %d] in ", line)
		fmt.Printf("%s()\n", function.Name)
	}
}

func (vm *VM) currentFrame() *CallFrame {
	return vm.Frames[vm.FrameCount-1]
}

func (vm *VM) run() error {
	for {
		frame := vm.currentFrame()
		ip := &frame.Ip
		instructions := frame.Instructions()

		if *ip > len(instructions)-1 {
			return nil
		}

		opcode := OpCode(instructions[*ip])
		definition := definitions[opcode]
		line := vm.GetLine(*ip)
		fmt.Printf("Instruction: %s, Sp: %d, Ip: %d, Line: %d", definition.Name, vm.Sp, *ip, line)
		*ip += 1

		switch opcode {
		case OP_POP:
			vm.pop()
		case OP_RETURN:
			returnValue := vm.pop()
			if vm.FrameCount == 0 {
				return nil
			}
			frame := vm.popFrame()
			vm.Sp = frame.BasePointer - 1
			vm.push(returnValue)
		case OP_CLOSURE:
			index := ReadUint16(instructions[*ip:])
			*ip += 2
			nUpValues := ReadUint8(instructions[*ip:])
			*ip += 1
			object := vm.Constants[index]
			function, ok := object.(*CompiledFunction)
			if !ok {
				return fmt.Errorf("not a function +%v", object)
			}

			upvalues := make([]Object, nUpValues)
			for i := 0; i < int(nUpValues); i++ {
				upvalues[i] = vm.Stack[vm.Sp-int(nUpValues)+i]
			}
			vm.Sp = vm.Sp - int(nUpValues)

			closure := &Closure{Function: function, UpValues: upvalues}
			vm.push(closure)
		case OP_CLASS:
			index := ReadUint16(instructions[*ip:])
			vm.push(vm.Constants[index])
			*ip += 2
		case OP_FUNCTION:
			index := ReadUint16(instructions[*ip:])
			vm.push(vm.Constants[index])
			*ip += 2
		case OP_CALL:
			numArgs := ReadUint8(instructions[*ip:])
			err := vm.call(int(numArgs))
			if err != nil {
				return err
			}
			*ip += 1
		case OP_SET_PROPERTY:
			index := ReadUint8(instructions[*ip:])
			*ip += 1

			// Get property name from constatnts
			name := vm.Constants[index]

			// Ensure it is a string
			str, ok := name.(*StringObject)
			if !ok {
				return fmt.Errorf("property is not a string +%v", name)
			}

			object := vm.peek(1)

			instance, ok := object.(*CompiledInstanceObject)
			if !ok {
				return fmt.Errorf("only instance have properties, got +%v", object)
			}

			instance.Fields[str.Value] = vm.peek(0)
			// stored value
			value := vm.pop()
			// instance
			vm.pop()
			vm.push(value)
		case OP_GET_PROPERTY:
			index := ReadUint8(instructions[*ip:])
			*ip += 1
			object := vm.peek(0)
			name := vm.Constants[index]

			str, ok := name.(*StringObject)
			if !ok {
				fmt.Println("called1")
				return fmt.Errorf("property is not a string +%v", name)
			}

			instance, ok := object.(*CompiledInstanceObject)
			if !ok {
				fmt.Println("called2")
				return fmt.Errorf("only instance have properties, got +%v", object)
			}

			fmt.Printf("%s value , %v Field", str.Value, instance.Fields[str.Value])

			if value, ok := instance.Fields[str.Value]; ok {
				vm.pop()
				vm.push(value)
				continue
			}

			if err := vm.bindMethod(instance, str.Value); err != nil {
				return err
			}

		case OP_GET_BUILTIN:
			builinIndex := ReadUint8(instructions[*ip:])
			definition := Builtins[builinIndex]
			err := vm.push(definition.Builtin)
			if err != nil {
				return err
			}
			*ip += 1
		case OP_GET_GLOBAL:
			index := ReadUint16(instructions[*ip:])
			value := vm.Globals[index]
			err := vm.push(value)
			if err != nil {
				return err
			}
			*ip += 2
		case OP_GET_UPVALUE:
			index := ReadUint8(instructions[*ip:])
			*ip += 1

			closure := vm.currentFrame().Closure
			err := vm.push(closure.UpValues[index])
			if err != nil {
				return err
			}
		case OP_METHOD:
			index := ReadUint16(instructions[*ip:])
			*ip += 2
			object := vm.peek(0)
			object2 := vm.peek(1)
			name := vm.Constants[index]

			str, ok := name.(*StringObject)
			if !ok {
				return fmt.Errorf("function name is not a string +%v", name)
			}

			method, ok := object.(*Closure)
			if !ok {
				return fmt.Errorf("not a method: %+v", object)
			}

			class, ok := object2.(*CompiledClassObject)
			if !ok {
				return fmt.Errorf("not a Class: %+v", object2)
			}

			class.Methods[str.Value] = method

			// pop compiled function
			vm.pop()
		case OP_GET_LOCAL:
			index := ReadUint8(instructions[*ip:])
			err := vm.push(vm.Stack[frame.BasePointer+int(index)])
			if err != nil {
				return err
			}
			*ip += 1
		case OP_SET_LOCAL:
			index := ReadUint8(instructions[*ip:])
			*ip += 1
			value := vm.pop()
			vm.Stack[frame.BasePointer+int(index)] = value
		case OP_SET_GLOBAL:
			index := ReadUint16(instructions[*ip:])
			vm.Globals[index] = vm.pop()
			*ip += 2
		case OP_DEFINE_GLOBAL:
			index := ReadUint16(instructions[*ip:])
			vm.Globals[index] = vm.pop()
			*ip += 2
		case OP_DEFINE_LOCAL:
			index := ReadUint8(instructions[*ip:])
			*ip += 1
			value := vm.pop()
			vm.Stack[frame.BasePointer+int(index)] = value
		case OP_CONSTANT:
			index := ReadUint16(instructions[*ip:])
			*ip += 2

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
			offset := ReadUint16(instructions[*ip:])
			*ip += 2
			if !vm.isTruthy(vm.peek(0)) {
				fmt.Printf("Jumping by offset %d because top of stack is falsey\n", offset)
				*ip += int(offset)
				vm.pop()
			}
		case OP_JUMP:
			offset := ReadUint16(instructions[*ip:])
			fmt.Printf("Unconditional jump by offset %d\n", offset)
			*ip += int(offset)
		case OP_LOOP:
			offset := ReadUint16(instructions[*ip:])
			fmt.Printf("Looping back by offset %d\n", offset)
			*ip += 1
			*ip -= int(offset)
		}

		fmt.Printf(", NewSp: %d, BP: %d, Stack: %s\n", vm.Sp, frame.BasePointer, vm.printStack())
	}
}

func (vm *VM) call(numArgs int) error {
	value := vm.Stack[vm.Sp-1-numArgs]

	switch callee := value.(type) {
	case *CompiledBoundMethod:
		return vm.callBoundMethod(callee, numArgs)
	case *Closure:
		return vm.callFunction(callee, numArgs)
	case *Builtin:
		return vm.callBuiltin(callee, numArgs)
	case *CompiledClassObject:
		return vm.callClass(callee, numArgs)
	default:
		return fmt.Errorf("%s is not a function", value.Inspect())
	}
}

func (vm *VM) pushFrame(frame *CallFrame) {
	vm.Frames[vm.FrameCount] = frame
	vm.FrameCount++
}

func (vm *VM) popFrame() *CallFrame {
	vm.FrameCount--
	return vm.Frames[vm.FrameCount]
}

func (vm *VM) bindMethod(instance *CompiledInstanceObject, methodName string) error {
	if method, ok := instance.Class.Methods[methodName]; !ok {
		return fmt.Errorf("Undefined property %s.", methodName)
	} else {
		bound := &CompiledBoundMethod{Receiver: instance, Method: method}

		// pop the instance
		vm.pop()

		// push bound method on top of stack
		vm.push(bound)
		return nil
	}
}

func (vm *VM) callClass(class *CompiledClassObject, numArgs int) error {
	// check if the inti methods match
	instance := &CompiledInstanceObject{Class: class, Fields: make(map[string]Object)}

	// Check if the class has an "init" method (constructor)
	if initMethod, ok := class.Methods["init"]; ok {
		if initMethod.Function.NumParameters != numArgs {
			return fmt.Errorf("wrong number of arguments: want=%d, got=%d", initMethod.Function.NumParameters, numArgs)
		}

		vm.Stack[vm.Sp-1-numArgs] = instance
		err := vm.callBoundMethod(&CompiledBoundMethod{Receiver: instance, Method: initMethod}, numArgs)
		if err != nil {
			return err
		}
		vm.Stack[vm.Sp-1] = instance
		// account for this need to pop it
		return nil
	}

	vm.Stack[vm.Sp-1-numArgs] = instance
	// If there is no constructor, just adjust the stack pointer
	vm.Sp -= numArgs
	return nil
}

func (vm *VM) NewFrame(closure *Closure, bp int) *CallFrame {
	return &CallFrame{
		Closure:     closure,
		Ip:          0,
		BasePointer: bp,
	}
}

func (vm *VM) callFunction(callee *Closure, numArgs int) error {
	function := callee.Function
	if numArgs != function.NumParameters {
		return fmt.Errorf("wrong number of arguments: want=%d, got=%d", function.NumParameters, numArgs)
	}

	frame := &CallFrame{
		Closure:     callee,
		Ip:          0,
		BasePointer: vm.Sp - int(numArgs),
	}

	vm.pushFrame(frame)
	vm.Sp = frame.BasePointer + function.NumLocals
	fmt.Printf("\nElement at bp %v, Element at sp %v\n", vm.Stack[frame.BasePointer], vm.Stack[vm.Sp])
	return nil
}

func (vm *VM) callBoundMethod(callee *CompiledBoundMethod, numArgs int) error {
	closure := callee.Method
	function := closure.Function

	if numArgs != function.NumParameters {
		return fmt.Errorf("wrong number of arguments: want=%d, got=%d", function.NumParameters, numArgs)
	}

	frame := &CallFrame{
		Closure:     closure,
		Ip:          0,
		BasePointer: vm.Sp - numArgs,
	}

	copy(vm.Stack[frame.BasePointer+1:], vm.Stack[frame.BasePointer:vm.Sp])
	vm.Stack[frame.BasePointer] = callee.Receiver
	vm.Sp = frame.BasePointer + function.NumLocals

	vm.pushFrame(frame)
	fmt.Printf("BP: %d | SP: %d | STACK: %v\n", frame.BasePointer, vm.Sp, vm.printStack())
	return nil
}

func (vm *VM) callBuiltin(builtin *Builtin, numArgs int) error {
	args := vm.Stack[vm.Sp-numArgs : vm.Sp]
	fmt.Println(args)

	result := builtin.Fn(args...)
	vm.Sp = vm.Sp - numArgs - 1

	if result != nil {
		vm.push(result)
	} else {
		vm.push(Null)
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
		if vm.Stack[i] != nil {
			str.WriteString(vm.Stack[i].Inspect())
		} else {
			str.WriteString("nil")
		}

		if i != vm.Sp-1 {
			str.WriteString(",")
		}
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
