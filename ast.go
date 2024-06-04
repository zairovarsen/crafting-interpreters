package main

import (
	"fmt"
	"strings"
)

type Node interface {
	String() string
	TokenLiteral() string
	Accept(visitor Visitor) Object
}

// Expression
type Expression interface {
	Node
	expressionNode()
}

// Statement
type Statement interface {
	Node
	statementNode()
}

type Program struct {
	Statements []Statement
}

func (p *Program) String() string {
	var str strings.Builder

	for _, stmt := range p.Statements {
		str.WriteString(stmt.String())
	}

	return str.String()
}

func (p *Program) TokenLiteral() string {
	if len(p.Statements) > 1 {
		return p.Statements[0].TokenLiteral()
	}
	return ""
}

// either returns nil object or error object
func (p *Program) Accept(visitor Visitor) Object {
	return visitor.VisitProgram(p)
}

// Everything other than return statement
type ExpressionStatement struct {
	Token      Token
	Expression Expression
}

func (es *ExpressionStatement) statementNode() {}
func (es *ExpressionStatement) String() string {
	return es.Expression.String()
}
func (es *ExpressionStatement) TokenLiteral() string {
	return es.Token.Lexeme
}
func (es *ExpressionStatement) Accept(visitor Visitor) Object {
	return visitor.VisitExpressionStatement(es)
}

type BlockStatement struct {
	Token      Token
	Statements []Statement
}

func (b *BlockStatement) String() string {
	var str strings.Builder

	for _, stmt := range b.Statements {
		str.WriteString(stmt.String())
	}

	return str.String()
}
func (b *BlockStatement) TokenLiteral() string {
	return b.Token.Lexeme
}
func (b *BlockStatement) Accept(visitor Visitor) Object {
	return visitor.VisitBlockStatement(b)
}

type ReturnStatement struct {
	Token       Token
	ReturnValue Expression
}

func (r *ReturnStatement) statementNode() {}
func (r *ReturnStatement) TokenLiteral() string {
	return r.Token.Lexeme
}
func (r *ReturnStatement) String() string {
	var str strings.Builder

	str.WriteString(r.TokenLiteral() + " ")

	if r.ReturnValue != nil {
		str.WriteString(r.ReturnValue.String())
	}

	return str.String()
}
func (r *ReturnStatement) Accept(visitor Visitor) Object {
	return visitor.VisitReturnStatement(r)
}

type CommaExpression struct {
	Token      Token
	Expression []Expression
}

func (ce *CommaExpression) expressionNode() {}
func (ce *CommaExpression) TokenLiteral() string {
	return ce.Token.Lexeme
}
func (ce *CommaExpression) String() string {
	var str strings.Builder

	for i, e := range ce.Expression {
		str.WriteString(e.String())
		if i != len(ce.Expression)-1 {
			str.WriteString(" , ")
		}
	}

	return str.String()
}
func (ce *CommaExpression) Accept(visitor Visitor) Object {
	return visitor.VisitCommaExpression(ce)
}

type TernaryExpression struct {
	Token      Token
	Condition  Expression
	ThenBranch Expression
	ElseBranch Expression
}

func (te *TernaryExpression) expressionNode() {}
func (te *TernaryExpression) TokenLiteral() string {
	return te.Token.Lexeme
}
func (te *TernaryExpression) String() string {
	var str strings.Builder

	str.WriteString(te.Condition.String())
	str.WriteString(" ? ")
	str.WriteString(te.ThenBranch.String())
	str.WriteString(" : ")
	str.WriteString(te.ElseBranch.String())

	return str.String()
}
func (te *TernaryExpression) Accept(visitor Visitor) Object {
	return visitor.VisitTernaryExpression(te)
}

type Assignment struct {
	Token      Token
	Identifier Identifier
	Right      Expression
}

func (a *Assignment) expressionNode() {}
func (a *Assignment) TokenLiteral() string {
	return a.Token.Lexeme
}
func (a *Assignment) String() string {
	var str strings.Builder

	str.WriteString(a.Identifier.String())
	str.WriteString(" = ")
	str.WriteString(a.Right.String())

	return str.String()
}
func (a *Assignment) Accept(visitor Visitor) Object {
	return visitor.VisitAssignment(a)
}

