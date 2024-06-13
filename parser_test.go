package main

import (
	"fmt"
	"testing"
)

func createParseProgram(input string) *Program {
	scanner := NewScanner([]byte(input))
	scanner.scanTokens()

	fmt.Println(scanner.Tokens)
	parser := NewParser(scanner.Tokens)
	program := parser.parse()

	return program
}

func TestSetExpression(t *testing.T) {
	input := `breakfast.omelette.filling.meat = ham;`

	program := createParseProgram(input)

	if len(program.Statements) != 1 {
		t.Fatalf("program.Statements does not contain %d statements. got=%d\n", 1, len(program.Statements))
	}

	expr, ok := program.Statements[0].(*ExpressionStatement)
	if !ok {
		t.Fatalf("Expected %T, got=%T", &ExpressionStatement{}, program.Statements[0])
	}

	set, ok := expr.Expression.(*SetExpression)
	if !ok {
		t.Fatalf("Expected %T, got=%T", &Assignment{}, expr.Expression)
	}

	if !testLiteral(t, set.Value, "ham") {
		return
	}

	getExpr, ok := set.Object.(*GetExpression)
	if !ok {
		t.Fatalf("Expected %T, got=%T", &GetExpression{}, expr.Expression)
	}

	if !testLiteral(t, set.Property, "meat") {
		return
	}

	if !testLiteral(t, getExpr.Property, "filling") {
		return
	}

	getExpr, ok = getExpr.Object.(*GetExpression)
	if !testLiteral(t, getExpr.Property, "omelette") {
		return
	}

	if !testLiteral(t, getExpr.Object, "breakfast") {
		return
	}
}

func TestGetExpression(t *testing.T) {
	input := `breakfast.omelette.filling.meat;`

	program := createParseProgram(input)

	if len(program.Statements) != 1 {
		t.Fatalf("program.Statements does not contain %d statements. got=%d\n", 1, len(program.Statements))
	}

	expr, ok := program.Statements[0].(*ExpressionStatement)
	if !ok {
		t.Fatalf("Expected %T, got=%T", &ExpressionStatement{}, program.Statements[0])
	}

	getExpr, ok := expr.Expression.(*GetExpression)
	if !ok {
		t.Fatalf("Expected %T, got=%T", &GetExpression{}, expr.Expression)
	}

	if !testLiteral(t, getExpr.Property, "meat") {
		return
	}

	getExpr, ok = getExpr.Object.(*GetExpression)
	if !ok {
		t.Fatalf("Expected %T, got=%T", &GetExpression{}, expr.Expression)
	}
	if !testLiteral(t, getExpr.Property, "filling") {
		return
	}

	getExpr, ok = getExpr.Object.(*GetExpression)
	if !ok {
		t.Fatalf("Expected %T, got=%T", &GetExpression{}, expr.Expression)
	}
	if !testLiteral(t, getExpr.Property, "omelette") {
		return
	}
	if !testLiteral(t, getExpr.Object, "breakfast") {
		return
	}
}

func TestClassStatement(t *testing.T) {
	input := `class Breakfast {
		cook() {
			print("Eggs");
		}

		serve(who) {
			print("Enjoy your breakfast, " + who + ".");
		}
	}`

	program := createParseProgram(input)

	if len(program.Statements) != 1 {
		t.Fatalf("program.Statements does not contain %d statements. got=%d\n", 1, len(program.Statements))
	}

	class, ok := program.Statements[0].(*ClassStatement)
	if !ok {
		t.Fatalf("Expected %T, got=%T", &ClassStatement{}, program.Statements[0])
	}

	if len(class.Methods) != 2 {
		t.Errorf("Expected %d methods in class, got=%d\n", 2, len(class.Methods))
	}
}

func TestTernary(t *testing.T) {
	input := `var a = true ? 5 : "hello world";`

	program := createParseProgram(input)

	if len(program.Statements) != 1 {
		t.Fatalf("program.Statements does not contain %d statements. got=%d\n", 1, len(program.Statements))
	}

	varStatement, ok := program.Statements[0].(*VarStatement)
	if !ok {
		t.Fatalf("Expected %T, got=%T", &VarStatement{}, program.Statements[0])
	}

	expr, ok := varStatement.Expression.(*TernaryExpression)
	if !ok {
		t.Fatalf("Expected %T, got=%T", &TernaryExpression{}, varStatement.Expression)
	}

	if !testLiteral(t, expr.Condition, true) {
		return
	}

	if !testLiteral(t, expr.ThenBranch, 5) {
		return
	}

	if !testLiteral(t, expr.ElseBranch, "hello world") {
		return
	}
}

