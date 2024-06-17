package main

import "fmt"

type ContextType int

const (
	FunctionContext ContextType = iota
	MainContext
	LoopContext
	ClassMethodContext
	InitializerContext
)

type Context struct {
	Type ContextType
}

const (
	unknownOperatorError    = "unknown operator"
	typeMissMatchError      = "type mismatch"
	divisionByZero          = "divide by zero"
	identifierNotFoundError = "identifier not found"
	methodNotFoundError     = "method not found"
	notFunctionError        = "not a function"
	invalidSyntax           = "invalid syntax"
	notInitialzied          = "variable is not initialized"
	notClassError           = "not a class error"
	notInstanceError        = "not an instance of a class"
	redeclare               = "variable redeclaration"
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
	VisitTernaryExpression(node *TernaryExpression, env *Environment) Object
	VisitBlockStatement(node *BlockStatement, env *Environment) Object
	VisitExpressionStatement(node *ExpressionStatement, env *Environment) Object
	VisitReturnStatement(node *ReturnStatement, env *Environment) Object
	VisitVarStatement(node *VarStatement, env *Environment) Object
	VisitIfStatement(node *IfStatement, env *Environment) Object
	VisitLogical(node *Logical, env *Environment) Object
	VisitWhileStatement(node *While, env *Environment) Object
	VisitForStatement(node *For, env *Environment) Object
	VisitBreakStatement(node *BreakStatement, env *Environment) Object
	VisitContinueStatement(node *ContinueStatement, env *Environment) Object
	VisitCallExpression(node *CallExpression, env *Environment) Object
	VisitFunctionLiteral(node *FunctionLiteral, env *Environment) Object
	VisitFunctionDeclaration(node *FunctionDeclaration, env *Environment) Object
	VisitMethodDeclaration(node *MethodDeclaration, env *Environment) Object
	VisitClassStatement(node *ClassStatement, env *Environment) Object
	VisitGetExpression(node *GetExpression, env *Environment) Object
	VisitSetExpression(node *SetExpression, env *Environment) Object
	VisitThisExpression(node *This, env *Environment) Object
	VisitSuper(node *Super, env *Environment) Object
	VisitArrayLiteral(node *ArrayLiteral, env *Environment) Object
	VisitIndexExpression(node *IndexExpression, env *Environment) Object
	VisitHashLiteral(node *HashLiteral, env *Environment) Object
}

type Interpreter struct {
	contexts []Context
}

func NewInterpreter() *Interpreter {
	contexts := make([]Context, 0)
	interpreter := &Interpreter{contexts}
	interpreter.pushContext(MainContext)
	return interpreter
}

func (i *Interpreter) pushContext(contextType ContextType) {
	i.contexts = append(i.contexts, Context{Type: contextType})
}

func (i *Interpreter) popContext() {
	if len(i.contexts) > 0 {
		i.contexts = i.contexts[:len(i.contexts)-1]
	}
}

