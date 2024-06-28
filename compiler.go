package main

import (
	"encoding/binary"
	"fmt"
)

const (
	UINT16_MAX = 1 << 16
)

type LineInfo struct {
	Line  int
	Count int
}

type Compiler struct {
	Constants   []Object
	LineInfo    []LineInfo
	SymbolTable *SymbolTable
	Scopes      []Scope
	ScopeIndex  int
}

type Scope struct {
	Instructions Instructions
}

type ByteCode struct {
	Code      Instructions
	Constants []Object
}

func (c *Compiler) ByteCode() *ByteCode {
	return &ByteCode{
		Code:      c.currentInstructions(),
		Constants: c.Constants,
	}
}

func NewCompiler() *Compiler {
	mainScope := Scope{
		Instructions: Instructions{},
	}

	symbolTable := NewSymbolTable()

	for name, _ := range builtins {
		symbolTable.DefineBuiltin(name)
	}

	return &Compiler{Constants: make([]Object, 0), LineInfo: make([]LineInfo, 0), SymbolTable: symbolTable, Scopes: []Scope{mainScope}, ScopeIndex: 0}
}

func NewCompilerWithState(symbolTable *SymbolTable) *Compiler {
	return &Compiler{Constants: make([]Object, 0), LineInfo: make([]LineInfo, 0), SymbolTable: symbolTable}
}