func TestFunctionLiteral(t *testing.T) {
	input := `add(function(a) { print(a); });`

	program := createParseProgram(input)

	if len(program.Statements) != 1 {
		t.Fatalf("program.Statements does not contain %d statements. got=%d\n", 1, len(program.Statements))
	}

	expr, ok := program.Statements[0].(*ExpressionStatement)
	if !ok {
		t.Fatalf("Expected %T, got=%T", &ExpressionStatement{}, program.Statements[0])
	}

	call, ok := expr.Expression.(*CallExpression)
	if !ok {
		t.Fatalf("Expected %T, got=%T", &CallExpression{}, program.Statements[0])
	}

	fun, ok := call.Arguments[0].(*FunctionLiteral)
	if !ok {
		t.Fatalf("Expected %T, got=%T", &FunctionLiteral{}, call.Arguments[0])
	}

	if testLiteral(t, fun.Params[0], "a") {
		return
	}
}

func TestFuncDeclaration(t *testing.T) {
	input := `function add(a,b,c) { a + b; }`

	program := createParseProgram(input)

	if len(program.Statements) != 1 {
		t.Fatalf("program.Statements does not contain %d statements. got=%d\n", 1, len(program.Statements))
	}

	fun, ok := program.Statements[0].(*FunctionDeclaration)
	if !ok {
		t.Fatalf("Expected %T, got=%T", &FunctionLiteral{}, program.Statements[0])
	}

	if len(fun.Params) != 3 {
		t.Errorf("Expected length of params to be %d, got=%d", 3, len(fun.Params))
	}

	expr, ok := fun.Body.Statements[0].(*ExpressionStatement)
	if !ok {
		t.Errorf("Expected %T, got=%T", &ExpressionStatement{}, fun.Body.Statements[0])
	}

	if !testBinaryExpression(t, expr.Expression, "a", "+", "b") {
		return
	}
}

