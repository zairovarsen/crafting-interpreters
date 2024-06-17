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

	for !p.isAtEnd() {
		stmt := p.declaration()
		if stmt != nil {
			program.Statements = append(program.Statements, stmt)
		} else {
			// try to return back to some normal usable state
			p.synchronize()
		}
	}

	return program
}

func (p *Parser) declaration() Statement {
	switch p.peek().Type {
	case CLASS:
		return p.classDeclaration()
	case FUNCTION:
		return p.functionDeclaration()
	case VAR:
		return p.varDeclaration()
	default:
		return p.statement()
	}
}

func (p *Parser) statement() Statement {
	switch p.peek().Type {
	case CONTINUE:
		return p.continueStatement()
	case BREAK:
		return p.breakStatement()
	case FOR:
		return p.forStatement()
	case WHILE:
		return p.whileStatement()
	case IF:
		return p.ifStatement()
	case RETURN:
		return p.returnStatement()
	default:
		return p.expressionStatement()
	}
}

func (p *Parser) classDeclaration() *ClassStatement {
	class := &ClassStatement{Token: p.advance()}

	if !p.expectPeek(IDENTIFIER) {
		return nil
	}

	class.Name = &Identifier{Token: p.previous(), Value: p.previous().Lexeme}

	if p.check(EXTEND) {
		p.advance()
		if !p.expectPeek(IDENTIFIER) {
			return nil
		}
		if class.Name.Value == p.previous().Lexeme {
			err := &Error{Message: fmt.Sprintf("Class %s cannot inherit from itself", class.Name.Value), Line: p.peek().Line}
			p.addError(err)
			return nil
		}
		class.SuperClass = &Identifier{Token: p.previous(), Value: p.previous().Lexeme}
	}

	if !p.expectPeek(LEFT_BRACKET) {
		return nil
	}

	for !p.isAtEnd() && !p.check(RIGHT_BRACKET) {
		method := p.parseMethodDeclaration()
		if method != nil {
			class.Methods = append(class.Methods, method)
			continue
		}
		return nil
	}

	if !p.expectPeek(RIGHT_BRACKET) {
		return nil
	}

	return class
}

func (p *Parser) functionDeclaration() *FunctionDeclaration {
	fun := &FunctionDeclaration{}
	fun.Token = p.advance()

	if !p.expectPeek(IDENTIFIER) {
		return nil
	}

	fun.Name = &Identifier{Token: p.previous(), Value: p.previous().Lexeme}

	if !p.expectPeek(LEFT_PAREN) {
		return nil
	}

	fun.Params = p.parseFunctionParams()
	fun.Body = p.block()

	return fun
}

func (p *Parser) breakStatement() *BreakStatement {
	stmt := &BreakStatement{Token: p.advance()}

	if !p.expectPeek(SEMICOLON) {
		return nil
	}

	return stmt
}

func (p *Parser) continueStatement() *ContinueStatement {
	stmt := &ContinueStatement{Token: p.advance()}

	if !p.expectPeek(SEMICOLON) {
		return nil
	}

	return stmt
}

func (p *Parser) forStatement() *For {
	stmt := &For{Token: p.advance()}

	if !p.expectPeek(LEFT_PAREN) {
		return nil
	}

	switch p.peek().Type {
	case SEMICOLON:
		stmt.Initializer = nil
		p.advance()
	case VAR:
		stmt.Initializer = p.varDeclaration()
	default:
		stmt.Initializer = p.expressionStatement()
	}

	if !p.match(SEMICOLON) {
		stmt.Condition = p.expression()
	}
	if !p.expectPeek(SEMICOLON) {
		return nil
	}

	if !p.match(RIGHT_PAREN) {
		stmt.Increment = p.expression()
	}
	if !p.expectPeek(RIGHT_PAREN) {
		return nil
	}
	stmt.Body = p.block()

	return stmt
}

func (p *Parser) whileStatement() *While {
	stmt := &While{Token: p.advance()}
	if !p.expectPeek(LEFT_PAREN) {
		return nil
	}
	stmt.Condition = p.expression()
	if !p.expectPeek(RIGHT_PAREN) {
		return nil
	}
	stmt.Body = p.block()

	return stmt
}

func (p *Parser) ifStatement() *IfStatement {
	stmt := &IfStatement{Token: p.advance()}
	if !p.expectPeek(LEFT_PAREN) {
		return nil
	}
	stmt.Condition = p.expression()

	if !p.expectPeek(RIGHT_PAREN) {
		return nil
	}

	stmt.ThenBranch = p.block()
	if p.match(ELSE) {
		stmt.ElseBranch = p.block()
	}

	return stmt
}

