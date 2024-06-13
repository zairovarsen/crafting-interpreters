package main

import (
	"bytes"
	"io"
	"os"
	"testing"
)

func runInterpreter(input []byte) Object {
	scanner := NewScanner(input)
	scanner.scanTokens()
	env := NewEnvironment()

	if scanner.Errors.HasErrors() {
		scanner.Errors.PrintErrors()
		return nil
	}

	parser := NewParser(scanner.Tokens)
	program := parser.parse()

	if parser.Errors.HasErrors() {
		parser.Errors.PrintErrors()
		return nil
	}
	interpreter := NewInterpreter()
	return interpreter.Interpret(program, env)
}

func TestThis(t *testing.T) {
	tests := []struct {
		code     string
		expected string
	}{
		{
			`class Cake {
				taste() {
					var adjective = "delicious";
					print("The " + this.flavor + " cake is " + adjective + "!");
				}
			}
			
			var cake = Cake();
			cake.flavor = "German";
			cake.taste();
			`,
			"The German cake is delicious!\n",
		},
	}

	for _, test := range tests {

		// Save the current stdout
		originalStdout := os.Stdout

		// Create a pipe to capture stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		// Run the interpreter
		runInterpreter([]byte(test.code))

		// Close the writer and restore stdout
		w.Close()
		os.Stdout = originalStdout

		// Read the output
		var buf bytes.Buffer
		io.Copy(&buf, r)

		if buf.String() != test.expected {
			t.Errorf("Expected=%s, got=%s", test.expected, buf.String())
		}
	}
}

func TestSetProperty(t *testing.T) {
	tests := []struct {
		code     string
		expected float64
	}{
		{
			`class Hello {}
			 var h = Hello();
			 h.something = 20;
			`,
			20,
		},
	}

	for _, test := range tests {
		result := runInterpreter([]byte(test.code))

		if !testLiteralObject(t, result, test.expected) {
			return
		}
	}
}

func TestGetProperty(t *testing.T) {
	tests := []struct {
		code     string
		expected float64
	}{
		{
			`class Hello {}
			 var h = Hello();
			 h.something = 20;
			 h.something;
			`,
			20,
		},
	}

	for _, test := range tests {
		result := runInterpreter([]byte(test.code))

		if !testLiteralObject(t, result, test.expected) {
			return
		}
	}
}

func TestRecursion(t *testing.T) {
	tests := []struct {
		code     string
		expected float64
	}{
		{
			`function fib(a) {
	  if (a == 0) {
	    return 0;
	  } else if (a == 1) {
	    return 1;
	  } else {
	    return fib(a - 2) + fib(a - 1);
	  }
	}
	fib(10);`,
			55,
		},
	}

	for _, test := range tests {
		result := runInterpreter([]byte(test.code))

		if !testLiteralObject(t, result, test.expected) {
			return
		}
	}
}

func TestFunction(t *testing.T) {
	tests := []struct {
		code     string
		expected interface{}
	}{
		{
			code: `
				function add(a,b) {
					print(a + b);
				}
				add(4,5);
			`,
			expected: 9,
		},
		{
			code: `
				var a = 20;
				function add(b) {
					print(a + b);
				}
				add(5)
			`,
			expected: 25,
		},
	}

	for _, test := range tests {
		result := runInterpreter([]byte(test.code))

		if !testLiteralObject(t, result, test.expected) {
			return
		}
	}
}

func TestForLoop(t *testing.T) {
	tests := []struct {
		code     string
		expected interface{}
	}{
		{
			code: `
			var i = 0;
			for (; i < 5; i = i + 1) {}
			i;	
			`,
			expected: 5,
		},
		{
			code: `
				var sum = 0;
				var i = 1;
				for (; i <= 5; i = i + 1) {
					sum = sum + i;
					i = i + 1;
				}
				sum;
			`,
			expected: 15,
		},
		{
			code: `
				var sum = 0;
				var i = 1;
				for (; i <= 5; i = i + 1) {
					if (sum == 5) {
						break;
					}
					sum = sum + i;
					i = i + 1;
				}
				sum;
			`,
			expected: 5,
		},
		{
			code: `
				var sum = 0;
				var i = 1;
				for (; i <= 5; i = i + 1) {
					if (sum / 5 == 0) {
						continue;
					}
					sum = sum + i;
					i = i + 1;
				}
				sum;
			`,
			expected: 10,
		},
	}

	for _, test := range tests {
		result := runInterpreter([]byte(test.code))

		if !testLiteralObject(t, result, test.expected) {
			return
		}
	}
}

