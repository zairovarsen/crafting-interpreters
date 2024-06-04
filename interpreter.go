package main

import "fmt"

const (
	unknownOperatorError    = "unknown operator"
	typeMissMatchError      = "type mismatch"
	divisionByZero          = "divide by zero"
	identifierNotFoundError = "identifier not found"
	notFunctionError        = "not a function"
)

var (
	Null  = &NilObject{}
	True  = &BooleanObject{Value: true}
	False = &BooleanObject{Value: false}
)

// java way is terrible but whatever :)
type Visitor interface {
	VisitProgram(node *Program) Object
	VisitIdentifier(node *Identifier) Object
	VisitBinary(node *Binary) Object
	VisitUnary(node *Unary) Object
	VisitStringLiteral(node *StringLiteral) Object
	VisitNumberLiteral(node *NumberLiteral) Object
	VisitBooleanLiteral(node *BooleanLiteral) Object
	VisitNilLiteral(node *NilLiteral) Object
	VisitGroupedExpression(node *GroupedExpression) Object
	VisitAssignment(node *Assignment) Object
	VisitCommaExpression(node *CommaExpression) Object
	VisitTernaryExpression(node *TernaryExpression) Object
	VisitBlockStatement(node *BlockStatement) Object
	VisitExpressionStatement(node *ExpressionStatement) Object
	VisitReturnStatement(node *ReturnStatement) Object
	VisitPrintStatement(node *PrintStatement) Object
	VisitVarStatement(node *VarStatement) Object
}

type Environment struct {
	store map[string]Object
}

func NewEnvironment() *Environment {
	s := make(map[string]Object)
	return &Environment{store: s}
}

func (e *Environment) Set(name string, value Object) Object {
	e.store[name] = value
	return value
}

func (e *Environment) Get(name string) (Object, bool) {
	obj, ok := e.store[name]
	return obj, ok
}

type Interpreter struct {
	env *Environment
}

func NewInterpreter(env *Environment) *Interpreter {
	return &Interpreter{env}
}

func (i *Interpreter) nativeToBooleanObject(input bool) Object {
	if input {
		return True
	}
	return False
}

func (i *Interpreter) isError(obj Object) bool {
	if obj != nil {
		return obj.Type() == ErrorObj
	}
	return false
}

