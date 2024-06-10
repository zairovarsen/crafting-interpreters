package main

import (
	"fmt"
	"strings"
)

type ObjectType string

const (
	ContinueObj         = "Continue"
	BreakObj            = "Break"
	FloatObj            = "Float"
	BooleanObj          = "Boolean"
	NillObj             = "Nill"
	ReturnValueObj      = "ReturnValue"
	ErrorObj            = "Error"
	FunctionObj         = "Function"
	StringObj           = "String"
	BuiltinObj          = "Builtin"
	ArrayObj            = "Array"
	HashObj             = "Hash"
	CompiledFunctionObj = "CompiledFunction"
	ClosureObj          = "Closure"
)

type BuiltinFunction func(args ...Object) Object

type Builtin struct {
	Fn BuiltinFunction
}

func (b *Builtin) Type() ObjectType { return BuiltinObj }
func (b *Builtin) Inspect() string  { return "builtin function" }

type Object interface {
	Type() ObjectType
	Inspect() string
}

type ContinueSignal struct{}

func (b *ContinueSignal) Type() ObjectType { return ContinueObj }
func (b *ContinueSignal) Inspect() string  { return "CONTINUE" }

type BreakSignal struct{}

func (b *BreakSignal) Type() ObjectType { return BreakObj }
func (b *BreakSignal) Inspect() string  { return "BREAK" }

type ErrorObject struct {
	Message string
}

func (e *ErrorObject) Type() ObjectType { return ErrorObj }
func (e *ErrorObject) Inspect() string  { return "ERROR: " + e.Message }

type FloatObject struct {
	Value float64
}

func (f *FloatObject) Type() ObjectType { return FloatObj }
func (f *FloatObject) Inspect() string  { return fmt.Sprintf("%.0f", f.Value) }

type BooleanObject struct {
	Value bool
}

func (b *BooleanObject) Type() ObjectType { return BooleanObj }
func (b *BooleanObject) Inspect() string  { return fmt.Sprintf("%t", b.Value) }

type StringObject struct {
	Value string
}

func (s *StringObject) Type() ObjectType { return StringObj }
func (s *StringObject) Inspect() string  { return s.Value }

type NilObject struct {
}

func (n *NilObject) Type() ObjectType { return NillObj }
func (n *NilObject) Inspect() string  { return "nill" }

type ReturnValueObject struct {
	Value Object
}

func (r *ReturnValueObject) Type() ObjectType { return ReturnValueObj }
func (r *ReturnValueObject) Inspect() string  { return r.Value.Inspect() }

type Function struct {
	Parameters []*Identifier
	Body       *BlockStatement
	Env        *Environment
}

func (f *Function) Type() ObjectType { return FunctionObj }
func (f *Function) Inspect() string {
	var str strings.Builder

	var params []string
	for _, p := range f.Parameters {
		params = append(params, p.String())
	}

	str.WriteString("function")
	str.WriteString("(")
	str.WriteString(strings.Join(params, ", "))
	str.WriteString(") {\n")
	str.WriteString(f.Body.String())
	str.WriteString("\n}")

	return str.String()
}