func TestCallExpression(t *testing.T) {
	input := `add(1, 2 * 3,4 + 5);`

	program := createParseProgram(input)

	if len(program.Statements) != 1 {
		t.Fatalf("program.Statements does not contain %d statements. got=%d\n", 1, len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ExpressionStatement)
	if !ok {
		t.Fatalf("Expected %T, got=%T", &ExpressionStatement{}, program.Statements[0])
	}

	exp, ok := stmt.Expression.(*CallExpression)
	if !ok {
		t.Fatalf("Expression is not %T. got=%T", &CallExpression{}, stmt.Expression)
	}

	if !testLiteral(t, exp.Callee, "add") {
		return
	}

	if len(exp.Arguments) != 3 {
		t.Fatalf("wrong length of arguments. got=%d", len(exp.Arguments))
	}

	if !testLiteral(t, exp.Arguments[0], 1) {
		return
	}
	if !testBinaryExpression(t, exp.Arguments[1], 2, "*", 3) {
		return
	}

	if !testBinaryExpression(t, exp.Arguments[2], 4, "+", 5) {
		return
	}
}

func TestFor(t *testing.T) {
	input := `for (var x = 5; x > 0; x = x - 1) { "hello world"; }`

	program := createParseProgram(input)

	if len(program.Statements) != 1 {
		t.Fatalf("Expected length of statements to be %d, got=%d", 1, len(program.Statements))
	}

	forStmt, ok := program.Statements[0].(*For)
	if !ok {
		t.Fatalf("Expected %T, got=%T", &For{}, program.Statements[0])
	}

	if !testVarStatement(t, forStmt.Initializer, "x", 5) {
		return
	}

	if !testBinaryExpression(t, forStmt.Condition, "x", ">", 0) {
		return
	}

	assignment, ok := forStmt.Increment.(*Assignment)
	if !ok {
		t.Fatalf("Expected=%T, got=%T", &Identifier{}, forStmt.Increment)
	}

	if assignment.Identifier.Value != "x" {
		t.Errorf("Expected assignment value %s, got=%s", "a", assignment.Identifier.Value)
	}

	if len(forStmt.Body.Statements) != 1 {
		t.Errorf("Expected length of statement to be %d, got=%d", 1, len(forStmt.Body.Statements))
	}

	expr, ok := forStmt.Body.Statements[0].(*ExpressionStatement)
	if !ok {
		t.Errorf("Expected %T, got %T", &ExpressionStatement{}, forStmt.Body.Statements[0])
	}

	if !testLiteral(t, expr.Expression, "hello world") {
		return
	}
}

func TestContinueStatement(t *testing.T) {
	input := `while (5 > 2) { continue; }`

	program := createParseProgram(input)

	if len(program.Statements) != 1 {
		t.Fatalf("Expected length of statements to be %d, got=%d", 1, len(program.Statements))
	}

	whileStmt, ok := program.Statements[0].(*While)
	if !ok {
		t.Fatalf("Expected %T, got=%T", &While{}, program.Statements[0])
	}

	breakStatement, ok := whileStmt.Body.Statements[0].(*ContinueStatement)
	if !ok {
		t.Errorf("Expected %T, got=%T", &BreakStatement{}, whileStmt.Body.Statements[0])
	}

	if breakStatement.TokenLiteral() != "continue" {
		t.Errorf("Expected TokenLiteral to be %s, got=%s", "break", breakStatement.TokenLiteral())
	}
}

func TestBreakStatement(t *testing.T) {
	input := `while (5 > 2) { break; }`

	program := createParseProgram(input)

	if len(program.Statements) != 1 {
		t.Fatalf("Expected length of statements to be %d, got=%d", 1, len(program.Statements))
	}

	whileStmt, ok := program.Statements[0].(*While)
	if !ok {
		t.Fatalf("Expected %T, got=%T", &While{}, program.Statements[0])
	}

	breakStatement, ok := whileStmt.Body.Statements[0].(*BreakStatement)
	if !ok {
		t.Errorf("Expected %T, got=%T", &BreakStatement{}, whileStmt.Body.Statements[0])
	}

	if breakStatement.TokenLiteral() != "break" {
		t.Errorf("Expected TokenLiteral to be %s, got=%s", "break", breakStatement.TokenLiteral())
	}
}

func TestWhile(t *testing.T) {
	input := `while (5 > 2) { "hello world"; }`

	program := createParseProgram(input)

	if len(program.Statements) != 1 {
		t.Fatalf("Expected length of statements to be %d, got=%d", 1, len(program.Statements))
	}

	whileStmt, ok := program.Statements[0].(*While)
	if !ok {
		t.Fatalf("Expected %T, got=%T", &While{}, program.Statements[0])
	}

	condition, ok := whileStmt.Condition.(*Binary)
	if !ok {
		t.Errorf("Expected %T, got %T", &Binary{}, whileStmt.Condition)
	}

	if condition.Token.Type != GREATER {
		t.Errorf("Expected operator %s, got=%s", GREATER, condition.Token.Type)
	}

	if !testLiteral(t, condition.Left, 5) {
		return
	}

	if !testLiteral(t, condition.Right, 2) {
		return
	}

	if len(whileStmt.Body.Statements) != 1 {
		t.Errorf("Expected length of statement to be %d, got=%d", 1, len(whileStmt.Body.Statements))
	}

	expr, ok := whileStmt.Body.Statements[0].(*ExpressionStatement)
	if !ok {
		t.Errorf("Expected %T, got %T", &ExpressionStatement{}, whileStmt.Body.Statements[0])
	}

	if !testLiteral(t, expr.Expression, "hello world") {
		return
	}
}

func TestLogicalAnd(t *testing.T) {
	tests := []struct {
		code     string
		left     interface{}
		operator string
		right    interface{}
	}{
		{"false and true;", false, "and", true},
		{"true and 5;", true, "and", 5},
		{`"hello" and "world";`, "hello", "and", "world"},
	}

	for _, test := range tests {
		program := createParseProgram(test.code)

		if len(program.Statements) != 1 {
			t.Fatalf("Expected length of statements to be %d, got=%d", 1, len(program.Statements))
		}

		expr, ok := program.Statements[0].(*ExpressionStatement)
		if !ok {
			t.Fatalf("Expected %T, got=%T", &ExpressionStatement{}, program.Statements[0])
		}

		logicaland, ok := expr.Expression.(*Logical)

		if logicaland.Operator != test.operator {
			t.Errorf("Expected operator %s, got=%s", test.operator, logicaland.Operator)
		}

		if !testLiteral(t, logicaland.Left, test.left) {
			return
		}

		if !testLiteral(t, logicaland.Right, test.right) {
			return
		}
	}
}

func TestLogicalOr(t *testing.T) {
	tests := []struct {
		code     string
		left     interface{}
		operator string
		right    interface{}
	}{
		{"false or true;", false, "or", true},
		{"true or 5;", true, "or", 5},
		{`"hello" or "world";`, "hello", "or", "world"},
	}

	for _, test := range tests {
		program := createParseProgram(test.code)

		if len(program.Statements) != 1 {
			t.Fatalf("Expected length of statements to be %d, got=%d", 1, len(program.Statements))
		}

		expr, ok := program.Statements[0].(*ExpressionStatement)
		if !ok {
			t.Fatalf("Expected %T, got=%T", &ExpressionStatement{}, program.Statements[0])
		}

		logicaland, ok := expr.Expression.(*Logical)

		if logicaland.Operator != test.operator {
			t.Errorf("Expected operator %s, got=%s", test.operator, logicaland.Operator)
		}

		if !testLiteral(t, logicaland.Left, test.left) {
			return
		}

		if !testLiteral(t, logicaland.Right, test.right) {
			return
		}

	}
}

func TestIfStatement(t *testing.T) {
	input := `if (x < y) { x; }`

	program := createParseProgram(input)

	if len(program.Statements) != 1 {
		t.Fatalf("Expected length of statements to be %d, got=%d", 1, len(program.Statements))
	}

	ifstmt, ok := program.Statements[0].(*IfStatement)
	if !ok {
		t.Fatalf("Expected %T, got=%T", &IfStatement{}, program.Statements[0])
	}

	if !testBinaryExpression(t, ifstmt.Condition, "x", "<", "y") {
		return
	}

	block, ok := ifstmt.ThenBranch.(*BlockStatement)
	if !ok {
		t.Errorf("Expected %T, got=%T", &BlockStatement{}, ifstmt.ThenBranch)
	}

	if len(block.Statements) != 1 {
		t.Errorf("Expected length of then to be %d, got=%d", 1, len(block.Statements))
	}
}

func TestIfElseStatement(t *testing.T) {
	input := `if (x < y) { x; } else { y; }`

	program := createParseProgram(input)

	if len(program.Statements) != 1 {
		t.Fatalf("Expected length of statements to be %d, got=%d", 1, len(program.Statements))
	}

	ifstmt, ok := program.Statements[0].(*IfStatement)
	if !ok {
		t.Fatalf("Expected %T, got=%T", &IfStatement{}, program.Statements[0])
	}

	if !testBinaryExpression(t, ifstmt.Condition, "x", "<", "y") {
		return
	}

	then, ok := ifstmt.ThenBranch.(*BlockStatement)
	if !ok {
		t.Errorf("Expected %T, got=%T", &BlockStatement{}, ifstmt.ThenBranch)
	}

	if len(then.Statements) != 1 {
		t.Errorf("Expected length of then to be %d, got=%d", 1, len(then.Statements))
	}

	thenExpr, ok := then.Statements[0].(*ExpressionStatement)

	if !testLiteral(t, thenExpr.Expression, "x") {
		return
	}

	alternative, ok := ifstmt.ElseBranch.(*BlockStatement)
	if !ok {
		t.Errorf("Expected %T, got=%T", &BlockStatement{}, ifstmt.ThenBranch)
	}

	if len(alternative.Statements) != 1 {
		t.Errorf("Expected length of alternative to be %d, got=%d", 1, len(alternative.Statements))
	}

	alternativeExpr, ok := alternative.Statements[0].(*ExpressionStatement)

	if !testLiteral(t, alternativeExpr.Expression, "y") {
		return
	}
}

func TestBlockStatement(t *testing.T) {
	input := `
		{
			var a = 10;
			var b = 20;
			c = a + b;
		}
	`

	program := createParseProgram(input)

	if len(program.Statements) != 1 {
		t.Fatalf("Expected length of statements to be %d, got=%d", 3, len(program.Statements))
	}

	block, ok := program.Statements[0].(*BlockStatement)
	if !ok {
		t.Errorf("Expected=%T, got=%T", &BlockStatement{}, program.Statements[0])
	}

	if len(block.Statements) != 3 {
		t.Errorf("Expected length of block statements to be %d, got=%d", 3, len(program.Statements))
	}

	if !testVarStatement(t, block.Statements[0], "a", 10) {
		return
	}
}

func TestReturnStatement(t *testing.T) {
	input := `
		return 10;
		return 20;
		return 30;
		return;
	`

	program := createParseProgram(input)

	if len(program.Statements) != 4 {
		t.Fatalf("Expected length of statements to be %d, got=%d", 3, len(program.Statements))
	}

	for _, stmt := range program.Statements {
		ret, ok := stmt.(*ReturnStatement)
		if !ok {
			t.Errorf("Expected=%T, got=%T", &ReturnStatement{}, stmt)
		}

		if ret.TokenLiteral() != "return" {
			t.Errorf("Expected token literal 'return', got=%q", ret.TokenLiteral())
		}
	}
}

func TestVarStatement(t *testing.T) {
	tests := []struct {
		input              string
		expectedIdentifier string
		expectedValue      interface{}
	}{
		{
			"var a = 20;",
			"a",
			20,
		},
		{
			`var b =  "hello world";`,
			"b",
			"hello world",
		},
		{
			"var c = true;",
			"c",
			true,
		},
		{
			"var d = 5 + 5;",
			"d",
			10,
		},
	}

	for _, test := range tests {
		program := createParseProgram(test.input)

		if len(program.Statements) != 1 {
			t.Fatalf("Expected length of statement to be %d, got %d", 1, len(program.Statements))
		}

		if !testVarStatement(t, program.Statements[0], test.expectedIdentifier, test.expectedValue) {
			return
		}
	}
}

func testVarStatement(t *testing.T, s Statement, expectedIdentifer string, expectedValue interface{}) bool {
	variable, ok := s.(*VarStatement)
	if !ok {
		t.Errorf("Expected=%T, got=%T", &VarStatement{}, s)
		return false
	}

	if variable.TokenLiteral() != "var" {
		t.Errorf("Expected TokenLiteral to be %s, got=%s", "var", variable.TokenLiteral())
		return false
	}

	if variable.Identifier.Value != expectedIdentifer {
		t.Errorf("Expected Identifier %s, got=%s", expectedIdentifer, variable.Identifier.Value)
		return false
	}

	if variable.Identifier.TokenLiteral() != expectedIdentifer {
		t.Errorf("Expected Identifier %s, got=%s", expectedIdentifer, variable.Identifier.TokenLiteral())
		return false
	}

	if !testLiteral(t, variable.Expression, expectedValue) {
		return false
	}

	return true
}

func TestGroupedExpression(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{
			"(5);",
			5,
		},
		{
			`("hello world");`,
			"hello world",
		},
	}

	for _, test := range tests {
		program := createParseProgram(test.input)

		if len(program.Statements) != 1 {
			t.Fatalf("Expected length of statement to be %d, got %d", 1, len(program.Statements))
		}

		stmt, ok := program.Statements[0].(*ExpressionStatement)
		if !ok {
			t.Fatalf("Expected=%T, got=%T", &ExpressionStatement{}, program.Statements[0])
		}

		group, ok := stmt.Expression.(*GroupedExpression)
		if !ok {
			t.Fatalf("Expected=%T, got=%T", &GroupedExpression{}, stmt.Expression)
		}

		if !testLiteral(t, group.Expression, test.expected) {
			return
		}
	}
}

