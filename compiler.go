package main

import (
	"encoding/binary"
	"fmt"
)

type LineInfo struct {
	Line  int
	Count int
}

type Compiler struct {
	Code      Instructions
	Constants []Object
	LineInfo  []LineInfo
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
	return &Compiler{Code: make([]byte, 0), Constants: make([]Object, 0), LineInfo: make([]LineInfo, 0)}
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

	for offset < len(c.Code) {
		offset = c.disassembleInstruction(offset)
	}
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
	}

	return nil
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
			fmt.Printf(" %f", c.Constants[ReadUint16(c.Code[newOffset:])])
		case 1:
			fmt.Printf(" %f", c.Constants[ReadUint8(c.Code[newOffset:])])
		}
		newOffset += w
	}
	fmt.Printf("\n")

	return newOffset
}
