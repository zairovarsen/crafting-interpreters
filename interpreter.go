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

var builtins = map[string]*Builtin{
	BuiltinFuncNamePrint: GetBuiltinByName(BuiltinFuncNamePrint),
}

// java way is terrible but whatever :)
type Visitor interface {
	VisitProgram(node *Program, env *Environment, parent Statement) Object
	VisitIdentifier(node *Identifier, env *Environment, parent Statement) Object
	VisitBinary(node *Binary, env *Environment, parent Statement) Object
	VisitUnary(node *Unary, env *Environment, parent Statement) Object
	VisitStringLiteral(node *StringLiteral, env *Environment, parent Statement) Object
	VisitNumberLiteral(node *NumberLiteral, env *Environment, parent Statement) Object
	VisitBooleanLiteral(node *BooleanLiteral, env *Environment, parent Statement) Object
	VisitNilLiteral(node *NilLiteral, env *Environment, parent Statement) Object
	VisitGroupedExpression(node *GroupedExpression, env *Environment, parent Statement) Object
	VisitAssignment(node *Assignment, env *Environment, parent Statement) Object
	VisitTernaryExpression(node *TernaryExpression, env *Environment, parent Statement) Object
	VisitBlockStatement(node *BlockStatement, env *Environment, parent Statement) Object
	VisitExpressionStatement(node *ExpressionStatement, env *Environment, parent Statement) Object
	VisitReturnStatement(node *ReturnStatement, env *Environment, parent Statement) Object
	VisitVarStatement(node *VarStatement, env *Environment, parent Statement) Object
	VisitIfStatement(node *IfStatement, env *Environment, parent Statement) Object
	VisitLogical(node *Logical, env *Environment, parent Statement) Object
	VisitWhileStatement(node *While, env *Environment, parent Statement) Object
	VisitForStatement(node *For, env *Environment, parent Statement) Object
	VisitBreakStatement(node *BreakStatement, env *Environment, parent Statement) Object
	VisitContinueStatement(node *ContinueStatement, env *Environment, parent Statement) Object
	VisitCallExpression(node *CallExpression, env *Environment, parent Statement) Object
	VisitFunctionLiteral(node *FunctionLiteral, env *Environment, parent Statement) Object
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

func (i *Interpreter) isContinue(obj Object) bool {
	if obj != nil {
		return obj.Type() == ContinueObj
	}
	return false
}

func (i *Interpreter) isBreak(obj Object) bool {
	if obj != nil {
		return obj.Type() == BreakObj
	}
	return false
}

func (i *Interpreter) isReturn(obj Object) bool {
	if obj != nil {
		return obj.Type() == ReturnValueObj
	}
	return false
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

func (i *Interpreter) VisitFunctionLiteral(node *FunctionLiteral, env *Environment, parent Statement) Object {
	function := &Function{Parameters: node.Params, Body: node.Body, Env: env}
	env.Set(node.Token.Lexeme, function)
	return nil
}

func (i *Interpreter) VisitContinueStatement(node *ContinueStatement, env *Environment, parent Statement) Object {
	switch parent.(type) {
	case *For, *While:
		return &ContinueSignal{}
	default:
		return i.newError("%s: %s", invalidSyntax, "Continue statement not within loop")
	}
}

func (i *Interpreter) VisitBreakStatement(node *BreakStatement, env *Environment, parent Statement) Object {
	switch parent.(type) {
	case *For, *While:
		return &BreakSignal{}
	default:
		return i.newError("%s: %s", invalidSyntax, "Break statement not within loop")
	}
}

func (i *Interpreter) VisitForStatement(node *For, env *Environment, parent Statement) Object {
	if node.Initializer != nil {
		initResult := node.Initializer.Accept(i, env, node)
		if i.isError(initResult) {
			return initResult
		}
	}

	for {
		condition := node.Condition.Accept(i, env, node)
		if i.isError(condition) {
			return condition
		}

		if !i.isTruthy(condition) {
			break
		}

		body := node.Body.Accept(i, env, node)
		if i.isContinue(body) {
			goto increment
		}
		if i.isError(body) || i.isBreak(body) || i.isReturn(body) {
			return body
		}

	increment:
		if node.Increment != nil {
			incrementResult := node.Increment.Accept(i, env, node)
			if i.isError(incrementResult) {
				return incrementResult
			}
		}
	}

	return nil
}

func (i *Interpreter) VisitWhileStatement(node *While, env *Environment, parent Statement) Object {
	for {
		condition := node.Condition.Accept(i, env, node)
		if i.isError(condition) {
			return condition
		}

		if !i.isTruthy(condition) {
			break
		}

		body := node.Body.Accept(i, env, node)
		if i.isContinue(body) {
			continue
		}
		if i.isError(body) || i.isBreak(body) || i.isReturn(body) {
			return body
		}
	}

	return nil
}

func (i *Interpreter) VisitIdentifier(node *Identifier, env *Environment, parent Statement) Object {
	if obj, ok := env.Get(node.Value); ok {
		if obj.Type() == NillObj {
			return i.newError("%s: %s", notInitialzied, node.TokenLiteral())
		}
		return obj
	}

	if builtin, ok := builtins[node.Value]; ok {
		return builtin
	}

	return i.newError("%s: %s", identifierNotFoundError, node.Value)
}

func (i *Interpreter) VisitLogical(node *Logical, env *Environment, parent Statement) Object {
	left := node.Left.Accept(i, env, parent)

	if node.Token.Type == OR {
		if i.isTruthy(left) {
			return left
		}
	} else {
		if !i.isTruthy(left) {
			return left
		}
	}

	return node.Right.Accept(i, env, parent)
}

func (i *Interpreter) VisitGroupedExpression(node *GroupedExpression, env *Environment, parent Statement) Object {
	return node.Expression.Accept(i, env, parent)
}

func (i *Interpreter) VisitIfStatement(node *IfStatement, env *Environment, parent Statement) Object {
	condition := node.Condition.Accept(i, env, parent)
	if i.isError(condition) {
		return condition
	}

	if i.isTruthy(condition) {
		return node.ThenBranch.Accept(i, env, parent)
	} else if node.ElseBranch != nil {
		return node.ElseBranch.Accept(i, env, parent)
	} else {
		return Null
	}
}

func (i *Interpreter) VisitVarStatement(node *VarStatement, env *Environment, parent Statement) Object {
	right := node.Expression.Accept(i, env, parent)

	if !i.isError(right) {
		// define a variable
		env.Define(node.Identifier.Value, right)
	}

	return right
}

func (i *Interpreter) VisitBinary(node *Binary, env *Environment, parent Statement) Object {
	left := node.Left.Accept(i, env, parent)
	right := node.Right.Accept(i, env, parent)

	switch {
	case left.Type() == ErrorObj:
		return left
	case right.Type() == ErrorObj:
		return right
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

func (i *Interpreter) VisitUnary(node *Unary, env *Environment, parent Statement) Object {
	right := node.Right.Accept(i, env, parent)

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

func (i *Interpreter) VisitStringLiteral(node *StringLiteral, env *Environment, parent Statement) Object {
	return &StringObject{Value: node.Value}
}

func (i *Interpreter) VisitNumberLiteral(node *NumberLiteral, env *Environment, parent Statement) Object {
	return &FloatObject{Value: node.Value}
}

func (i *Interpreter) VisitBooleanLiteral(node *BooleanLiteral, env *Environment, parent Statement) Object {
	return &BooleanObject{Value: node.Value}
}

func (i *Interpreter) VisitNilLiteral(node *NilLiteral, env *Environment, parent Statement) Object {
	return &NilObject{}
}

func (i *Interpreter) VisitAssignment(node *Assignment, env *Environment, parent Statement) Object {
	right := node.Expression.Accept(i, env, parent)

	if !i.isError(right) {
		env.Set(node.Identifier.Value, right)
	}

	return right
}

func (i *Interpreter) VisitCallExpression(node *CallExpression, env *Environment, parent Statement) Object {
	var arguments []Object
	callee := node.Callee.Accept(i, env, parent)

	if i.isError(callee) {
		return i.newError("%s: %s", notFunctionError, callee.Inspect())
	}

	for _, argument := range node.Arguments {
		result := argument.Accept(i, env, parent)
		if i.isError(result) {
			return nil
		}
		arguments = append(arguments, result)
	}

	return i.applyFunction(callee, arguments, parent)
}

func (i *Interpreter) unwrapReturnValue(obj Object) Object {
	if returnValue, ok := obj.(*ReturnValueObject); ok {
		return returnValue.Value
	}

	return obj
}

func (i *Interpreter) applyFunction(fn Object, args []Object, parent Statement) Object {
	switch fn := fn.(type) {
	case *Function:
		if len(fn.Parameters) != len(args) {
			return i.newError("%s: Expected %d arguments, but got %d .", invalidSyntax, len(fn.Parameters), len(args))
		}
		extendedEnv := i.extendedFunctionEnv(fn, args)
		evaluated := fn.Body.Accept(i, extendedEnv, parent)
		return i.unwrapReturnValue(evaluated)
	case *Builtin:
		if result := fn.Fn(args...); result != nil {
			return result
		} else {
			return Null
		}
	default:
		return i.newError("%s: %s", notFunctionError, fn.Type())
	}
}

func (i *Interpreter) extendedFunctionEnv(fn *Function, args []Object) *Environment {
	env := NewEnclosingEnvironment(fn.Env)

	for paramIdx, param := range fn.Parameters {
		env.Set(param.Value, args[paramIdx])
	}

	return env
}

func (i *Interpreter) VisitTernaryExpression(node *TernaryExpression, env *Environment, parent Statement) Object {
	return nil
}

func (i *Interpreter) VisitReturnStatement(node *ReturnStatement, env *Environment, parent Statement) Object {
	value := node.ReturnValue.Accept(i, env, parent)
	return &ReturnValueObject{Value: value}
}

func (i *Interpreter) VisitProgram(node *Program, env *Environment, parent Statement) Object {
	var result Object

	for _, s := range node.Statements {
		result = s.Accept(i, env, node)
		if i.isError(result) {
			return result
		}
	}

	return result
}

func (i *Interpreter) VisitBlockStatement(node *BlockStatement, env *Environment, parent Statement) Object {
	newEnv := NewEnclosingEnvironment(env)
	return i.executeBlock(node.Statements, newEnv, parent)
}

func (i *Interpreter) executeBlock(statements []Statement, env *Environment, parent Statement) Object {
	var result Object

	for _, stmt := range statements {
		result = stmt.Accept(i, env, parent)
		if i.isBreak(result) || i.isReturn(result) || i.isError(result) || i.isContinue(result) {
			return result
		}
	}

	return nil
}

func (i *Interpreter) VisitExpressionStatement(node *ExpressionStatement, env *Environment, parent Statement) Object {
	return node.Expression.Accept(i, env, parent)
}

func (i *Interpreter) Interpret(node Node, env *Environment, parent Statement) Object {
	return node.Accept(i, env, parent)
}