func (i *Interpreter) isTruthy(value Object) bool {
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

func (i *Interpreter) newError(format string, args ...interface{}) Object {
	msg := fmt.Sprintf(format, args...)
	return &ErrorObject{msg}
}

func (i *Interpreter) VisitPrintStatement(node *PrintStatement) Object {
	value := node.PrintValue.Accept(i)
	fmt.Println(value.Inspect())
	return nil
}

func (i *Interpreter) VisitIdentifier(node *Identifier) Object {
	obj, ok := i.env.Get(node.Value)
	if !ok {
		return i.newError("%s: %s", identifierNotFoundError, node.Value)
	}
	return obj
}

func (i *Interpreter) VisitGroupedExpression(node *GroupedExpression) Object {
	return nil
}

func (i *Interpreter) VisitVarStatement(node *VarStatement) Object {
	right := node.Expression.Accept(i)

	if !i.isError(right) {
		// define a variable
		i.env.Set(node.Identifier.Value, right)
	}

	return right
}

func (i *Interpreter) VisitBinary(node *Binary) Object {
	left := node.Left.Accept(i)
	right := node.Right.Accept(i)

	switch {
	case left.Type() == StringObj && right.Type() == StringObj:
		return i.stringarithmetic(left, right, node.Operator)
	case left.Type() == FloatObj && right.Type() == FloatObj:
		return i.floatarithmetic(left, right, node.Operator)
	case node.Operator == "+" && ((left.Type() == FloatObj && right.Type() == StringObj) || (right.Type() == FloatObj && left.Type() == StringObj)):
		if left.Type() == FloatObj {
			left = &StringObject{Value: left.Inspect()}
		} else {
			right = &StringObject{Value: right.Inspect()}
		}
		return i.stringarithmetic(left, right, node.Operator)
	case node.Operator == "==":
		return i.nativeToBooleanObject(left == right)
	case node.Operator == "!=":
		return i.nativeToBooleanObject(left != right)
	case left.Type() != right.Type():
		return i.newError("%s: %s %s %s", typeMissMatchError, left.Type(), node.Operator, right.Type())
	default:
		return i.newError("%s: %s %s %s", unknownOperatorError, left.Type(), node.Operator, right.Type())
	}
}

func (i *Interpreter) stringarithmetic(left Object, right Object, op string) Object {
	l, _ := left.(*StringObject)
	r, _ := right.(*StringObject)

	switch op {
	case "+":
		return &StringObject{Value: l.Value + r.Value}
	case ">":
		return &BooleanObject{Value: l.Value > r.Value}
	case "<":
		return &BooleanObject{Value: l.Value < r.Value}
	case ">=":
		return &BooleanObject{Value: l.Value >= r.Value}
	case "<=":
		return &BooleanObject{Value: l.Value <= r.Value}
	case "==":
		return &BooleanObject{Value: l.Value == r.Value}
	default:
		return i.newError("%s: %s %s %s", unknownOperatorError, left.Type(), op, right.Type())
	}
}

func (i *Interpreter) floatarithmetic(left Object, right Object, op string) Object {
	l, _ := left.(*FloatObject)
	r, _ := right.(*FloatObject)

	switch op {
	case "+":
		return &FloatObject{Value: l.Value + r.Value}
	case "-":
		return &FloatObject{Value: l.Value - r.Value}
	case "/":
		if r.Value == 0 {
			return i.newError("%s: %f %s %f", divisionByZero, l.Value, op, r.Value)
		}
		return &FloatObject{Value: l.Value / r.Value}
	case "*":
		return &FloatObject{Value: l.Value * r.Value}
	case ">":
		return &BooleanObject{Value: l.Value > r.Value}
	case "<":
		return &BooleanObject{Value: l.Value < r.Value}
	case ">=":
		return &BooleanObject{Value: l.Value >= r.Value}
	case "<=":
		return &BooleanObject{Value: l.Value <= r.Value}
	case "==":
		return &BooleanObject{Value: l.Value == r.Value}
	default:
		return i.newError("%s: %s %s %s", unknownOperatorError, left.Type(), op, right.Type())
	}
}

func (i *Interpreter) VisitUnary(node *Unary) Object {
	right := node.Right.Accept(i)

	switch node.Operator {
	case "-":
		old, ok := right.(*FloatObject)
		if !ok {
			return i.newError("%s: %s", typeMissMatchError, fmt.Sprintf("operand must be a float, got=%s", right.Type()))
		}
		return &FloatObject{Value: -old.Value}
	case "!":
		return &BooleanObject{Value: !i.isTruthy(right)}
	default:
		return i.newError("%s: %s%s", unknownOperatorError, node.Operator, right.Type())
	}
}

func (i *Interpreter) VisitStringLiteral(node *StringLiteral) Object {
	return &StringObject{Value: node.Value}
}

func (i *Interpreter) VisitNumberLiteral(node *NumberLiteral) Object {
	return &FloatObject{Value: node.Value}
}

func (i *Interpreter) VisitBooleanLiteral(node *BooleanLiteral) Object {
	return &BooleanObject{Value: node.Value}
}

func (i *Interpreter) VisitNilLiteral(node *NilLiteral) Object {
	return &NilObject{}
}

func (i *Interpreter) VisitAssignment(node *Assignment) Object {
	right := node.Expression.Accept(i)

	if !i.isError(right) {
		i.env.Set(node.Identifier.Value, right)
	}

	return right
}

func (i *Interpreter) VisitCommaExpression(node *CommaExpression) Object {
	return nil
}

func (i *Interpreter) VisitTernaryExpression(node *TernaryExpression) Object {
	return nil
}

func (i *Interpreter) VisitReturnStatement(node *ReturnStatement) Object {
	return nil
}

func (i *Interpreter) VisitProgram(node *Program) Object {
	var result Object

	for _, s := range node.Statements {
		result = s.Accept(i)
		if i.isError(result) {
			fmt.Println(result.Inspect())
			return result
		}
	}

	return result
}

func (i *Interpreter) VisitBlockStatement(node *BlockStatement) Object {
	return nil
}

func (i *Interpreter) VisitExpressionStatement(node *ExpressionStatement) Object {
	return node.Expression.Accept(i)
}

func (i *Interpreter) Interpret(node Node) Object {
	return node.Accept(i)
}