func (p *Parser) block() *BlockStatement {
	blockStmt := &BlockStatement{Token: p.advance()}

	for !p.check(RIGHT_BRACKET) && !p.isAtEnd() {
		stmt := p.declaration()
		if stmt != nil {
			blockStmt.Statements = append(blockStmt.Statements, stmt)
		}
	}

	if !p.expectPeek(RIGHT_BRACKET) {
		return nil
	}

	return blockStmt
}

func (p *Parser) varDeclaration() *VarStatement {
	stmt := &VarStatement{Token: p.advance()}

	if !p.expectPeek(IDENTIFIER) {
		return nil
	}

	stmt.Identifier = &Identifier{Token: p.previous(), Value: p.previous().Lexeme}

	if !p.check(EQUAL) {
		// nil
		if p.expectPeek(SEMICOLON) {
			stmt.Expression = &NilLiteral{}
			return stmt
		}
		return nil
	}
	p.advance()

	stmt.Expression = p.expression()

	if !p.expectPeek(SEMICOLON) {
		return nil
	}
	return stmt
}

func (p *Parser) returnStatement() *ReturnStatement {
	stmt := &ReturnStatement{Token: p.advance()}

	if p.check(SEMICOLON) {
		p.advance()
		stmt.ReturnValue = nil
		return stmt
	}

	stmt.ReturnValue = p.expression()

	if !p.expectPeek(SEMICOLON) {
		return nil
	}
	return stmt
}

func (p *Parser) expressionStatement() Statement {
	stmt := &ExpressionStatement{Token: p.peek()}

	stmt.Expression = p.expression()

	if !p.expectPeek(SEMICOLON) {
		return nil
	}
	return stmt
}

func (p *Parser) expectPeek(tokenType TokenType) bool {
	if !p.check(tokenType) {
		err := &Error{Message: fmt.Sprintf("Expect peek to be %s, got=%s", tokenType, p.peek().Type), Line: p.peek().Line}
		p.addError(err)
		return false
	}

	p.advance()
	return true
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
		case CLASS, FUNCTION, VAR, FOR, IF, WHILE, RETURN:
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
		if p.expectPeek(RIGHT_PAREN) {
			return &GroupedExpression{Token: p.previous(), Expression: expr}
		}
	}
	if p.match(LEFT_BRACKET) {
		hash := &HashLiteral{Token: p.previous()}
		hash.Pairs = make(map[Expression]Expression)

		for !p.isAtEnd() && !p.check(RIGHT_BRACKET) {
			key := p.expression()

			if !p.expectPeek(COLON) {
				return nil
			}

			value := p.expression()
			hash.Pairs[key] = value

			if !p.check(RIGHT_BRACKET) && !p.check(COMMA) {
				return nil
			}
			if p.check(COMMA) {
				p.advance()
			}
		}

		if !p.expectPeek(RIGHT_BRACKET) {
			return nil
		}

		return hash
	}
	if p.match(LEFT_BRACE) {
		array := &ArrayLiteral{Token: p.previous()}
		array.Elements = p.parseExpressionList(RIGHT_BRACE)
		return array
	}
	if p.match(IDENTIFIER) {
		return &Identifier{
			Token: p.previous(),
			Value: p.previous().Lexeme,
		}
	}
	if p.match(FUNCTION) {
		return p.parseFunctionLiteral()
	}
	if p.match(THIS) {
		return &This{Token: p.previous()}
	}
	if p.match(SUPER) {
		super := &Super{Token: p.previous()}
		if !p.expectPeek(DOT) {
			return nil
		}
		if !p.expectPeek(IDENTIFIER) {
			return nil
		}
		super.Method = &Identifier{Token: p.previous(), Value: p.previous().Lexeme}
		return super
	}

	p.addError(&Error{Message: "Expect expression.", Line: p.previous().Line})
	return nil
}

func (p *Parser) parseMethodDeclaration() *MethodDeclaration {
	method := &MethodDeclaration{IsStatic: false, IsGetter: false}

	if p.check(STATIC) {
		p.advance()
		method.IsStatic = true
	}

	if !p.expectPeek(IDENTIFIER) {
		return nil
	}

	method.Token = p.previous()
	method.Name = &Identifier{Token: p.previous(), Value: p.previous().Lexeme}

	switch p.peek().Type {
	case LEFT_BRACKET:
		method.IsGetter = true
		method.Body = p.block()
	case LEFT_PAREN:
		p.advance()
		method.Params = p.parseFunctionParams()
		method.Body = p.block()
	default:
		err := &Error{Token: p.peek(), Message: "Invalid method declaration", Line: p.peek().Line}
		p.addError(err)
		return nil
	}

	return method
}

func (p *Parser) parseFunctionLiteral() *FunctionLiteral {
	fun := &FunctionLiteral{}
	fun.Token = p.previous()

	if p.check(IDENTIFIER) {
		fun.Name = &Identifier{Token: p.previous(), Value: p.previous().Lexeme}
	}

	if !p.expectPeek(LEFT_PAREN) {
		return nil
	}

	fun.Params = p.parseFunctionParams()
	fun.Body = p.block()

	return fun
}

