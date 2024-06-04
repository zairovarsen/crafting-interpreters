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
	expr := parser.parse()

	return expr
}

func TestReturnStatement(t *testing.T) {
	input := `
		return 10
		return 20
		return 30
	`

	program := createParseProgram(input)

	if len(program.Statements) != 3 {
		t.Fatalf("Expected length of statements to be %d, got=%d", len(program.Statements), 3)
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

func TestGroupedExpression(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{
			"(5)",
			5,
		},
		{
			`("hello world")`,
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
			"-a * b",
			"((-a) * b)",
		},
		{
			"!-a",
			"(!(-a))",
		},
		{
			"a + b + c",
			"((a + b) + c)",
		},
		{
			"a * b * c",
			"((a * b) * c)",
		},
		{
			"a * b / c",
			"((a * b) / c)",
		},
		{
			"a + b * c + d / e - f",
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
		{"!5", "!", 5},
		{"!false", "!", false},
		{"-10", "-", 10},
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
		{"5+5", 5, "+", 5},
		{"5-5", 5, "-", 5},
		{"5/5", 5, "/", 5},
		{"5*5", 5, "*", 5},
		{"5>5", 5, ">", 5},
		{"5<5", 5, "<", 5},
		{"5!=5", 5, "!=", 5},
		{"5==5", 5, "==", 5},
		{"true == false", true, "==", false},
		{"false != false", false, "==", false},
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

		binary, ok := stmt.Expression.(*Binary)
		if !ok {
			t.Fatalf("Expected=%T, got=%T", &Binary{}, stmt.Expression)
		}

		if binary.Operator != test.operator {
			t.Errorf("Expected operator to be %s, got=%s", test.operator, binary.Operator)
		}

		if !testLiteral(t, binary.Left, test.left) {
			return
		}

		if !testLiteral(t, binary.Right, test.right) {
			return
		}
	}
}

func TestAssignment(t *testing.T) {
	tests := []struct {
		input              string
		expectedIdentifier string
		expectedValue      interface{}
	}{
		{
			input:              "a = 20",
			expectedIdentifier: "a",
			expectedValue:      20,
		},
		{
			input:              `b = "hello world"`,
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

		if !testLiteral(t, assignment.Right, test.expectedValue) {
			return
		}
	}
}

func TestIntegerLiteral(t *testing.T) {
	input := `1`

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
			input:    "true",
			expected: true,
		},
		{
			input:    "false",
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
	input := `"hello world"`

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
	str, ok := expr.(*StringLiteral)
	if !ok {
		t.Fatalf("Expected=%T, got=%T", &StringLiteral{}, expr)
	}

	if str.Value != expected {
		t.Errorf("Expected StringLiteral value to be %s, got=%s", expected, str.Value)
		return false
	}

	if str.TokenLiteral() != expected {
		t.Errorf("Expected StringLiteral token to be %s, got=%s", expected, str.Value)
		return false
	}

	return true
}