func TestWhileLoop(t *testing.T) {
	tests := []struct {
		code     string
		expected interface{}
	}{
		{
			code: `
				var i = 0;
				while (i < 5) {
					i = i + 1;
				}
				i;
			`,
			expected: 5,
		},
		{
			code: `
				var sum = 0;
				var i = 1;
				while (i <= 5) {
					sum = sum + i;
					i = i + 1;
				}
				sum;
			`,
			expected: 15,
		},
		{
			code: `
				var a = 0;
				var b = 10;
				while (a < b) {
					if (a == 5) {
						break;
					}
					a = a + 1;
				}
				a;
			`,
			expected: 5,
		},
		{
			code: `
				var i = 0;
				var result = 0;
				while (i < 5) {
					i = i + 1;
					if (i > 2) {
						continue;
					}
					result = result + i;
				}
				result;
			`,
			expected: 3, // 1 + 3 + 5
		},
	}

	for _, test := range tests {
		result := runInterpreter([]byte(test.code))

		if !testLiteralObject(t, result, test.expected) {
			return
		}
	}
}

func TestLogicalAndConditional(t *testing.T) {
	tests := []struct {
		code     string
		expected interface{}
	}{
		{"true and 5;", 5},
		{"0 and false;", 0},
		{`"hello world" and true;`, true},
		{"true && 5 and false;", false},
	}

	for _, test := range tests {
		result := runInterpreter([]byte(test.code))

		if !testLiteralObject(t, result, test.expected) {
			return
		}
	}
}

func TestLogicalOrConditional(t *testing.T) {
	tests := []struct {
		code     string
		expected interface{}
	}{
		{"true or 5;", true},
		{"0 or false;", false},
		{`"hello world" or true;`, "hello world"},
		{"true and 5 or false;", 5},
	}

	for _, test := range tests {
		result := runInterpreter([]byte(test.code))

		if !testLiteralObject(t, result, test.expected) {
			return
		}
	}
}

func TestTernaryConditional(t *testing.T) {
	tests := []struct {
		code     string
		expected interface{}
	}{
		{"true ? 5 : 10", 5},
		{`false ? 5 : "hello world"`, "hello world"},
		{`5 ? 10 : 15`, 10},
	}

	for _, test := range tests {
		result := runInterpreter([]byte(test.code))

		if !testLiteralObject(t, result, test.expected) {
			return
		}
	}
}

func TestConditionalThenBranch(t *testing.T) {
	tests := []struct {
		code     string
		expected interface{}
	}{
		{"if (5 > 2) { 5; }", 5},
		{`if (2 < 3) { "hello world"; }`, "hello world"},
	}

	for _, test := range tests {
		result := runInterpreter([]byte(test.code))

		if !testLiteralObject(t, result, test.expected) {
			return
		}
	}
}

func TestConditionalElseBranch(t *testing.T) {
	tests := []struct {
		code     string
		expected interface{}
	}{
		{"if (5 > 2) { 10; } else { 5; }", 5},
		{`if (2 > 3) { "john doe"; } else { "hello world"; }`, "hello world"},
	}

	for _, test := range tests {
		result := runInterpreter([]byte(test.code))

		if !testLiteralObject(t, result, test.expected) {
			return
		}
	}
}

