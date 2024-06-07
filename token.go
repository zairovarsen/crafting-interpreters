package main

import (
	"fmt"
)

type TokenType string

const (
	// Single character tokens
	LEFT_PAREN    TokenType = "("
	RIGHT_PAREN   TokenType = ")"
	LEFT_BRACE    TokenType = "["
	RIGHT_BRACE   TokenType = "]"
	LEFT_BRACKET  TokenType = "{"
	RIGHT_BRACKET TokenType = "}"
	COMMA         TokenType = ","
	DOT           TokenType = "."
	MINUS         TokenType = "-"
	PLUS          TokenType = "+"
	SEMICOLON     TokenType = ";"
	SLASH         TokenType = "/"
	STAR          TokenType = "*"
	QUESTION      TokenType = "?"
	COLON         TokenType = ":"

	// One or two character tokens.
	BANG          TokenType = "!"
	BANG_EQUAL    TokenType = "!="
	EQUAL         TokenType = "="
	EQUAL_EQUAL   TokenType = "=="
	GREATER       TokenType = ">"
	GREATER_EQUAL TokenType = ">="
	LESS          TokenType = "<"
	LESS_EQUAL    TokenType = "<="

	// Literals.
	IDENTIFIER TokenType = "IDENTIFIER"
	STRING     TokenType = "STRING"
	NUMBER     TokenType = "NUMBER"

	// Keywords.
	AND      TokenType = "AND"
	CLASS    TokenType = "CLASS"
	ELSE     TokenType = "ELSE"
	FALSE    TokenType = "FALSE"
	FUN      TokenType = "FUNCTION"
	FOR      TokenType = "FOR"
	IF       TokenType = "IF"
	NIL      TokenType = "NIL"
	OR       TokenType = "OR"
	PRINT    TokenType = "PRINT"
	RETURN   TokenType = "RETURN"
	SUPER    TokenType = "SUPER"
	THIS     TokenType = "THIS"
	TRUE     TokenType = "TRUE"
	VAR      TokenType = "VAR"
	WHILE    TokenType = "WHILE"
	BREAK    TokenType = "BREAK"
	CONTINUE TokenType = "CONTINUE"

	EOF TokenType = "EOF"
)

var reserved = map[string]TokenType{
	"and":      AND,
	"or":       OR,
	"class":    CLASS,
	"else":     ELSE,
	"false":    FALSE,
	"for":      FOR,
	"fun":      FUN,
	"if":       IF,
	"nil":      NIL,
	"print":    PRINT,
	"return":   RETURN,
	"super":    SUPER,
	"this":     THIS,
	"true":     TRUE,
	"var":      VAR,
	"while":    WHILE,
	"break":    BREAK,
	"continue": CONTINUE,
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