func TestOperatorPrecedence(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			"-a * b;",
			"((-a) * b)",
		},
		{
			"!-a;",
			"(!(-a))",
		},
		{
			"a + b + c;",
			"((a + b) + c)",
		},
		{
			"a * b * c;",
			"((a * b) * c)",
		},
		{
			"a * b / c;",
			"((a * b) / c)",
		},
		{
			"a + b * c + d / e - f;",
			"(((a + (b * c)) + (d / e)) - f)",
		},
	}

	for _, test := range tests {
		program := createParseProgram(test.input)

		actual := program.String()
		if actual != test.expected {
			t.Errorf("expected=%q, got=%q", test.expected, actual)
		}
	}
}

func TestUnary(t *testing.T) {
	tests := []struct {
		input    string
		operator string
		right    interface{}
	}{
		{"!5;", "!", 5},
		{"!false;", "!", false},
		{"-10;", "-", 10},
	}

	for _, test := range tests {
		program := createParseProgram(test.input)

		if len(program.Statements) != 1 {
			t.Fatalf("Expected length of statements to be %d, got=%d", 1, len(program.Statements))
		}

		stmt, ok := program.Statements[0].(*ExpressionStatement)
		if !ok {
			t.Fatalf("Expected=%T, got=%T", &ExpressionStatement{}, program.Statements[0])
		}

		unary, ok := stmt.Expression.(*Unary)
		if !ok {
			t.Fatalf("Expected=%T, got=%T", &Unary{}, stmt.Expression)
		}

		if unary.Operator != test.operator {
			t.Errorf("Expected operator to be %s, got=%s", test.operator, unary.Operator)
		}

		if !testLiteral(t, unary.Right, test.right) {
			return
		}
	}
}