func (p *Parser) parseExpressionList(end TokenType) []Expression {
	var list []Expression

	if p.peek().Type == end {
		p.advance()
		return list
	}

	list = append(list, p.expression())

	for p.match(COMMA) {
		list = append(list, p.expression())
		if len(list) >= 255 {
			err := &Error{Token: p.peek(), Message: "Can't have more than 255 arguments", Line: p.peek().Line}
			p.addError(err)
			return nil
		}
	}

	if !p.expectPeek(end) {
		return nil
	}

	return list
}

func (p *Parser) parseFunctionParams() []*Identifier {
	var identifiers []*Identifier

	if p.check(RIGHT_PAREN) {
		p.advance()
		return identifiers
	}

	if !p.expectPeek(IDENTIFIER) {
		return nil
	}

	identifier := &Identifier{Token: p.previous(), Value: p.previous().Lexeme}
	identifiers = append(identifiers, identifier)

	for p.match(COMMA) {
		if len(identifiers) >= 255 {
			p.addError(&Error{Message: "Can't have more than 255 parameters", Line: p.peek().Line})
			return nil
		}
		if !p.expectPeek(IDENTIFIER) {
			return nil
		}
		identifier := &Identifier{Token: p.previous(), Value: p.previous().Lexeme}
		identifiers = append(identifiers, identifier)
	}

	if !p.expectPeek(RIGHT_PAREN) {
		return nil
	}

	return identifiers
}

func (p *Parser) call() Expression {
	expr := p.primary()

	for {
		if p.match(LEFT_BRACE) {
			index := &IndexExpression{Left: expr, Token: p.previous()}
			index.Index = p.expression()
			if !p.expectPeek(RIGHT_BRACE) {
				return nil
			}
			return index
		}
		if p.match(LEFT_PAREN) {
			operator := p.previous()
			exp := &CallExpression{Token: operator, Callee: expr}
			exp.Arguments = p.parseExpressionList(RIGHT_PAREN)
			expr = exp
		} else if p.match(DOT) {
			operator := p.previous()
			if !p.expectPeek(IDENTIFIER) {
				return nil
			}
			identifier := &Identifier{Token: p.previous(), Value: p.previous().Lexeme}
			exp := &GetExpression{Token: operator, Object: expr, Property: identifier}
			expr = exp
		} else {
			break
		}
	}

	return expr
}

func (p *Parser) unary() Expression {
	for p.match(BANG, MINUS) {
		operator := p.previous()
		right := p.unary()
		return &Unary{
			Token:    operator,
			Operator: operator.Lexeme,
			Right:    right,
		}
	}

	return p.call()
}

func (p *Parser) factor() Expression {
	expr := p.unary()

	for p.match(SLASH, STAR) {
		operator := p.previous()
		right := p.unary()

		expr = &Binary{
			Token:    operator,
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
			Token:    operator,
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
			Token:    operator,
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
			Token:    operator,
			Left:     expr,
			Operator: operator.Lexeme,
			Right:    right,
		}
	}

	return expr
}

func (p *Parser) ternary() Expression {
	expr := p.equality()

	if p.match(QUESTION) {
		operator := p.previous()
		thenBranch := p.expression()
		if !p.expectPeek(COLON) {
			return expr
		}
		elseBranch := p.ternary()
		expr = &TernaryExpression{
			Token:      operator,
			Condition:  expr,
			ThenBranch: thenBranch,
			ElseBranch: elseBranch,
		}
	}

	return expr
}

func (p *Parser) and() Expression {
	expr := p.ternary()

	for p.match(AND) {
		operator := p.previous()
		right := p.ternary()
		expr = &Logical{
			Token:    operator,
			Left:     expr,
			Operator: operator.Lexeme,
			Right:    right,
		}
	}

	return expr
}

func (p *Parser) or() Expression {
	expr := p.and()

	for p.match(OR) {
		operator := p.previous()
		right := p.and()
		expr = &Logical{
			Token:    operator,
			Left:     expr,
			Operator: operator.Lexeme,
			Right:    right,
		}
	}

	return expr
}

func (p *Parser) assignment() Expression {
	expr := p.or()

	for p.match(EQUAL) {
		equals := p.previous()
		// chaining assignemnts x = y = z = 10
		right := p.assignment()

		switch target := expr.(type) {
		case *Identifier:
			return &Assignment{
				Token:      equals,
				Identifier: *target,
				Expression: right,
			}
		case *GetExpression:
			return &SetExpression{
				Token:    equals,
				Object:   target.Object,
				Property: target.Property,
				Value:    right,
			}
		default:
			p.addError(&Error{Message: "Invalid assignment target.", Line: p.previous().Line})
			return nil
		}
	}

	return expr
}

func (p *Parser) expression() Expression {
	return p.assignment()
}
