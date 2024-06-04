package main

import (
	"testing"
)

func TestScanner(t *testing.T) {
	tests := []struct {
		input    string
		expected []*Token
	}{
		{
			// single character tokens
			input: "()[],.-+;/*",
			expected: []*Token{NewToken(LEFT_PAREN, "(", 1),
				NewToken(RIGHT_PAREN, ")", 1),
				NewToken(LEFT_BRACE, "[", 1),
				NewToken(RIGHT_BRACE, "]", 1),
				NewToken(COMMA, ",", 1),
				NewToken(DOT, ".", 1),
				NewToken(MINUS, "-", 1),
				NewToken(PLUS, "+", 1),
				NewToken(SEMICOLON, ";", 1),
				NewToken(SLASH, "/", 1),
				NewToken(STAR, "*", 1),
				NewToken(EOF, "0", 1),
			},
		},
		{
			// One or two characters tokens.
			input: "!!====>>=<<=",
			expected: []*Token{
				NewToken(BANG, "!", 1),
				NewToken(BANG_EQUAL, "!=", 1),
				NewToken(EQUAL_EQUAL, "==", 1),
				NewToken(EQUAL, "=", 1),
				NewToken(GREATER, ">", 1),
				NewToken(GREATER_EQUAL, ">=", 1),
				NewToken(LESS, "<", 1),
				NewToken(LESS_EQUAL, "<=", 1),
				NewToken(EOF, "0", 1),
			},
		},
		{
			input: `"This is a string"`,
			expected: []*Token{
				NewToken(STRING, `This is a string`, 1),
				NewToken(EOF, "0", 1),
			},
		},
		{
			input: "counter",
			expected: []*Token{
				NewToken(IDENTIFIER, "counter", 1),
				NewToken(EOF, "0", 1),
			},
		},
		{
			input: "if",
			expected: []*Token{
				NewToken(IF, "if", 1),
				NewToken(EOF, "0", 1),
			},
		},
		{
			input: "while",
			expected: []*Token{
				NewToken(WHILE, "while", 1),
				NewToken(EOF, "0", 1),
			},
		},
		{
			input: "this",
			expected: []*Token{
				NewToken(THIS, "this", 1),
				NewToken(EOF, "0", 1),
			},
		},
	}

	for _, test := range tests {
		scanner := NewScanner([]byte(test.input))
		scanner.scanTokens()

		if len(scanner.Tokens) != len(test.expected) {
			t.Fatalf("Incorrect tokens length: expected=%d, got=%d", len(test.expected), len(scanner.Tokens))
		}

		for i, token := range scanner.Tokens {
			if token.Type != test.expected[i].Type {
				t.Errorf("Token type mismatch: expected=%s, got=%s", test.expected[i].Type, token.Type)
			}

			if token.Lexeme != test.expected[i].Lexeme {
				t.Errorf("Token lexeme mismatch: expeceted=%s, got=%s", test.expected[i].Lexeme, token.Lexeme)
			}

			if token.Line != test.expected[i].Line {
				t.Errorf("Token line mismatch: expected=%d, got=%d", test.expected[i].Line, token.Line)
			}

		}
	}
}