func TestBinary(t *testing.T) {
	tests := []struct {
		input    string
		left     interface{}
		operator string
		right    interface{}
	}{
		{"5+5;", 5, "+", 5},
		{"5-5;", 5, "-", 5},
		{"5/5;", 5, "/", 5},
		{"5*5;", 5, "*", 5},
		{"5>5;", 5, ">", 5},
		{"5<5;", 5, "<", 5},
		{"5!=5;", 5, "!=", 5},
		{"5==5;", 5, "==", 5},
		{"true == false;", true, "==", false},
		{"false != false;", false, "==", false},
	}

	for _, test := range tests {
		program := createParseProgram(test.input)

		if len(program.Statements) != 1 {
			t.Fatalf("Expected length of statements to be %d, got=%d", 1, len(program.Statements))
		}

		stmt, ok := program.Statements[0].(*ExpressionStatement)
		if !ok {
			t.Fatalf("Expected=%T, got=%T", &ExpressionStatement{}, program.Statements[0])
		}

		if !testBinaryExpression(t, stmt.Expression, test.left, test.operator, test.right) {
			return
		}
	}
}

func testBinaryExpression(t *testing.T, exp Expression, left interface{}, operator string, right interface{}) bool {
	binary, ok := exp.(*Binary)
	if !ok {
		t.Fatalf("Expected=%T, got=%T", &Binary{}, exp)
		return false
	}

	if !testLiteral(t, binary.Left, left) {
		return false
	}

	if binary.Operator != operator {
		t.Errorf("Expected operator to be %s, got=%s", operator, binary.Operator)
		return false
	}

	if !testLiteral(t, binary.Right, right) {
		return false
	}

	return true
}