func TestIntegerObject(t *testing.T) {
	tests := []struct {
		code     string
		expected float64
	}{
		{"5;", 5},
		{"10;", 10},
		{"10 + 10;", 20},
		{"-5;", -5},
		{"5 * 2 + 10;", 20},
		{"20 / 10 * 2;", 4},
		{"10 + 2 * 4;", 18},
		{"3 * (5 + 5);", 30},
	}

	for _, test := range tests {
		result := runInterpreter([]byte(test.code))

		if !testFloatObject(t, result, test.expected) {
			return
		}
	}
}

func TestBooleanObject(t *testing.T) {
	tests := []struct {
		code     string
		expected bool
	}{
		{"true;", true},
		{"false;", false},
		{"!true;", false},
		{"!false;", true},
		{"!5;", false},
	}

	for _, test := range tests {
		result := runInterpreter([]byte(test.code))

		if !testBooleanObject(t, result, test.expected) {
			return
		}
	}
}

func TestBlockStatements(t *testing.T) {
	tests := []struct {
		code     string
		expected string
	}{
		{
			code: `
				var a = "global a";
				var b = "global b";
				var c = "global c";
				{
					var a = "outer a";
					var b = "outer b";
					{
						var a = "inner a";
						print(a);
						print(b);
						print(c);
					}
					print(a);
					print(b);
					print(c);
				}
				print(a);
				print(b);
				print(c);
			`,
			expected: "inner a\nouter b\nglobal c\nouter a\nouter b\nglobal c\nglobal a\nglobal b\nglobal c\n",
		},
	}

	for _, test := range tests {

		// Save the current stdout
		originalStdout := os.Stdout

		// Create a pipe to capture stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		// Run the interpreter
		runInterpreter([]byte(test.code))

		// Close the writer and restore stdout
		w.Close()
		os.Stdout = originalStdout

		// Read the output
		var buf bytes.Buffer
		io.Copy(&buf, r)

		if buf.String() != test.expected {
			t.Errorf("Expected %s, got=%s", test.expected, buf.String())
		}
	}
}

func testLiteralObject(t *testing.T, obj Object, expected interface{}) bool {
	switch exp := expected.(type) {
	case string:
		return testStringObject(t, obj, exp)
	case float64:
		return testFloatObject(t, obj, exp)
	case bool:
		return testBooleanObject(t, obj, exp)
	default:
		return false
	}
}

func testFloatObject(t *testing.T, obj Object, expected float64) bool {
	if obj.Type() != FloatObj {
		t.Errorf("object is not %T, got=%T", FloatObj, obj.Type())
		return false
	}

	result, ok := obj.(*FloatObject)
	if !ok {
		t.Errorf("object is not %T, got=%T", &FloatObject{}, result)
		return false
	}

	if result.Value != expected {
		t.Errorf("object has wrong value. got=%f, want=%f", result.Value, expected)
		return false
	}

	return true
}

func testStringObject(t *testing.T, obj Object, expected string) bool {
	if obj.Type() != StringObj {
		t.Errorf("object is not %s, got=%s", StringObj, obj.Type())
		return false
	}

	result, ok := obj.(*StringObject)
	if !ok {
		t.Errorf("object is not %T, got=%T", &StringObject{}, obj)
		return false
	}

	if result.Value != expected {
		t.Errorf("object has wrong value. got=%s, want=%s", result.Value, expected)
		return false
	}

	return true
}

func testBooleanObject(t *testing.T, obj Object, expected bool) bool {
	if obj.Type() != BooleanObj {
		t.Errorf("object is not %T, got=%T", BooleanObj, obj.Type())
		return false
	}

	result, ok := obj.(*BooleanObject)
	if !ok {
		t.Errorf("object is not %T, got=%T", &BooleanObject{}, obj)
		return false
	}

	if result.Value != expected {
		t.Errorf("object has wrong value. got=%v, want=%v", result.Value, expected)
		return false
	}

	return true
}

func testNullObject(t *testing.T, obj Object) bool {
	if obj.Type() != NillObj {
		t.Errorf("object is not Nil. got=%T (%+v)", obj, obj)
		return false
	}
	return true
}