func (c *Compiler) Compile(ast Node) error {

	switch node := ast.(type) {
	case *Program:
		for _, s := range node.Statements {
			err := c.Compile(s)
			if err != nil {
				return err
			}
		}
	case *ExpressionStatement:
		err := c.Compile(node.Expression)
		if err != nil {
			return err
		}

		if _, ok := node.Expression.(*Assignment); !ok {
			c.WriteChunk(OP_POP, node.Token.Line)
		}
	case *BlockStatement:
		for _, s := range node.Statements {
			err := c.Compile(s)
			if err != nil {
				return err
			}
		}
	case *ReturnStatement:
		if c.ScopeIndex == 0 {
			return fmt.Errorf("Can't return from top-level code")
		}
		if err := c.Compile(node.ReturnValue); err != nil {
			return err
		}

		c.WriteChunk(OP_RETURN, node.Token.Line)
	case *CallExpression:
		if err := c.Compile(node.Callee); err != nil {
			return err
		}

		argCount := len(node.Arguments)

		if argCount >= 255 {
			return fmt.Errorf("Can't have more than 255 arguments.")
		}

		for _, arg := range node.Arguments {
			if err := c.Compile(arg); err != nil {
				return err
			}
		}

		c.WriteChunk(OP_CALL, node.Token.Line, argCount)
	case *ClassStatement:
		symbol := c.SymbolTable.Define(node.Name.Value)
		class := &CompiledClassObject{Name: node.Name, Methods: make(map[string]*Closure)}

		classIndex := c.MakeConstant(class)
		c.WriteChunk(OP_CLASS, node.Token.Line, classIndex)

		if c.ScopeIndex == 0 {
			c.WriteChunk(OP_DEFINE_GLOBAL, node.Token.Line, symbol.Index)
		} else {
			c.WriteChunk(OP_DEFINE_LOCAL, node.Token.Line, symbol.Index)
		}

		for index, method := range node.Methods {
			if err := c.CompileMethod(method, node.Name.Value); err != nil {
				return err
			}

			if index != len(node.Methods)-1 {
				c.WriteByte(byte(OP_POP))
			}
		}

		c.WriteByte(byte(OP_POP))
	case *This:
		symbol, ok := c.SymbolTable.Resolve("this")
		if !ok {
			return fmt.Errorf("undefined variable 'this'")
		}
		c.loadSymbol(symbol, node.Token.Line)
	case *FunctionDeclaration:

		symbol := c.SymbolTable.Define(node.Name.Value)
		c.enterScope()

		for _, p := range node.Params {
			c.SymbolTable.Define(p.Value)
		}

		if err := c.Compile(node.Body); err != nil {
			return err
		}

		if !c.lastInstructionIsReturn() {
			c.WriteChunk(OP_RETURN, node.Token.Line)
		}

		upvalues := c.SymbolTable.upvalues
		numLocals := c.SymbolTable.numDefinitions

		fmt.Printf("Bytecode for `%s`\n", node.Name.Value)
		c.DisassembleChunks()
		instructions := c.leaveScope()

		for _, upvalue := range upvalues {
			c.loadSymbol(upvalue, node.Token.Line)
		}

		compiledFunction := &CompiledFunction{
			Instructions:  instructions,
			NumLocals:     numLocals,
			NumParameters: len(node.Params),
			Name:          node.Name.Value,
		}

		fnIndex := c.MakeConstant(compiledFunction)
		c.WriteChunk(OP_CLOSURE, node.Token.Line, fnIndex, len(upvalues))

		if c.ScopeIndex == 0 {
			c.WriteChunk(OP_DEFINE_GLOBAL, node.Token.Line, symbol.Index)
		} else {
			c.WriteChunk(OP_DEFINE_LOCAL, node.Token.Line, symbol.Index)
		}
	case *VarStatement:
		_, ok := c.SymbolTable.ResolveInner(node.Identifier.Value)
		if ok {
			return fmt.Errorf("Already variable with this name in this scope: %s", node.Identifier.Value)
		}
		symbol := c.SymbolTable.Define(node.Identifier.Value)

		if node.Expression != nil {
			err := c.Compile(node.Expression)
			if err != nil {
				return err
			}
		} else {
			c.WriteChunk(OP_NIL, node.Token.Line)
		}

		if symbol.Scope == GLOBAL_SCOPE {
			c.WriteChunk(OP_DEFINE_GLOBAL, node.Token.Line, symbol.Index)
		} else {
			c.WriteChunk(OP_DEFINE_LOCAL, node.Token.Line, symbol.Index)
		}
	case *While:
		loopStart := len(c.currentInstructions())

		if err := c.Compile(node.Condition); err != nil {
			return err
		}

		c.WriteChunk(OP_JUMP_IF_FALSE, node.Token.Line, 9999)
		exitJump := len(c.currentInstructions()) - 2 // offset of the emitted instruction
		c.WriteByte(byte(OP_POP))

		if len(node.Body.Statements) != 0 {
			if err := c.Compile(node.Body); err != nil {
				return err
			}

			if c.lastInstructionIsPop() {
				c.removeLastPop()
			}
		} else {
			c.WriteChunk(OP_NIL, node.Token.Line)
			c.WriteByte(byte(OP_POP))
		}

		offset := len(c.currentInstructions()) - loopStart + 2
		c.WriteChunk(OP_LOOP, node.Token.Line, offset)

		if err := c.patchJump(exitJump); err != nil {
			return err
		}

		c.WriteByte(byte(OP_POP))

	case *For:
		c.enterScope()

		if node.Initializer != nil {
			if err := c.Compile(node.Initializer); err != nil {
				return err
			}

			if c.lastInstructionIsPop() {
				c.removeLastPop()
			}
		}

		conditionalStart := len(c.currentInstructions())

		exitJump := -1
		if node.Condition != nil {
			if err := c.Compile(node.Condition); err != nil {
				return err
			}

			if c.lastInstructionIsPop() {
				c.removeLastPop()
			}

			c.WriteChunk(OP_JUMP_IF_FALSE, node.Token.Line, 9999)
			exitJump = len(c.currentInstructions()) - 2
			c.WriteByte(byte(OP_POP))
		}

		if err := c.Compile(node.Body); err != nil {
			return err
		}

		if node.Increment != nil {
			if err := c.Compile(node.Increment); err != nil {
				return err
			}

			if c.lastInstructionIsPop() {
				c.removeLastPop()
			}
		}

		offset := len(c.currentInstructions()) - conditionalStart + 2
		c.WriteChunk(OP_LOOP, node.Token.Line, offset)

		if exitJump != -1 {
			if err := c.patchJump(exitJump); err != nil {
				return err
			}
			c.WriteByte(byte(OP_POP))
		}

		c.leaveScope()
	case *GetExpression:
		if err := c.Compile(node.Object); err != nil {
			return err
		}

		constant := c.MakeConstant(&StringObject{Value: node.Property.Value})
		c.WriteChunk(OP_GET_PROPERTY, node.Token.Line, constant)
	case *SetExpression:
		if err := c.Compile(node.Object); err != nil {
			return err
		}

		if err := c.Compile(node.Value); err != nil {
			return err
		}

		constant := c.MakeConstant(&StringObject{Value: node.Property.Value})
		c.WriteChunk(OP_SET_PROPERTY, node.Token.Line, constant)
	case *Assignment:
		if symbol, ok := c.SymbolTable.Resolve(node.Identifier.Value); !ok {
			return fmt.Errorf("Undeclared identifier: %s", node.Identifier.Value)
		} else {
			if node.Expression != nil {
				err := c.Compile(node.Expression)
				if err != nil {
					return err
				}
			} else {
				c.WriteChunk(OP_NIL, node.Token.Line)
			}

			c.setSymbol(symbol, node.Token.Line)
		}
	case *Identifier:
		symbol, ok := c.SymbolTable.Resolve(node.Value)
		if !ok {
			return fmt.Errorf("Undefined variable: %s", node.Value)
		}
		c.loadSymbol(symbol, node.Token.Line)
	case *StringLiteral:
		value := &StringObject{Value: node.Value}
		c.WriteChunk(OP_CONSTANT, node.Token.Line, c.MakeConstant(value))
	case *NumberLiteral:
		value := &FloatObject{Value: node.Value}
		c.WriteChunk(OP_CONSTANT, node.Token.Line, c.MakeConstant(value))
	case *Unary:
		err := c.Compile(node.Right)
		if err != nil {
			return err
		}

		switch node.Operator {
		case "-":
			c.WriteChunk(OP_NEGATE, node.Token.Line)
		case "!":
			c.WriteChunk(OP_NOT, node.Token.Line)
		default:
			fmt.Errorf("unknown unary operator: %s", node.Operator)

		}
	case *BooleanLiteral:
		if node.Value {
			c.WriteChunk(OP_TRUE, node.Token.Line)
		} else {
			c.WriteChunk(OP_FALSE, node.Token.Line)
		}
	case *NilLiteral:
		c.WriteChunk(OP_NIL, node.Token.Line)
	case *GroupedExpression:
		c.Compile(node.Expression)
	case *Binary:
		left := c.Compile(node.Left)
		if left != nil {
			return left
		}

		right := c.Compile(node.Right)
		if right != nil {
			return right
		}

		switch node.Operator {
		case "+":
			c.WriteChunk(OP_ADD, node.Token.Line)
		case "-":
			c.WriteChunk(OP_SUBTRACT, node.Token.Line)
		case "/":
			c.WriteChunk(OP_DIVIDE, node.Token.Line)
		case "*":
			c.WriteChunk(OP_MULTIPLY, node.Token.Line)
		case "!=":
			c.WriteChunk(OP_EQUAL, node.Token.Line)
			c.WriteChunk(OP_NOT, node.Token.Line)
		case "==":
			c.WriteChunk(OP_EQUAL, node.Token.Line)
		case ">":
			c.WriteChunk(OP_GREATER, node.Token.Line)
		case ">=":
			c.WriteChunk(OP_LESS, node.Token.Line)
			c.WriteChunk(OP_NOT, node.Token.Line)
		case "<":
			c.WriteChunk(OP_LESS, node.Token.Line)
		case "<=":
			c.WriteChunk(OP_GREATER, node.Token.Line)
			c.WriteChunk(OP_NOT, node.Token.Line)
		}
	case *Logical:
		switch node.Operator {
		case "and":
			err := c.and(node)
			if err != nil {
				return err
			}
		case "or":
			err := c.or(node)
			if err != nil {
				return err
			}
		default:
			return fmt.Errorf("Invalid logical operator: %s", node.Operator)
		}
	case *IfStatement:
		condition := c.Compile(node.Condition)
		if condition != nil {
			return condition
		}

		c.WriteChunk(OP_JUMP_IF_FALSE, node.Token.Line, 9999)
		thenJump := len(c.currentInstructions()) - 2 // offset of the emitted instruction

		// in case not expression statement remove condition
		if c.lastInstruction() != byte(OP_POP) {
			c.WriteChunk(OP_POP, node.Token.Line)
		}

		err := c.Compile(node.ThenBranch)
		if err != nil {
			return err
		}

		// Emit OP_JUMP with a placeholder offset
		c.WriteChunk(OP_JUMP, node.Token.Line, 9999)
		elseJump := len(c.currentInstructions()) - 2 // offset of the emitted instruction for else

		err = c.patchJump(thenJump)
		if err != nil {
			return err
		}
		if c.lastInstruction() != byte(OP_POP) {
			c.WriteChunk(OP_POP, node.Token.Line)
		}

		if node.ElseBranch != nil {
			err = c.Compile(node.ElseBranch)
			if err != nil {
				return err
			}

		} else {
			c.WriteChunk(OP_NIL, node.Token.Line)
		}

		err = c.patchJump(elseJump)
		if err != nil {
			return err
		}

	}

	return nil
}