func TestAssignment(t *testing.T) {
	tests := []struct {
		input              string
		expectedIdentifier string
		expectedValue      interface{}
	}{
		{
			input:              "a = 20;",
			expectedIdentifier: "a",
			expectedValue:      20,
		},
		{
			input:              `b = "hello world";`,
			expectedIdentifier: "b",
			expectedValue:      "hello world",
		},
	}

	for _, test := range tests {
		program := createParseProgram(test.input)

		if len(program.Statements) != 1 {
			t.Fatalf("Expected length of statements to be %d, got=%d", 1, len(program.Statements))
		}

		stmt, ok := program.Statements[0].(*ExpressionStatement)
		if !ok {
			t.Fatalf("Expected=%T, got=%T", &ExpressionStatement{}, program.Statements[0])
		}

		assignment, ok := stmt.Expression.(*Assignment)
		if !ok {
			t.Fatalf("Expected=%T, got=%T", &Identifier{}, stmt.Expression)
		}

		if assignment.Identifier.Value != test.expectedIdentifier {
			t.Errorf("Expected assignment value %s, got=%s", test.expectedIdentifier, assignment.Identifier.Value)
		}

		if !testLiteral(t, assignment.Expression, test.expectedValue) {
			return
		}
	}
}

func TestIntegerLiteral(t *testing.T) {
	input := `1;`

	program := createParseProgram(input)

	if len(program.Statements) != 1 {
		t.Fatalf("Expected length of statement to be %d, got=%d", 1, len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ExpressionStatement)
	if !ok {
		t.Fatalf("Expected=%T, got=%T", &ExpressionStatement{}, program.Statements[0])
	}

	if !testFloat(t, stmt.Expression, 1) {
		return
	}
}

func TestBooleanLiteral(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{
			input:    "true;",
			expected: true,
		},
		{
			input:    "false;",
			expected: false,
		},
	}

	for _, test := range tests {
		program := createParseProgram(test.input)

		if len(program.Statements) != 1 {
			t.Fatalf("Expected length of statements to be %d, got=%d", 1, len(program.Statements))
		}

		stmt, ok := program.Statements[0].(*ExpressionStatement)
		if !ok {
			t.Fatalf("Expected=%T, got=%T", &ExpressionStatement{}, program.Statements[0])
		}

		if !testBoolean(t, stmt.Expression, test.expected) {
			return
		}
	}
}

