package main

import (
	"fmt"
	"strconv"
)

type Parser struct {
	Errors  *ErrorHandler
	Tokens  []*Token
	current int
}

func NewParser(tokens []*Token) *Parser {
	errors := NewErrorHandler()

	return &Parser{
		Tokens:  tokens,
		current: 0,
		Errors:  errors,
	}
}

func (p *Parser) addError(err *Error) {
	p.Errors.AddError(err)
}

func (p *Parser) match(types ...TokenType) bool {
	for _, t := range types {
		if p.check(t) {
			p.advance()
			return true
		}
	}

	return false
}

func (p *Parser) peek() Token {
	return *p.Tokens[p.current]
}

func (p *Parser) previous() Token {
	return *p.Tokens[p.current-1]
}

func (p *Parser) isAtEnd() bool {
	return p.peek().Type == EOF
}

func (p *Parser) check(tokenType TokenType) bool {
	if p.isAtEnd() {
		return false
	}
	return p.peek().Type == tokenType
}

func (p *Parser) advance() Token {
	if !p.isAtEnd() {
		p.current++
	}
	return p.previous()
}

func (p *Parser) parse() *Program {
	program := &Program{}
	stmts := make([]Statement, 0)

	for !p.isAtEnd() {
		stmt := p.parseStatement()
		if stmt != nil {
			stmts = append(stmts, stmt)
		} else {
			// try to return back to some normal usable state
			p.synchronize()
		}
	}
	program.Statements = stmts
	return program
}

func (p *Parser) parseStatement() Statement {
	switch p.peek().Type {
	case RETURN:
		return p.parseReturnStatement()
	default:
		return p.parseExpressionStatement()
	}
}

func (p *Parser) parseReturnStatement() Statement {
	stmt := &ReturnStatement{Token: p.peek()}

	p.advance()

	stmt.ReturnValue = p.expression()
	return stmt
}

func (p *Parser) parseExpressionStatement() Statement {
	stmt := &ExpressionStatement{Token: p.peek()}

	stmt.Expression = p.expression()
	return stmt
}

func (p *Parser) consume(tokenType TokenType) bool {
	if !p.check(tokenType) {
		err := &Error{Message: fmt.Sprintf("Expect peek to be %s, got=%s", tokenType, p.peek().Type), Line: p.peek().Line}
		p.addError(err)
		return false
	} else {
		p.advance()
		return true
	}
}

// Discard tokens until we're right at the beginning of the next statement
// After catching parse error we'll call this and then we are hopefully back in sync.
func (p *Parser) synchronize() {
	p.advance()

	for !p.isAtEnd() {
		if p.previous().Type == SEMICOLON {
			return
		}

		switch p.peek().Type {
		case CLASS, FUN, VAR, FOR, IF, WHILE, PRINT, RETURN:
			return
		}

		p.advance()
	}
}

func (p *Parser) primary() Expression {
	if p.match(FALSE) {
		return &BooleanLiteral{Token: p.previous(), Value: false}
	}
	if p.match(TRUE) {
		return &BooleanLiteral{Token: p.previous(), Value: true}
	}
	if p.match(NIL) {
		return &NilLiteral{Token: p.previous()}
	}
	if p.match(STRING) {
		return &StringLiteral{Token: p.previous(), Value: p.previous().Lexeme}
	}
	if p.match(NUMBER) {
		num, err := strconv.ParseFloat(p.previous().Lexeme, 64)
		if err != nil {
			p.addError(&Error{Message: err.Error(), Line: p.previous().Line})
			return nil
		}
		return &NumberLiteral{Token: p.previous(), Value: num}
	}
	if p.match(LEFT_PAREN) {
		expr := p.expression()
		if p.consume(RIGHT_PAREN) {
			return &GroupedExpression{Token: p.previous(), Expression: expr}
		}
	}
	if p.match(IDENTIFIER) {
		return &Identifier{
			Token: p.previous(),
			Value: p.previous().Lexeme,
		}
	}

	p.addError(&Error{Message: "Expect expression.", Line: p.previous().Line})
	return nil
}

func (p *Parser) unary() Expression {
	for p.match(BANG, MINUS) {
		operator := p.previous()
		right := p.unary()
		return &Unary{
			Token:    p.peek(),
			Operator: operator.Lexeme,
			Right:    right,
		}
	}

	return p.primary()
}

func (p *Parser) factor() Expression {
	expr := p.unary()

	for p.match(SLASH, STAR) {
		operator := p.previous()
		right := p.unary()

		expr = &Binary{
			Token:    p.peek(),
			Left:     expr,
			Operator: operator.Lexeme,
			Right:    right,
		}
	}

	return expr
}

func (p *Parser) term() Expression {
	expr := p.factor()

	for p.match(MINUS, PLUS) {
		operator := p.previous()
		right := p.factor()
		expr = &Binary{
			Token:    p.peek(),
			Left:     expr,
			Operator: operator.Lexeme,
			Right:    right,
		}
	}

	return expr
}

func (p *Parser) comparison() Expression {
	expr := p.term()

	for p.match(GREATER, GREATER_EQUAL, LESS, LESS_EQUAL) {
		operator := p.previous()
		right := p.term()
		expr = &Binary{
			Token:    p.peek(),
			Left:     expr,
			Operator: operator.Lexeme,
			Right:    right,
		}
	}

	return expr
}

func (p *Parser) equality() Expression {
	// matches equality or anything of higher precedence
	expr := p.comparison()

	for p.match(BANG_EQUAL, EQUAL_EQUAL) {
		operator := p.previous()
		right := p.comparison()
		expr = &Binary{
			Token:    p.peek(),
			Left:     expr,
			Operator: operator.Lexeme,
			Right:    right,
		}
	}

	return expr
}

func (p *Parser) assignment() Expression {
	expr := p.ternary()

	for p.match(EQUAL) {
		equals := p.previous()
		right := p.assignment()

		if identifier, ok := expr.(*Identifier); ok {
			return &Assignment{
				Token:      equals,
				Identifier: *identifier,
				Right:      right,
			}
		}
	}

	return expr
}

func (p *Parser) ternary() Expression {
	expression := p.equality()

	if p.match(QUESTION) {
		thenBranch := p.expression()
		if !p.consume(COLON) {
			return expression
		}
		elseBranch := p.ternary()
		expression = &TernaryExpression{Condition: expression, ThenBranch: thenBranch, ElseBranch: elseBranch}
	}

	return expression
}

func (p *Parser) comma() Expression {
	expr := p.assignment()

	for p.match(COMMA) {
		right := p.assignment()
		expr = &CommaExpression{Expression: []Expression{expr, right}}
	}

	return expr
}

func (p *Parser) expression() Expression {
	return p.comma()
}
