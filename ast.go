package main

import (
	"fmt"
	"strings"
)

type Node interface {
	String() string
	TokenLiteral() string
	Accept(visitor Visitor, env *Environment) Object
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
func (p *Program) Accept(visitor Visitor, env *Environment) Object {
	return visitor.VisitProgram(p, env)
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
	str.WriteString(LEFT_PAREN + is.Condition.String() + RIGHT_PAREN + " " + LEFT_BRACKET + "\n")
	str.WriteString(is.ThenBranch.String() + RIGHT_BRACKET + " ")

	if is.ElseBranch != nil {
		str.WriteString("else {\n")
		str.WriteString(is.ElseBranch.String() + "} ")
	}

	return str.String()
}
func (is *IfStatement) TokenLiteral() string {
	return is.Token.Lexeme
}
func (is *IfStatement) Accept(visitor Visitor, env *Environment) Object {
	return visitor.VisitIfStatement(is, env)
}

type ClassStatement struct {
	Token   Token // class
	Name    *Identifier
	Methods []*MethodDeclaration
}

func (cs *ClassStatement) statementNode() {}
func (cs *ClassStatement) TokenLiteral() string {
	return cs.Token.Lexeme
}
func (cs *ClassStatement) String() string {
	var str strings.Builder

	var methods []string
	for _, m := range cs.Methods {
		methods = append(methods, m.String())
	}

	str.WriteString(cs.TokenLiteral())
	str.WriteString(cs.Name.String())
	str.WriteString(strings.Join(methods, "\n"))

	return str.String()
}
func (cs *ClassStatement) Accept(visitor Visitor, env *Environment) Object {
	return visitor.VisitClassStatement(cs, env)
}

type FunctionCommon struct {
	Token  Token // ( token
	Name   *Identifier
	Params []*Identifier
	Body   *BlockStatement
}

func (fc *FunctionCommon) TokenLiteral() string {
	return fc.Token.Lexeme
}

func (fc *FunctionCommon) String() string {
	var str strings.Builder

	var params []string
	for _, p := range fc.Params {
		params = append(params, p.String())
	}

	str.WriteString(fc.TokenLiteral())
	if fc.Name != nil {
		str.WriteString(fc.Name.String())
	}
	str.WriteString(LEFT_PAREN)
	str.WriteString(strings.Join(params, COMMA+" "))
	str.WriteString(RIGHT_PAREN)
	str.WriteString(fc.Body.String())

	return str.String()
}

type FunctionDeclaration struct {
	FunctionCommon
}

func (fl *FunctionDeclaration) statementNode() {}
func (fl *FunctionDeclaration) Accept(visitor Visitor, env *Environment) Object {
	return visitor.VisitFunctionDeclaration(fl, env)
}

type FunctionLiteral struct {
	FunctionCommon
}

func (fl *FunctionLiteral) expressionNode() {}
func (fl *FunctionLiteral) Accept(visitor Visitor, env *Environment) Object {
	return visitor.VisitFunctionLiteral(fl, env)
}

type MethodDeclaration struct {
	FunctionCommon
	// Receiver *Identifier //
}

func (md *MethodDeclaration) statementNode() {}
func (md *MethodDeclaration) Accept(visitor Visitor, env *Environment) Object {
	return visitor.VisitMethodDeclaration(md, env)
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
func (es *ExpressionStatement) Accept(visitor Visitor, env *Environment) Object {
	return visitor.VisitExpressionStatement(es, env)
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
func (vs *VarStatement) Accept(visitor Visitor, env *Environment) Object {
	return visitor.VisitVarStatement(vs, env)
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
func (b *BlockStatement) Accept(visitor Visitor, env *Environment) Object {
	return visitor.VisitBlockStatement(b, env)
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
func (c *ContinueStatement) Accept(visitor Visitor, env *Environment) Object {
	return visitor.VisitContinueStatement(c, env)
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
func (b *BreakStatement) Accept(visitor Visitor, env *Environment) Object {
	return visitor.VisitBreakStatement(b, env)
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
func (r *ReturnStatement) Accept(visitor Visitor, env *Environment) Object {
	return visitor.VisitReturnStatement(r, env)
}

type GetExpression struct {
	Token    Token
	Object   Expression
	Property *Identifier
}

func (ge *GetExpression) expressionNode()      {}
func (ge *GetExpression) TokenLiteral() string { return ge.Token.Lexeme }
func (ge *GetExpression) String() string {
	var str strings.Builder

	str.WriteString(ge.Object.String())
	str.WriteString(ge.Property.String())

	return str.String()
}
func (ge *GetExpression) Accept(visitor Visitor, env *Environment) Object {
	return visitor.VisitGetExpression(ge, env)
}

type CallExpression struct {
	Token     Token // '(' token
	Callee    Expression
	Arguments []Expression
}

func (ce *CallExpression) expressionNode()      {}
func (ce *CallExpression) TokenLiteral() string { return ce.Token.Lexeme }
func (ce *CallExpression) String() string {
	var str strings.Builder

	var args []string
	for _, a := range ce.Arguments {
		args = append(args, a.String())
	}

	str.WriteString(ce.Callee.String())
	str.WriteString(LEFT_PAREN)
	str.WriteString(strings.Join(args, COMMA+" "))
	str.WriteString(RIGHT_PAREN)

	return str.String()
}
func (ce *CallExpression) Accept(visitor Visitor, env *Environment) Object {
	return visitor.VisitCallExpression(ce, env)
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
func (te *TernaryExpression) Accept(visitor Visitor, env *Environment) Object {
	return visitor.VisitTernaryExpression(te, env)
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
func (a *Assignment) Accept(visitor Visitor, env *Environment) Object {
	return visitor.VisitAssignment(a, env)
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
func (i *Identifier) Accept(visitor Visitor, env *Environment) Object {
	return visitor.VisitIdentifier(i, env)
}

type GroupedExpression struct {
	Token      Token
	Expression Expression
}

func (g *GroupedExpression) expressionNode() {}
func (g *GroupedExpression) TokenLiteral() string {
	return g.Token.Lexeme
}
func (g *GroupedExpression) Accept(visitor Visitor, env *Environment) Object {
	return visitor.VisitGroupedExpression(g, env)
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
func (bl *BooleanLiteral) Accept(visitor Visitor, env *Environment) Object {
	return visitor.VisitBooleanLiteral(bl, env)
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
func (nl *NilLiteral) Accept(visitor Visitor, env *Environment) Object {
	return visitor.VisitNilLiteral(nl, env)
}

type NumberLiteral struct {
	Token Token
	Value float64
}

func (nl *NumberLiteral) expressionNode() {}
func (nl *NumberLiteral) TokenLiteral() string {
	return nl.Token.Lexeme
}
func (nl *NumberLiteral) Accept(visitor Visitor, env *Environment) Object {
	return visitor.VisitNumberLiteral(nl, env)
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
func (sl *StringLiteral) Accept(visitor Visitor, env *Environment) Object {
	return visitor.VisitStringLiteral(sl, env)
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
func (u *Unary) Accept(visitor Visitor, env *Environment) Object {
	return visitor.VisitUnary(u, env)
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
func (b *Binary) Accept(visitor Visitor, env *Environment) Object {
	return visitor.VisitBinary(b, env)
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
func (l *Logical) Accept(visitor Visitor, env *Environment) Object {
	return visitor.VisitLogical(l, env)
}

type For struct {
	Token       Token
	Initializer Statement
	Condition   Expression
	Increment   Expression
	Body        *BlockStatement
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
func (f *For) Accept(visitor Visitor, env *Environment) Object {
	return visitor.VisitForStatement(f, env)
}

type While struct {
	Token     Token
	Condition Expression
	Body      *BlockStatement
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
func (w *While) Accept(visitor Visitor, env *Environment) Object {
	return visitor.VisitWhileStatement(w, env)
}
