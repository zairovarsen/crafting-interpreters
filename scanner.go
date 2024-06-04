package main

import (
	"fmt"
)

type Scanner struct {
	source  []byte
	start   int
	current int
	line    int
	Tokens  []*Token
	Errors  *ErrorHandler
}

func NewScanner(source []byte) *Scanner {
	tokens := make([]*Token, 0)
	errors := NewErrorHandler()
	return &Scanner{source, 0, 0, 1, tokens, errors}
}

func (s *Scanner) isAtEnd() bool {
	return s.current >= len(s.source)
}

// positive lookahead doesn't consume the character
func (s *Scanner) peek() byte {
	if s.isAtEnd() {
		return '\000'
	}
	return s.source[s.current]

}

func (s *Scanner) isDigit(r byte) bool {
	return '0' <= r && r <= '9'
}

func (s *Scanner) isAlpha(c byte) bool {
	return (c >= 'a' && c <= 'z') ||
		(c >= 'A' && c <= 'Z') ||
		c == '_'
}

func (s *Scanner) addError(err *Error) {
	s.Errors.AddError(err)
}

func (s *Scanner) addToken(tokenType TokenType, lexeme string) {
	token := NewToken(tokenType, lexeme, s.line)
	s.Tokens = append(s.Tokens, token)
}

func (s *Scanner) advance() byte {
	s.current++
	return s.source[s.current-1]
}

func (s *Scanner) match(expected byte) bool {
	if s.isAtEnd() {
		return false
	}

	if s.source[s.current] != expected {
		return false
	}

	s.current++
	return true
}

func (s *Scanner) scanTokens() []*Error {
	for !s.isAtEnd() {
		// we are at the beginning of next lexeme
		s.start = s.current
		s.scanToken()
	}

	s.addToken(EOF, "0")
	return nil
}

func (s *Scanner) getNumber() {
	for s.isDigit(s.peek()) {
		s.advance()
	}

	// Look for a fractional part
	if s.peek() == '.' && s.isDigit(s.peekNext()) {
		// Consume the "."
		s.advance()

		for s.isDigit(s.peek()) {
			s.advance()
		}
	}
	s.addToken(NUMBER, string(s.source[s.start:s.current]))
}

func (s *Scanner) isAlphaNumeric(c byte) bool {
	return s.isDigit(c) || s.isAlpha(c)
}

func (s *Scanner) identifier() {
	for s.isAlphaNumeric(s.peek()) {
		s.advance()
	}

	text := string(s.source[s.start:s.current])
	tokenType := GetReserved(text)
	s.addToken(tokenType, text)
}

func (s *Scanner) peekNext() byte {
	if s.current+1 >= len(s.source) {
		return '\000'
	}
	return s.source[s.current+1]
}

func (s *Scanner) getString() {
	for s.peek() != '"' && !s.isAtEnd() {
		if s.peek() == '\n' {
			s.line++
		}
		s.advance()
	}

	if s.isAtEnd() {
		msg := fmt.Sprintf("Unexpected character: %q", s.peek())
		err := &Error{Message: msg, Line: s.line}
		s.addError(err)
	}

	// The closing  ".
	s.advance()

	text := string(s.source[s.start+1 : s.current-1])
	s.addToken(STRING, text)
}

func (s *Scanner) scanToken() {
	c := s.advance()
	switch c {
	case '(':
		s.addToken(LEFT_PAREN, "(")
	case ')':
		s.addToken(RIGHT_PAREN, ")")
	case '[':
		s.addToken(LEFT_BRACE, "[")
	case ']':
		s.addToken(RIGHT_BRACE, "]")
	case '{':
		s.addToken(LEFT_BRACKET, "{")
	case '}':
		s.addToken(RIGHT_BRACKET, "}")
	case ',':
		s.addToken(COMMA, ",")
	case '.':
		s.addToken(DOT, ".")
	case '-':
		s.addToken(MINUS, "-")
	case '+':
		s.addToken(PLUS, "+")
	case ';':
		s.addToken(SEMICOLON, ";")
	case ':':
		s.addToken(COLON, ":")
	case '?':
		s.addToken(QUESTION, "?")
	case '*':
		s.addToken(STAR, "*")
	case '!':
		if s.match('=') {
			s.addToken(BANG_EQUAL, "!=")
		} else {
			s.addToken(BANG, "!")
		}
	case '=':
		if s.match('=') {
			s.addToken(EQUAL_EQUAL, "==")
		} else {
			s.addToken(EQUAL, "=")
		}
	case '<':
		if s.match('=') {
			s.addToken(LESS_EQUAL, "<=")
		} else {
			s.addToken(LESS, "<")
		}
	case '>':
		if s.match('=') {
			s.addToken(GREATER_EQUAL, ">=")
		} else {
			s.addToken(GREATER, ">")
		}
	case '/':
		if s.match('/') {
			for s.peek() != '\n' && !s.isAtEnd() {
				s.advance()
			}
		} else {
			s.addToken(SLASH, "/")
		}
	case ' ', '\r', '\t':
		// Ignore whitespaces
	case '\n':
		s.line++
		break
	case '"':
		s.getString()
	default:
		if s.isDigit(c) {
			s.getNumber()
		} else if s.isAlpha(c) {
			s.identifier()
		} else {
			msg := fmt.Sprintf("Unexpected character: %q", c)
			err := &Error{Message: msg, Line: s.line}
			s.addError(err)
		}
	}
}
