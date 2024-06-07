package main

import (
	"fmt"
	"strings"
)

type Node interface {
	String() string
	TokenLiteral() string
	Accept(visitor Visitor, env *Environment, parent Statement) Object
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

func (p *Program) statementNode() {}

// either returns nil object or error object
func (p *Program) Accept(visitor Visitor, env *Environment, parent Statement) Object {
	return visitor.VisitProgram(p, env, parent)
}

type IfStatement struct {
	Token      Token
	Condition  Expression
	ThenBranch Statement
	ElseBranch Statement
}

func (is *IfStatement) statementNode() {}
func (is *IfStatement) String() string {
	var str strings.Builder

	str.WriteString(is.TokenLiteral())
	str.WriteString(" ( " + is.Condition.String() + " ) {\n")
	str.WriteString(is.ThenBranch.String() + "} ")

	if is.ElseBranch != nil {
		str.WriteString("else {\n")
		str.WriteString(is.ElseBranch.String() + "} ")
	}

	return str.String()
}
func (is *IfStatement) TokenLiteral() string {
	return is.Token.Lexeme
}
func (is *IfStatement) Accept(visitor Visitor, env *Environment, parent Statement) Object {
	return visitor.VisitIfStatement(is, env, parent)
}

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
func (es *ExpressionStatement) Accept(visitor Visitor, env *Environment, parent Statement) Object {
	return visitor.VisitExpressionStatement(es, env, parent)
}

type VarStatement struct {
	Token      Token
	Identifier *Identifier
	Expression Expression
}

func (vs *VarStatement) statementNode() {}
func (vs *VarStatement) String() string {
	var str strings.Builder

	str.WriteString(vs.TokenLiteral() + " ")
	str.WriteString(vs.Identifier.String() + " = ")

	if vs.Expression != nil {
		str.WriteString(vs.Expression.String())
	}

	return str.String()
}
func (vs *VarStatement) TokenLiteral() string {
	return vs.Token.Lexeme
}
func (vs *VarStatement) Accept(visitor Visitor, env *Environment, parent Statement) Object {
	return visitor.VisitVarStatement(vs, env, parent)
}

type BlockStatement struct {
	Token      Token
	Statements []Statement
}

func (b *BlockStatement) statementNode() {}
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
func (b *BlockStatement) Accept(visitor Visitor, env *Environment, parent Statement) Object {
	return visitor.VisitBlockStatement(b, env, parent)
}

type PrintStatement struct {
	Token      Token
	PrintValue Expression
}

func (p *PrintStatement) statementNode() {}
func (p *PrintStatement) TokenLiteral() string {
	return p.Token.Lexeme
}
func (p *PrintStatement) String() string {
	var str strings.Builder

	str.WriteString(p.TokenLiteral() + " ")

	if p.PrintValue != nil {
		str.WriteString(p.PrintValue.String())
	}

	return str.String()
}
func (p *PrintStatement) Accept(visitor Visitor, env *Environment, parent Statement) Object {
	return visitor.VisitPrintStatement(p, env, parent)
}

type ContinueStatement struct {
	Token Token
}

func (c *ContinueStatement) statementNode() {}
func (c *ContinueStatement) TokenLiteral() string {
	return c.Token.Lexeme
}
func (c *ContinueStatement) String() string {
	var str strings.Builder

	str.WriteString(c.TokenLiteral())

	return str.String()
}
func (c *ContinueStatement) Accept(visitor Visitor, env *Environment, parent Statement) Object {
	return visitor.VisitContinueStatement(c, env, parent)
}

type BreakStatement struct {
	Token Token
}

func (b *BreakStatement) statementNode() {}
func (b *BreakStatement) TokenLiteral() string {
	return b.Token.Lexeme
}
func (b *BreakStatement) String() string {
	var str strings.Builder

	str.WriteString(b.TokenLiteral())

	return str.String()
}
func (b *BreakStatement) Accept(visitor Visitor, env *Environment, parent Statement) Object {
	return visitor.VisitBreakStatement(b, env, parent)
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
func (r *ReturnStatement) Accept(visitor Visitor, env *Environment, parent Statement) Object {
	return visitor.VisitReturnStatement(r, env, parent)
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
func (ce *CommaExpression) Accept(visitor Visitor, env *Environment, parent Statement) Object {
	return visitor.VisitCommaExpression(ce, env, parent)
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
func (te *TernaryExpression) Accept(visitor Visitor, env *Environment, parent Statement) Object {
	return visitor.VisitTernaryExpression(te, env, parent)
}

type Assignment struct {
	Token      Token
	Identifier Identifier
	Expression Expression
}

func (a *Assignment) expressionNode() {}
func (a *Assignment) TokenLiteral() string {
	return a.Token.Lexeme
}
func (a *Assignment) String() string {
	var str strings.Builder

	str.WriteString(a.Identifier.String())
	str.WriteString(" = ")
	str.WriteString(a.Expression.String())

	return str.String()
}
func (a *Assignment) Accept(visitor Visitor, env *Environment, parent Statement) Object {
	return visitor.VisitAssignment(a, env, parent)
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
func (i *Identifier) Accept(visitor Visitor, env *Environment, parent Statement) Object {
	return visitor.VisitIdentifier(i, env, parent)
}

type GroupedExpression struct {
	Token      Token
	Expression Expression
}

func (g *GroupedExpression) expressionNode() {}
func (g *GroupedExpression) TokenLiteral() string {
	return g.Token.Lexeme
}
func (g *GroupedExpression) Accept(visitor Visitor, env *Environment, parent Statement) Object {
	return visitor.VisitGroupedExpression(g, env, parent)
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
func (bl *BooleanLiteral) Accept(visitor Visitor, env *Environment, parent Statement) Object {
	return visitor.VisitBooleanLiteral(bl, env, parent)
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
func (nl *NilLiteral) Accept(visitor Visitor, env *Environment, parent Statement) Object {
	return visitor.VisitNilLiteral(nl, env, parent)
}

type NumberLiteral struct {
	Token Token
	Value float64
}

func (nl *NumberLiteral) expressionNode() {}
func (nl *NumberLiteral) TokenLiteral() string {
	return nl.Token.Lexeme
}
func (nl *NumberLiteral) Accept(visitor Visitor, env *Environment, parent Statement) Object {
	return visitor.VisitNumberLiteral(nl, env, parent)
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
func (sl *StringLiteral) Accept(visitor Visitor, env *Environment, parent Statement) Object {
	return visitor.VisitStringLiteral(sl, env, parent)
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
func (u *Unary) Accept(visitor Visitor, env *Environment, parent Statement) Object {
	return visitor.VisitUnary(u, env, parent)
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
func (b *Binary) Accept(visitor Visitor, env *Environment, parent Statement) Object {
	return visitor.VisitBinary(b, env, parent)
}

type Logical struct {
	Token    Token // operator token
	Left     Expression
	Operator string
	Right    Expression
}

func (l *Logical) TokenLiteral() string {
	return l.Token.Lexeme
}
func (l *Logical) expressionNode() {}
func (l *Logical) String() string {
	var str strings.Builder

	str.WriteString("(")
	str.WriteString(l.Left.String())
	str.WriteString(" " + l.Operator + " ")
	str.WriteString(l.Right.String())
	str.WriteString(")")

	return str.String()
}
func (l *Logical) Accept(visitor Visitor, env *Environment, parent Statement) Object {
	return visitor.VisitLogical(l, env, parent)
}

type For struct {
	Token       Token
	Initializer Statement
	Condition   Expression
	Increment   Expression
	Body        Statement
}

func (f *For) TokenLiteral() string {
	return f.Token.Lexeme
}
func (f *For) statementNode() {}
func (f *For) String() string {
	var str strings.Builder

	str.WriteString(f.TokenLiteral())
	str.WriteString("(")
	str.WriteString(f.Initializer.String())
	str.WriteString(";")
	str.WriteString(f.Condition.String())
	str.WriteString(";")
	str.WriteString(f.Increment.String())
	str.WriteString(")")
	str.WriteString(f.Body.String())

	return str.String()
}
func (f *For) Accept(visitor Visitor, env *Environment, parent Statement) Object {
	return visitor.VisitForStatement(f, env, parent)
}

type While struct {
	Token     Token
	Condition Expression
	Body      Statement
}

func (w *While) TokenLiteral() string {
	return w.Token.Lexeme
}
func (w *While) statementNode() {}
func (w *While) String() string {
	var str strings.Builder

	str.WriteString(w.TokenLiteral())
	str.WriteString("(")
	str.WriteString(w.Condition.String())
	str.WriteString(")")
	str.WriteString(w.Body.String())

	return str.String()
}
func (w *While) Accept(visitor Visitor, env *Environment, parent Statement) Object {
	return visitor.VisitWhileStatement(w, env, parent)
}