func (c *Compiler) CompileMethod(method *MethodDeclaration, className string) error {
	symbol, ok := c.SymbolTable.Resolve(className)
	if !ok {
		return fmt.Errorf("Undefined class: %s", className)
	}
	c.loadSymbol(symbol, method.Token.Line)

	methodName := c.MakeConstant(&StringObject{Value: method.Name.Value})
	c.enterScope()

	c.SymbolTable.Define("this")

	for _, p := range method.Params {
		c.SymbolTable.Define(p.Value)
	}

	if err := c.Compile(method.Body); err != nil {
		return err
	}

	if !c.lastInstructionIsReturn() {
		c.WriteChunk(OP_RETURN, method.Token.Line)
	}

	upvalues := c.SymbolTable.upvalues
	numLocals := c.SymbolTable.numDefinitions

	fmt.Printf("Bytecode for `%s`\n", method.Name.Value)
	c.DisassembleChunks()
	instructions := c.leaveScope()

	for _, upvalue := range upvalues {
		c.loadSymbol(upvalue, method.Token.Line)
	}

	compiledFunction := &CompiledFunction{
		Instructions:  instructions,
		NumLocals:     numLocals,
		NumParameters: len(method.Params),
		Name:          method.Name.Value,
	}

	fnIndex := c.MakeConstant(compiledFunction)
	c.WriteChunk(OP_CLOSURE, method.Token.Line, fnIndex, len(upvalues))

	c.WriteChunk(OP_METHOD, method.Token.Line, methodName)

	return nil
}

