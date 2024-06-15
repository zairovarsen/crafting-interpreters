package main

import (
	"fmt"
)

type TokenType string

const (
	// Single character tokens
	LEFT_PAREN    = "("
	RIGHT_PAREN   = ")"
	LEFT_BRACE    = "["
	RIGHT_BRACE   = "]"
	LEFT_BRACKET  = "{"
	RIGHT_BRACKET = "}"
	COMMA         = ","
	DOT           = "."
	MINUS         = "-"
	PLUS          = "+"
	SEMICOLON     = ";"
	SLASH         = "/"
	STAR          = "*"
	QUESTION      = "?"
	COLON         = ":"

	// One or two character tokens.
	BANG          = "!"
	BANG_EQUAL    = "!="
	EQUAL         = "="
	EQUAL_EQUAL   = "=="
	GREATER       = ">"
	GREATER_EQUAL = ">="
	LESS          = "<"
	LESS_EQUAL    = "<="

	// Literals.
	IDENTIFIER = "IDENTIFIER"
	STRING     = "STRING"
	NUMBER     = "NUMBER"

	// Keywords.
	AND      = "AND"
	CLASS    = "CLASS"
	ELSE     = "ELSE"
	FALSE    = "FALSE"
	FUNCTION = "FUNCTION"
	FOR      = "FOR"
	IF       = "IF"
	NIL      = "NIL"
	OR       = "OR"
	RETURN   = "RETURN"
	SUPER    = "SUPER"
	THIS     = "THIS"
	TRUE     = "TRUE"
	VAR      = "VAR"
	WHILE    = "WHILE"
	BREAK    = "BREAK"
	CONTINUE = "CONTINUE"
	STATIC   = "STATIC"

	EOF = "EOF"
)

var reserved = map[string]TokenType{
	"and":      AND,
	"or":       OR,
	"class":    CLASS,
	"else":     ELSE,
	"false":    FALSE,
	"for":      FOR,
	"function": FUNCTION,
	"if":       IF,
	"nil":      NIL,
	"return":   RETURN,
	"super":    SUPER,
	"this":     THIS,
	"true":     TRUE,
	"var":      VAR,
	"while":    WHILE,
	"break":    BREAK,
	"continue": CONTINUE,
	"static":   STATIC,
}

type Token struct {
	Type   TokenType
	Lexeme string
	Line   int
}

func NewToken(tokenType TokenType, lexeme string, line int) *Token {
	return &Token{
		Type:   tokenType,
		Lexeme: lexeme,
		Line:   line,
	}
}

func GetReserved(text string) TokenType {
	r, ok := reserved[text]
	if ok {
		return r
	}
	return IDENTIFIER
}

func (t *Token) String() string {
	return fmt.Sprintf("Token [Type: %s, Lexeme: %s, Line: %d]", t.Type, t.Lexeme, t.Line)
}
