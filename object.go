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
	ClassObj            = "Class"
	InstanceObj         = "Instance"
	BoundObj            = "BoundObj"
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

type ClassObject struct {
	Name    *Identifier
	Methods map[string]*Function
}

func (c *ClassObject) Type() ObjectType { return ClassObj }
func (c *ClassObject) Inspect() string {
	var str strings.Builder

	var methods []string
	for _, m := range c.Methods {
		methods = append(methods, m.Inspect())
	}

	str.WriteString("class")
	if c.Name != nil {
		str.WriteString(" " + c.Name.Value)
	}
	str.WriteString("{\n")
	str.WriteString(strings.Join(methods, "\n"))
	str.WriteString("}")

	return str.String()
}

type InstanceObject struct {
	Class  *ClassObject
	Fields map[string]Object
}

func (io *InstanceObject) Type() ObjectType { return InstanceObj }
func (io *InstanceObject) Inspect() string {
	fields := []string{}
	for key, value := range io.Fields {
		fields = append(fields, fmt.Sprintf("%s: %v", key, value))
	}
	return fmt.Sprintf("<instance of %s> {%s}", io.Class.Name, strings.Join(fields, ", "))
}

func (io *InstanceObject) GetField(name string) (Object, bool) {
	value, ok := io.Fields[name]
	return value, ok
}

func (io *InstanceObject) SetField(name string, value Object) {
	io.Fields[name] = value
}

func (io *InstanceObject) GetMethod(name string) (*Function, bool) {
	method, ok := io.Class.Methods[name]
	return method, ok
}

type BoundMethod struct {
	Receiver *InstanceObject
	Method   *Function
}

func (bm *BoundMethod) Type() ObjectType { return BoundObj }
func (bm *BoundMethod) Inspect() string {
	return fmt.Sprintf("<bound method %s of %s>", bm.Method.Name.Value, bm.Receiver.Class.Name)
}

type Function struct {
	Name       *Identifier
	Parameters []*Identifier
	Body       *BlockStatement
	Env        *Environment // capture current environment for closures
}

func (f *Function) Type() ObjectType { return FunctionObj }
func (f *Function) Inspect() string {
	var str strings.Builder

	var params []string
	for _, p := range f.Parameters {
		params = append(params, p.String())
	}

	str.WriteString("function")
	if f.Name != nil {
		str.WriteString(" " + f.Name.Value)
	}
	str.WriteString("(")
	str.WriteString(strings.Join(params, ", "))
	str.WriteString(") {\n")
	str.WriteString(f.Body.String())
	str.WriteString("\n}")

	return str.String()
}