type Identifier struct {
	Token Token
	Value string
}

func (i *Identifier) expressionNode() {}
func (i *Identifier) TokenLiteral() string {
	return i.Token.Lexeme
}
func (i *Identifier) String() string { return i.Value }
func (i *Identifier) Accept(visitor Visitor) Object {
	return visitor.VisitIdentifier(i)
}

type GroupedExpression struct {
	Token      Token
	Expression Expression
}

func (g *GroupedExpression) expressionNode() {}
func (g *GroupedExpression) TokenLiteral() string {
	return g.Token.Lexeme
}
func (g *GroupedExpression) Accept(visitor Visitor) Object {
	return visitor.VisitGroupedExpression(g)
}
func (g *GroupedExpression) String() string {
	var str strings.Builder
	str.WriteString("( ")
	str.WriteString(g.Expression.String())
	str.WriteString(" )")
	return str.String()
}

type BooleanLiteral struct {
	Token Token
	Value bool
}

func (bl *BooleanLiteral) expressionNode() {}
func (bl *BooleanLiteral) TokenLiteral() string {
	return bl.Token.Lexeme
}
func (bl *BooleanLiteral) Accept(visitor Visitor) Object {
	return visitor.VisitBooleanLiteral(bl)
}
func (bl *BooleanLiteral) String() string {
	return fmt.Sprintf("%t", bl.Value)
}

type NilLiteral struct {
	Token Token
}

func (nl *NilLiteral) expressionNode() {}
func (nl *NilLiteral) TokenLiteral() string {
	return nl.Token.Lexeme
}
func (nl *NilLiteral) String() string {
	return "nil"
}
func (nl *NilLiteral) Accept(visitor Visitor) Object {
	return visitor.VisitNilLiteral(nl)
}

type NumberLiteral struct {
	Token Token
	Value float64
}

func (nl *NumberLiteral) expressionNode() {}
func (nl *NumberLiteral) TokenLiteral() string {
	return nl.Token.Lexeme
}
func (nl *NumberLiteral) Accept(visitor Visitor) Object {
	return visitor.VisitNumberLiteral(nl)
}
func (nl *NumberLiteral) String() string {
	return fmt.Sprintf("%g", nl.Value)
}

type StringLiteral struct {
	Token Token
	Value string
}

func (sl *StringLiteral) expressionNode() {}
func (sl *StringLiteral) TokenLiteral() string {
	return sl.Token.Lexeme
}
func (sl *StringLiteral) Accept(visitor Visitor) Object {
	return visitor.VisitStringLiteral(sl)
}
func (sl *StringLiteral) String() string {
	return fmt.Sprintf("%q", sl.Value)
}

type Unary struct {
	Token    Token
	Operator string
	Right    Expression
}

func (u *Unary) TokenLiteral() string {
	return u.Token.Lexeme
}
func (u *Unary) expressionNode() {}
func (u *Unary) String() string {
	var str strings.Builder

	str.WriteString("(")
	str.WriteString(u.Operator)
	str.WriteString(u.Right.String())
	str.WriteString(")")

	return str.String()
}
func (u *Unary) Accept(visitor Visitor) Object {
	return visitor.VisitUnary(u)
}

type Binary struct {
	Token    Token // operator token
	Left     Expression
	Operator string
	Right    Expression
}

func (b *Binary) TokenLiteral() string {
	return b.Token.Lexeme
}
func (b *Binary) expressionNode() {}
func (b *Binary) String() string {
	var str strings.Builder

	str.WriteString("(")
	str.WriteString(b.Left.String())
	str.WriteString(" " + b.Operator + " ")
	str.WriteString(b.Right.String())
	str.WriteString(")")

	return str.String()
}
func (b *Binary) Accept(visitor Visitor) Object {
	return visitor.VisitBinary(b)
}