func (c *Compiler) loadSymbol(symbol Symbol, line int) {
	switch symbol.Scope {
	case GLOBAL_SCOPE:
		c.WriteChunk(OP_GET_GLOBAL, line, symbol.Index)
	case LOCAL_SCOPE:
		c.WriteChunk(OP_GET_LOCAL, line, symbol.Index)
	case BUILTIN_SCOPE:
		c.WriteChunk(OP_GET_BUILTIN, line, symbol.Index)
	case UPVALUE_SCOPE:
		c.WriteChunk(OP_GET_UPVALUE, line, symbol.Index)
	}
}

func (c *Compiler) setSymbol(symbol Symbol, line int) {
	switch symbol.Scope {
	case GLOBAL_SCOPE:
		c.WriteChunk(OP_SET_GLOBAL, line, symbol.Index)
	case LOCAL_SCOPE:
		c.WriteChunk(OP_SET_LOCAL, line, symbol.Index)
	}
}

func (c *Compiler) or(node *Logical) error {
	if err := c.Compile(node.Left); err != nil {
		return err
	}

	// when falsey does a tiny jump over the unconditional jump over the code for right operand
	c.WriteChunk(OP_JUMP_IF_FALSE, node.Token.Line, 9999)
	elseJump := len(c.currentInstructions()) - 2

	c.WriteChunk(OP_JUMP, node.Token.Line, 9999)
	endJump := len(c.currentInstructions()) - 2

	if err := c.patchJump(elseJump); err != nil {
		return err
	}
	if c.lastInstruction() != byte(OP_POP) {
		c.WriteChunk(OP_POP, node.Token.Line)
	}

	if err := c.Compile(node.Right); err != nil {
		return err
	}

	if err := c.patchJump(endJump); err != nil {
		return err
	}

	return nil
}

func (c *Compiler) and(node *Logical) error {
	left := c.Compile(node.Left)
	if left != nil {
		return left
	}

	c.WriteChunk(OP_JUMP_IF_FALSE, node.Token.Line, 9999)
	endJump := len(c.currentInstructions()) - 2

	if c.lastInstruction() != byte(OP_POP) {
		c.WriteChunk(OP_POP, node.Token.Line)
	}

	right := c.Compile(node.Right)
	if right != nil {
		return right
	}

	err := c.patchJump(endJump)
	if err != nil {
		return err
	}
	return nil

}

func (c *Compiler) MakeConstant(value Object) int {
	return int(c.writeValue(value))
}

func (c *Compiler) DisassembleChunks() {
	var offset int

	length := len(c.currentInstructions())

	for offset < length {
		offset = c.disassembleInstruction(offset)
	}
}

func (c *Compiler) WriteByte(b byte) {
	c.Scopes[c.ScopeIndex].Instructions = append(c.Scopes[c.ScopeIndex].Instructions, b)
}