func (i *Interpreter) currentContext() *Context {
	if len(i.contexts) > 0 {
		return &i.contexts[len(i.contexts)-1]
	}
	return nil
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

func (i *Interpreter) VisitIndexExpression(node *IndexExpression, env *Environment) Object {
	left := node.Left.Accept(i, env)
	if i.isError(left) {
		return left
	}

	index := node.Index.Accept(i, env)
	if i.isError(index) {
		return index
	}

	return i.evalIndexExpression(left, index)
}

func (i *Interpreter) VisitArrayLiteral(node *ArrayLiteral, env *Environment) Object {
	elements := i.evalExpressions(node.Elements, env)
	if len(elements) == 1 && i.isError(elements[0]) {
		return elements[0]
	}
	return &Array{Elements: elements}
}

func (i *Interpreter) VisitHashLiteral(node *HashLiteral, env *Environment) Object {
	pairs := make(map[HashKey]HashPair)

	for keyNode, valueNode := range node.Pairs {
		key := keyNode.Accept(i, env)
		if i.isError(key) {
			return key
		}

		hashKey, ok := key.(Hashable)
		if !ok {
			return i.newError("%s: %s", invalidSyntax, "unusable hash key")
		}

		value := valueNode.Accept(i, env)
		if i.isError(value) {
			return value
		}

		hashed := hashKey.HashKey()
		pairs[hashed] = HashPair{Key: key, Value: value}
	}

	return &Hash{Pairs: pairs}
}

func (i *Interpreter) VisitSuper(node *Super, env *Environment) Object {
	if i.currentContext().Type != ClassMethodContext && i.currentContext().Type != InitializerContext {
		return i.newError("%s: %s", invalidSyntax, "[super] cannot be used outside of class method")
	}
	if obj, ok := env.Get(node.Token.Lexeme); ok {
		if obj.Type() != ClassObj {
			return i.newError("%s: %s", invalidSyntax, "[super] must be a superclass")
		}
		class := obj.(*ClassObject)
		if method, ok := class.GetSuperMethod(node.Method.Value); ok {
			return method
		}

		return i.newError("%s: %s", methodNotFoundError, node.Method.Value)
	}

	return i.newError("%s: %s", invalidSyntax, "Can't use super in a class with no superclass.")
}

func (i *Interpreter) VisitThisExpression(node *This, env *Environment) Object {
	if i.currentContext().Type != ClassMethodContext && i.currentContext().Type != InitializerContext {
		return i.newError("%s: %s", invalidSyntax, "[this] cannot be used outside of class method")
	}
	if obj, ok := env.Get(node.Token.Lexeme); ok {
		return obj
	}

	return i.newError("%s: %s", identifierNotFoundError, "this")
}

func (i *Interpreter) VisitSetExpression(node *SetExpression, env *Environment) Object {
	object := node.Object.Accept(i, env)

	if object.Type() != InstanceObj {
		return i.newError("%s: %s %s", invalidSyntax, "Only instances have fields.", object.Type())
	}

	instance := object.(*InstanceObject)
	value := node.Value.Accept(i, env)
	instance.SetField(node.Property.Value, value)

	return value
}

func (i *Interpreter) VisitGetExpression(node *GetExpression, env *Environment) Object {
	object := node.Object.Accept(i, env)

	if object.Type() == ClassObj {
		class := object.(*ClassObject)
		propertyName := node.Property.Value

		if staticMethod, ok := class.StaticMethods[propertyName]; ok {
			return staticMethod
		}

		return i.newError("%s: %s", invalidSyntax, fmt.Sprintf("Undefined static property '%s' on class '%s'", propertyName, class.Name))
	}

	if object.Type() != InstanceObj {
		return i.newError("%s: %s %s", invalidSyntax, "Only instance have properties", object.Type())
	}

	instance := object.(*InstanceObject)
	propertyName := node.Property.Value

	if value, ok := instance.GetField(propertyName); ok {
		return value
	}

	if method, ok := instance.GetMethod(propertyName); ok {
		bm := &BoundMethod{Method: method, Receiver: instance}
		if method.IsGetter {
			return i.applyBoundMethod(bm, nil)
		}
		return &BoundMethod{Method: method, Receiver: instance}
	}

	return i.newError("%s: %s", invalidSyntax, fmt.Sprintf("Undefined property '%s' on instance of class '%s'", propertyName, instance.Class.Name))
}

func (i *Interpreter) VisitClassStatement(node *ClassStatement, env *Environment) Object {
	class := &ClassObject{Name: node.Name, Methods: make(map[string]*Function), StaticMethods: make(map[string]*Function)}
	if node.Name == nil {
		return i.newError("%s: %s", invalidSyntax, "missing class name in declaration")
	}

	if node.SuperClass != nil {
		// no hoisting for now
		if obj, ok := env.Get(node.SuperClass.Value); ok {
			if obj.Type() != ClassObj {
				return i.newError("%s: %s", invalidSyntax, "Superclass must be a class.")
			}

			super := obj.(*ClassObject)
			class.SuperClass = super
		} else {
			return i.newError("%s: %s", invalidSyntax, "Superclass doesn't not exist")
		}
	}

	i.pushContext(ClassMethodContext)
	for _, m := range node.Methods {
		method := m.Accept(i, env)
		if method.Type() != FunctionObj {
			return i.newError("%s: %s", invalidSyntax, "Invalid method declaration inside a class")
		}
		fun := method.(*Function)
		if _, ok := class.Methods[m.Name.Value]; ok {
			return i.newError("%s: %s %s", invalidSyntax, "duplicate method name", m.Name.Value)
		}
		if _, ok := class.StaticMethods[m.Name.Value]; ok {
			return i.newError("%s: %s %s", invalidSyntax, "duplicate method name", m.Name.Value)
		}

		if m.IsStatic {
			class.StaticMethods[m.Name.Value] = fun
		} else {
			class.Methods[m.Name.Value] = fun
		}
	}
	i.popContext()
	env.Set(node.Name.Value, class)
	return class
}

func (i *Interpreter) VisitFunctionDeclaration(node *FunctionDeclaration, env *Environment) Object {
	function := &Function{Name: node.Name, Parameters: node.Params, Body: node.Body, Env: env}
	if node.Name == nil {
		return i.newError("%s: %s", invalidSyntax, "missing function name in declaration")
	}
	env.Set(node.Name.Value, function)
	return function
}

func (i *Interpreter) VisitMethodDeclaration(node *MethodDeclaration, env *Environment) Object {
	return &Function{Name: node.Name, Parameters: node.Params, Body: node.Body, Env: env, IsStatic: node.IsStatic, IsGetter: node.IsGetter}
}

func (i *Interpreter) VisitFunctionLiteral(node *FunctionLiteral, env *Environment) Object {
	function := &Function{Parameters: node.Params, Body: node.Body, Env: env}
	if node.Name != nil {
		env.Set(node.Name.Value, function)
	}
	return function
}

func (i *Interpreter) VisitContinueStatement(node *ContinueStatement, env *Environment) Object {
	if i.currentContext().Type != LoopContext {
		return i.newError("%s: %s", invalidSyntax, "Continue statement not within loop")
	}
	return &ContinueSignal{}
}

func (i *Interpreter) VisitBreakStatement(node *BreakStatement, env *Environment) Object {
	if i.currentContext().Type != LoopContext {
		return i.newError("%s: %s", invalidSyntax, "Break statement not within loop")
	}
	return &BreakSignal{}
}

func (i *Interpreter) VisitForStatement(node *For, env *Environment) Object {
	if node.Initializer != nil {
		initResult := node.Initializer.Accept(i, env)
		if i.isError(initResult) {
			return initResult
		}
	}

	for {
		condition := node.Condition.Accept(i, env)
		if i.isError(condition) {
			return condition
		}

		if !i.isTruthy(condition) {
			break
		}

		i.pushContext(LoopContext)
		body := node.Body.Accept(i, env)
		i.popContext()
		if i.isContinue(body) {
			goto increment
		}
		if i.isError(body) || i.isBreak(body) || i.isReturn(body) {
			return body
		}

	increment:
		if node.Increment != nil {
			incrementResult := node.Increment.Accept(i, env)
			if i.isError(incrementResult) {
				return incrementResult
			}
		}
	}

	return nil
}

func (i *Interpreter) VisitWhileStatement(node *While, env *Environment) Object {
	for {
		condition := node.Condition.Accept(i, env)
		if i.isError(condition) {
			return condition
		}

		if !i.isTruthy(condition) {
			break
		}

		i.pushContext(LoopContext)
		body := node.Body.Accept(i, env)
		i.popContext()
		if i.isContinue(body) {
			continue
		}
		if i.isError(body) || i.isBreak(body) || i.isReturn(body) {
			return body
		}
	}

	return nil
}

func (i *Interpreter) VisitIdentifier(node *Identifier, env *Environment) Object {
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

	if _, ok := env.GetCurrentScope(node.Identifier.Value); ok {
		return i.newError("%s: %s", redeclare, node.Identifier.Value)
	}

	if !i.isError(right) {
		// define a variable
		env.Define(node.Identifier.Value, right)
	}

	return right
}

func (i *Interpreter) VisitBinary(node *Binary, env *Environment) Object {
	left := node.Left.Accept(i, env)
	right := node.Right.Accept(i, env)

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

func (i *Interpreter) VisitCallExpression(node *CallExpression, env *Environment) Object {
	var arguments []Object
	callee := node.Callee.Accept(i, env)

	if i.isError(callee) {
		return callee
	}

	for _, argument := range node.Arguments {
		result := argument.Accept(i, env)
		if i.isError(result) {
			return nil
		}
		arguments = append(arguments, result)
	}

	switch callee.Type() {
	case ClassObj:
		return i.instantiateClass(callee, arguments)
	case BoundObj:
		return i.applyBoundMethod(callee, arguments)
	default:
		return i.applyFunction(callee, arguments)
	}
}

func (i *Interpreter) unwrapReturnValue(obj Object) Object {
	if returnValue, ok := obj.(*ReturnValueObject); ok {
		return returnValue.Value
	}
	// implicit return
	return obj
}

func (i *Interpreter) applyBoundMethod(class Object, args []Object) Object {
	bm, ok := class.(*BoundMethod)
	if !ok {
		return i.newError("%s: %s", invalidSyntax, class.Type())
	}

	// Check if the method is the initializer
	if bm.Method.Name.Value == "init" {
		i.pushContext(InitializerContext)
		defer i.popContext()

		extendedEnv := NewEnclosingEnvironment(bm.Method.Env)
		extendedEnv.Set("this", bm.Receiver)
		if bm.Receiver.Class.SuperClass != nil {
			extendedEnv.Set("super", bm.Receiver.Class.SuperClass)
		}

		// Set parameters for the initializer
		for idx, param := range bm.Method.Parameters {
			extendedEnv.Set(param.Value, args[idx])
		}

		// Execute the initializer
		i.executeBlock(bm.Method.Body.Statements, extendedEnv)
		return bm.Receiver
	}

	extendedEnv := NewEnclosingEnvironment(bm.Method.Env)
	extendedEnv.Set("this", bm.Receiver)
	if bm.Receiver.Class.SuperClass != nil {
		extendedEnv.Set("super", bm.Receiver.Class.SuperClass)
	}

	i.pushContext(ClassMethodContext)
	defer func() { i.popContext() }()

	for idx, param := range bm.Method.Parameters {
		extendedEnv.Set(param.Value, args[idx])
	}

	result := i.executeBlock(bm.Method.Body.Statements, extendedEnv)

	// For chaining, return the instance if the method returns null or itself
	if result == nil || result.Type() == NillObj {
		return bm.Receiver
	}

	return result
}

func (i *Interpreter) instantiateClass(class Object, args []Object) Object {
	cl, ok := class.(*ClassObject)
	if !ok {
		return i.newError("%s: %s", invalidSyntax, class.Type())
	}

	instance := &InstanceObject{
		Class:  cl,
		Fields: make(map[string]Object),
	}

	if initMethod, ok := cl.Methods["init"]; ok {
		i.pushContext(InitializerContext)
		defer i.popContext()
		if len(initMethod.Parameters) != len(args) {
			return i.newError("constructor for class %s expected %d arguments, got %d", cl.Name.Value, len(initMethod.Parameters), len(args))
		}

		newEnv := NewEnclosingEnvironment(initMethod.Env)
		newEnv.Set("this", instance)
		if instance.Class.SuperClass != nil {
			newEnv.Set("super", instance.Class.SuperClass)
		}
		for idx, param := range initMethod.Parameters {
			newEnv.Set(param.Value, args[idx])
		}

		result := i.executeBlock(initMethod.Body.Statements, newEnv)
		if i.isError(result) {
			return result
		}
	}

	return instance
}

func (i *Interpreter) applyFunction(fn Object, args []Object) Object {

	switch fn := fn.(type) {
	case *Function:
		if len(fn.Parameters) != len(args) {
			return i.newError("%s: Expected %d arguments, but got %d .", invalidSyntax, len(fn.Parameters), len(args))
		}
		extendedEnv := i.extendedFunctionEnv(fn, args)
		i.pushContext(FunctionContext)
		defer func() { i.popContext() }()
		evaluated := fn.Body.Accept(i, extendedEnv)

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

func (i *Interpreter) VisitTernaryExpression(node *TernaryExpression, env *Environment) Object {
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

func (i *Interpreter) VisitReturnStatement(node *ReturnStatement, env *Environment) Object {
	var value Object

	if i.currentContext().Type != FunctionContext && i.currentContext().Type != InitializerContext && i.currentContext().Type != ClassMethodContext {
		return i.newError("%s: %s", invalidSyntax, "Cannot use 'return' outside of function")
	}

	if node.ReturnValue == nil {
		value = Null
	} else {
		value = node.ReturnValue.Accept(i, env)
		if i.currentContext().Type == InitializerContext {
			return i.newError("%s: %s", invalidSyntax, "Cannot use 'return' inside init method")
		}
	}

	return &ReturnValueObject{Value: value}
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
		if i.isBreak(result) || i.isReturn(result) || i.isError(result) || i.isContinue(result) {
			return result
		}
	}

	return nil
}

func (i *Interpreter) evalExpressions(exps []Expression, env *Environment) []Object {
	var result []Object

	for _, e := range exps {
		expr := e.Accept(i, env)
		if i.isError(expr) {
			return []Object{expr}
		}
		result = append(result, expr)
	}

	return result
}

func (i *Interpreter) evalIndexExpression(left, index Object) Object {
	switch {
	case left.Type() == ArrayObj && index.Type() == FloatObj:
		return i.evalArrayIndexExpression(left, index)
	case left.Type() == HashObj:
		return i.evalHashExpression(left, index)
	default:
		return i.newError("%s: %s", invalidSyntax, "index operator not supported")

	}
}

func (i *Interpreter) evalHashExpression(hash, index Object) Object {
	hashObject := hash.(*Hash)

	key, ok := index.(Hashable)
	if !ok {
		return i.newError("%s: %s", invalidSyntax, fmt.Sprintf("unusable hash key %s", index.Type()))
	}

	pair, ok := hashObject.Pairs[key.HashKey()]
	if !ok {
		return Null
	}

	return pair.Value
}

func (i *Interpreter) evalArrayIndexExpression(array, index Object) Object {
	arrayObj := array.(*Array)
	idx := int(index.(*FloatObject).Value)
	max := len(arrayObj.Elements) - 1

	if idx < 0 || idx > max {
		return Null
	}

	return arrayObj.Elements[idx]
}

func (i *Interpreter) VisitExpressionStatement(node *ExpressionStatement, env *Environment) Object {
	return node.Expression.Accept(i, env)
}

func (i *Interpreter) Interpret(node Node, env *Environment) Object {
	return node.Accept(i, env)
}
