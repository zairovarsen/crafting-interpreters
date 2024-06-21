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
	Code        Instructions
	Constants   []Object
	LineInfo    []LineInfo
	SymbolTable *SymbolTable
}

type ByteCode struct {
	Code      Instructions
	Constants []Object
}

func (c *Compiler) ByteCode() *ByteCode {
	return &ByteCode{
		Code:      c.Code,
		Constants: c.Constants,
	}
}

func NewCompiler() *Compiler {

	symbolTable := NewSymbolTable()

	for name, _ := range builtins {
		symbolTable.DefineBuiltin(name)
	}

	return &Compiler{Code: make([]byte, 0), Constants: make([]Object, 0), LineInfo: make([]LineInfo, 0), SymbolTable: symbolTable}
}

func NewCompilerWithState(symbolTable *SymbolTable) *Compiler {
	return &Compiler{Code: make([]byte, 0), Constants: make([]Object, 0), LineInfo: make([]LineInfo, 0), SymbolTable: symbolTable}
}

func ReadUint16(ins Instructions) uint16 {
	return binary.BigEndian.Uint16(ins)
}

func ReadUint8(ins Instructions) uint8 {
	return uint8(ins[0])
}

func (c *Compiler) MakeConstant(value Object) int {
	return int(c.writeValue(value))
}

func (c *Compiler) DisassembleChunks() {
	var offset int

	fmt.Println(len(c.Code))
	for offset < len(c.Code) {
		offset = c.disassembleInstruction(offset)
	}
}

func (c *Compiler) WriteByte(b byte) {
	c.Code = append(c.Code, b)
}

func (c *Compiler) WriteChunk(opcode OpCode, line int, operands ...int) {
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

func (c *Compiler) Compile(ast Node) error {

	switch node := ast.(type) {
	case *Program:
		for _, s := range node.Statements {
			err := c.Compile(s)
			if err != nil {
				return err
			}
		}
		c.WriteChunk(OP_RETURN, c.LineInfo[len(c.LineInfo)-1].Line)
	case *ExpressionStatement:
		err := c.Compile(node.Expression)
		if err != nil {
			return err
		}
		// c.WriteChunk(OP_POP, node.Token.Line)
	case *BlockStatement:
		for _, s := range node.Statements {
			err := c.Compile(s)
			if err != nil {
				return err
			}
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

			if symbol.Scope == GLOBAL_SCOPE {
				c.WriteChunk(OP_SET_GLOBAL, node.Token.Line, symbol.Index)
			} else {
				c.WriteChunk(OP_SET_LOCAL, node.Token.Line, symbol.Index)
			}
		}
	case *Identifier:
		symbol, ok := c.SymbolTable.Resolve(node.Value)
		if !ok {
			return fmt.Errorf("Undefined variable: %s", node.Value)
		}
		switch symbol.Scope {
		case GLOBAL_SCOPE:
			c.WriteChunk(OP_GET_GLOBAL, node.Token.Line, symbol.Index)
		case LOCAL_SCOPE:
			c.WriteChunk(OP_GET_LOCAL, node.Token.Line, symbol.Index)
		}
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
	case *IfStatement:
		condition := c.Compile(node.Condition)
		if condition != nil {
			return condition
		}

		c.WriteChunk(OP_JUMP_IF_FALSE, node.Token.Line, 9999)
		thenJump := len(c.Code) - 2 // offset of the emitted instruction
		c.WriteChunk(OP_POP, node.Token.Line)

		err := c.Compile(node.ThenBranch)
		if err != nil {
			return err
		}

		err = c.patchJump(thenJump)
		if err != nil {
			return err
		}
		c.WriteChunk(OP_POP, node.Token.Line)

		c.WriteChunk(OP_JUMP, node.Token.Line, 9999)
		elseJump := len(c.Code) - 2

		err = c.Compile(node.ElseBranch)
		if err != nil {
			return err
		}

		c.patchJump(elseJump)
	}

	return nil
}

func (c *Compiler) patchJump(jump int) error {
	offset := len(c.Code) - jump - 2 // calculate jump offset

	if offset > UINT16_MAX {
		return fmt.Errorf("Too much code to jump over.")
	}

	binary.BigEndian.PutUint16(c.Code[jump:], uint16(offset))
	return nil
}

func (c *Compiler) identifierConstant(name string) int {
	value := &StringObject{Value: name}
	global := c.MakeConstant(value)
	return global
}

func (c *Compiler) writeValue(value Object) uint16 {
	c.Constants = append(c.Constants, value)
	return uint16((len(c.Constants) - 1))
}

func (c *Compiler) disassembleInstruction(offset int) int {
	definition, err := Lookup(c.Code[offset])
	newOffset := offset + 1

	if err != nil {
		fmt.Println(err)
		return newOffset
	}
	fmt.Printf("%04d %4d %s", offset, c.GetLine(offset), definition.Name)

	for _, w := range definition.OperandWidths {
		switch w {
		case 2:
			fmt.Printf(" %v", ReadUint16(c.Code[newOffset:]))
		case 1:
			fmt.Printf(" %v", ReadUint8(c.Code[newOffset:]))
		}
		newOffset += w
	}
	fmt.Printf("\n")

	return newOffset
}