func (c *Compiler) WriteChunk(opcode OpCode, line int, operands ...int) {
	if len(c.LineInfo) > 0 && c.LineInfo[len(c.LineInfo)-1].Line == line {
		c.LineInfo[len(c.LineInfo)-1].Count++
	} else {
		c.LineInfo = append(c.LineInfo, LineInfo{Line: line, Count: 1})
	}
	c.Scopes[c.ScopeIndex].Instructions = append(c.Scopes[c.ScopeIndex].Instructions, byte(opcode))
	definition := definitions[opcode]

	for i, o := range operands {
		operandWidth := definition.OperandWidths[i]
		switch operandWidth {
		case 2:
			c.Scopes[c.ScopeIndex].Instructions = binary.BigEndian.AppendUint16(c.Scopes[c.ScopeIndex].Instructions, uint16(o))
		case 1:
			c.Scopes[c.ScopeIndex].Instructions = append(c.Scopes[c.ScopeIndex].Instructions, byte(o))
		}
	}
}

func (c *Compiler) GetLine(insIndex int) int {
	accumulatedCount := 0

	for _, info := range c.LineInfo {
		accumulatedCount += info.Count
		if insIndex < accumulatedCount {
			return info.Line
		}
	}

	return -1 // In case of an invalid instruction index
}

func (c *Compiler) lastInstruction() byte {
	instructions := c.currentInstructions()
	return instructions[len(instructions)-1]
}

func (c *Compiler) patchJump(jump int) error {
	offset := len(c.currentInstructions()) - jump - 2 // calculate jump offset

	if offset > UINT16_MAX {
		return fmt.Errorf("Too much code to jump over.")
	}

	if jump < 0 || jump >= len(c.currentInstructions()) {
		return fmt.Errorf("Invalid jump index: %d", jump)
	}

	binary.BigEndian.PutUint16(c.Scopes[c.ScopeIndex].Instructions[jump:], uint16(offset))
	return nil
}

func (c *Compiler) identifierConstant(name string) int {
	value := &StringObject{Value: name}
	global := c.MakeConstant(value)
	return global
}

func (c *Compiler) currentInstructions() Instructions {
	return c.Scopes[c.ScopeIndex].Instructions
}

func (c *Compiler) enterScope() {
	scope := Scope{
		Instructions: Instructions{},
	}
	c.Scopes = append(c.Scopes, scope)
	c.ScopeIndex++
	newSymbolTable := NewEnclosedSymbolTable(c.SymbolTable)
	c.SymbolTable = newSymbolTable
}

func (c *Compiler) leaveScope() Instructions {
	instructions := c.currentInstructions()

	c.Scopes = c.Scopes[:len(c.Scopes)-1]
	c.ScopeIndex -= 1
	c.SymbolTable = c.SymbolTable.Outer
	return instructions
}

func (c *Compiler) writeValue(value Object) uint16 {
	c.Constants = append(c.Constants, value)
	return uint16((len(c.Constants) - 1))
}

func (c *Compiler) disassembleInstruction(offset int) int {
	instructions := c.currentInstructions()
	definition, err := Lookup(instructions[offset])
	newOffset := offset + 1

	if err != nil {
		fmt.Println(err)
		return newOffset
	}
	fmt.Printf("%04d %4d %s", offset, c.GetLine(offset), definition.Name)

	for _, w := range definition.OperandWidths {
		switch w {
		case 2:
			fmt.Printf(" %v", ReadUint16(c.Scopes[c.ScopeIndex].Instructions[newOffset:]))
		case 1:
			fmt.Printf(" %v", ReadUint8(c.Scopes[c.ScopeIndex].Instructions[newOffset:]))
		}
		newOffset += w
	}
	fmt.Printf("\n")

	return newOffset
}

func (c *Compiler) lastInstructionIsReturn() bool {
	instructions := c.currentInstructions()
	return instructions[len(instructions)-1] == byte(OP_RETURN)
}

func (c *Compiler) lastInstructionIsPop() bool {
	instructions := c.currentInstructions()
	return instructions[len(instructions)-1] == byte(OP_POP)
}

func (c *Compiler) removeLastPop() {
	instructions := c.currentInstructions()
	c.Scopes[c.ScopeIndex].Instructions = c.Scopes[c.ScopeIndex].Instructions[:len(instructions)-1]
}

func ReadUint16(ins Instructions) uint16 {
	return binary.BigEndian.Uint16(ins)
}

func ReadUint8(ins Instructions) uint8 {
	return uint8(ins[0])
}