func TestStringLiteral(t *testing.T) {
	input := `"hello world";`

	program := createParseProgram(input)

	if len(program.Statements) != 1 {
		t.Fatalf("Expected length of statements to be %d, got %d", 1, len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ExpressionStatement)
	if !ok {
		t.Fatalf("Expected=%T, got=%T", &ExpressionStatement{}, program.Statements[0])
	}

	if !testString(t, stmt.Expression, "hello world") {
		return
	}
}

func testLiteral(t *testing.T, expr Expression, expected interface{}) bool {

	switch e := expected.(type) {
	case float64:
		return testFloat(t, expr, e)
	case string:
		return testString(t, expr, e)
	case bool:
		return testBoolean(t, expr, e)
	default:
		return false
	}
}

func testFloat(t *testing.T, expr Expression, expected float64) bool {
	t.Helper()
	num, ok := expr.(*NumberLiteral)
	if !ok {
		t.Fatalf("Expected=%T, got=%T", &NumberLiteral{}, expr)
		return false
	}

	if num.Value != expected {
		t.Errorf("Expected NumberLiteral value to be %g, got=%g", expected, num.Value)
		return false
	}

	if num.TokenLiteral() != fmt.Sprintf("%g", expected) {
		t.Errorf("Expected NumberLiteral to be %g, got=%s", expected, num.TokenLiteral())
		return false
	}

	return true
}

func testBoolean(t *testing.T, expr Expression, expected bool) bool {
	t.Helper()
	b, ok := expr.(*BooleanLiteral)
	if !ok {
		t.Fatalf("Expected=%T, got=%T", &BooleanLiteral{}, expr)
		return false
	}

	if b.Value != expected {
		t.Errorf("Expected BooleanLiteral value to be %t, got=%t", expected, b.Value)
		return false
	}

	if b.TokenLiteral() != fmt.Sprintf("%t", expected) {
		t.Errorf("Expected BooleanLiteral value to be %t, got %s", expected, b.TokenLiteral())
		return false
	}

	return true
}

func testString(t *testing.T, expr Expression, expected string) bool {
	t.Helper()
	var value, tokenLiteral string = "", ""

	str, ok := expr.(*StringLiteral)
	if ok {
		value = str.Value
		tokenLiteral = str.TokenLiteral()
	} else {
		str, ok := expr.(*Identifier)
		if !ok {
			t.Fatalf("Expected=%T, got=%T", &StringLiteral{}, expr)
			return false
		}
		value = str.Value
		tokenLiteral = str.Value
	}

	if value != expected {
		t.Errorf("Expected StringLiteral value to be %s, got=%s", expected, str.Value)
		return false
	}

	if tokenLiteral != expected {
		t.Errorf("Expected StringLiteral token to be %s, got=%s", expected, str.Value)
		return false
	}

	return true
}
