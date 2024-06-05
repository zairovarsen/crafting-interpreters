package main

import "fmt"

const (
	unknownOperatorError    = "unknown operator"
	typeMissMatchError      = "type mismatch"
	divisionByZero          = "divide by zero"
	identifierNotFoundError = "identifier not found"
	notFunctionError        = "not a function"
	invalidSyntax           = "invalid syntax"
	notInitialzied          = "variable is not initialized"
)

var (
	Null  = &NilObject{}
	True  = &BooleanObject{Value: true}
	False = &BooleanObject{Value: false}
)

// java way is terrible but whatever :)
type Visitor interface {
	VisitProgram(node *Program, env *Environment) Object
	VisitIdentifier(node *Identifier, env *Environment) Object
	VisitBinary(node *Binary, env *Environment) Object
	VisitUnary(node *Unary, env *Environment) Object
	VisitStringLiteral(node *StringLiteral, env *Environment) Object
	VisitNumberLiteral(node *NumberLiteral, env *Environment) Object
	VisitBooleanLiteral(node *BooleanLiteral, env *Environment) Object
	VisitNilLiteral(node *NilLiteral, env *Environment) Object
	VisitGroupedExpression(node *GroupedExpression, env *Environment) Object
	VisitAssignment(node *Assignment, env *Environment) Object
	VisitCommaExpression(node *CommaExpression, env *Environment) Object
	VisitTernaryExpression(node *TernaryExpression, env *Environment) Object
	VisitBlockStatement(node *BlockStatement, env *Environment) Object
	VisitExpressionStatement(node *ExpressionStatement, env *Environment) Object
	VisitReturnStatement(node *ReturnStatement, env *Environment) Object
	VisitPrintStatement(node *PrintStatement, env *Environment) Object
	VisitVarStatement(node *VarStatement, env *Environment) Object
	VisitIfStatement(node *IfStatement, env *Environment) Object
	VisitLogical(node *Logical, env *Environment) Object
}

type Interpreter struct{}

func NewInterpreter() *Interpreter {
	return &Interpreter{}
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

func (i *Interpreter) VisitPrintStatement(node *PrintStatement, env *Environment) Object {
	value := node.PrintValue.Accept(i, env)
	fmt.Println(value.Inspect())
	return nil
}

func (i *Interpreter) VisitIdentifier(node *Identifier, env *Environment) Object {
	obj, ok := env.Get(node.Value)
	if !ok {
		return i.newError("%s: %s", identifierNotFoundError, node.Value)
	}
	if obj.Type() == NillObj {
		return i.newError("%s: %s", notInitialzied, node.TokenLiteral())
	}
	return obj
}

func (i *Interpreter) VisitLogical(node *Logical, env *Environment) Object {
	left := node.Left.Accept(i, env)

	if node.Token.Type == OR {
		if i.isTruthy(left) {
			return left
		}
	} else {
		if !i.isTruthy(left) {
			return left
		}
	}

	return node.Right.Accept(i, env)
}

func (i *Interpreter) VisitGroupedExpression(node *GroupedExpression, env *Environment) Object {
	return node.Expression.Accept(i, env)
}

func (i *Interpreter) VisitIfStatement(node *IfStatement, env *Environment) Object {
	condition := node.Condition.Accept(i, env)
	if i.isError(condition) {
		return condition
	}

	if i.isTruthy(condition) {
		return node.ThenBranch.Accept(i, env)
	} else if node.ElseBranch != nil {
		return node.ElseBranch.Accept(i, env)
	} else {
		return Null
	}
}

func (i *Interpreter) VisitVarStatement(node *VarStatement, env *Environment) Object {
	right := node.Expression.Accept(i, env)

	if !i.isError(right) {
		// define a variable
		env.Set(node.Identifier.Value, right)
	}

	return right
}

func (i *Interpreter) VisitBinary(node *Binary, env *Environment) Object {
	left := node.Left.Accept(i, env)
	right := node.Right.Accept(i, env)

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

func (i *Interpreter) VisitUnary(node *Unary, env *Environment) Object {
	right := node.Right.Accept(i, env)

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

func (i *Interpreter) VisitStringLiteral(node *StringLiteral, env *Environment) Object {
	return &StringObject{Value: node.Value}
}

func (i *Interpreter) VisitNumberLiteral(node *NumberLiteral, env *Environment) Object {
	return &FloatObject{Value: node.Value}
}

func (i *Interpreter) VisitBooleanLiteral(node *BooleanLiteral, env *Environment) Object {
	return &BooleanObject{Value: node.Value}
}

func (i *Interpreter) VisitNilLiteral(node *NilLiteral, env *Environment) Object {
	return &NilObject{}
}

func (i *Interpreter) VisitAssignment(node *Assignment, env *Environment) Object {
	right := node.Expression.Accept(i, env)

	if !i.isError(right) {
		env.Set(node.Identifier.Value, right)
	}

	return right
}

func (i *Interpreter) VisitCommaExpression(node *CommaExpression, env *Environment) Object {
	return nil
}

func (i *Interpreter) VisitTernaryExpression(node *TernaryExpression, env *Environment) Object {
	return nil
}

func (i *Interpreter) VisitReturnStatement(node *ReturnStatement, env *Environment) Object {
	return nil
}

func (i *Interpreter) VisitProgram(node *Program, env *Environment) Object {
	var result Object

	for _, s := range node.Statements {
		result = s.Accept(i, env)
		if i.isError(result) {
			return result
		}
	}

	return result
}

func (i *Interpreter) VisitBlockStatement(node *BlockStatement, env *Environment) Object {
	newEnv := NewEnclosingEnvironment(env)
	return i.executeBlock(node.Statements, newEnv)
}

func (i *Interpreter) executeBlock(statements []Statement, env *Environment) Object {
	var result Object

	for _, stmt := range statements {
		result = stmt.Accept(i, env)
	}

	return result
}

func (i *Interpreter) VisitExpressionStatement(node *ExpressionStatement, env *Environment) Object {
	return node.Expression.Accept(i, env)
}

func (i *Interpreter) Interpret(node Node, env *Environment) Object {
	return node.Accept(i, env)
}
