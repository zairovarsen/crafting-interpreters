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
						print a;
						print b;
						print c;
					}
					print a;
					print b;
					print c;
				}
				print a;
				print b;
				print c;
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

func testIntegerObject(t *testing.T, obj Object, expected float64) bool {
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
