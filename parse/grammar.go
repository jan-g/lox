package parse

import (
	"fmt"
	"github.com/jan-g/lox/ast"
	"github.com/jan-g/lox/lex"
	"strconv"
)

func (p *parser) Parse() (e ast.Stmt, err error) {
	defer func() {
		if e := recover(); e != nil {
			err = e.(error)
		}
	}()
	e = p.Program()
	if p.Eof() {
		return e, nil
	}
	return nil, fmt.Errorf("unexpected token: %s %s", p.Peek().Lexeme, p.Peek().Start)
}

func (p *parser) Program() ast.Stmt {
	sts := []ast.Stmt{}
	for !p.Eof() {
		s := p.Decl()
		if s == nil {
			break
		}
		sts = append(sts, s)
	}
	return ast.ProgStmt(sts...)
}

func (p *parser) Decl() ast.Stmt {
	if p.Match(lex.TokKW, "var") {
		return p.DeclStmt()
	}
	return p.Stmt()
}

func (p *parser) DeclStmt() ast.Stmt {
	name := p.Consume("variable name expected", lex.TokId).Lexeme
	var init ast.Expr = ast.Nil
	if p.Match(lex.TokOp, "=") {
		init = p.Expr()
	}
	p.Consume("expect ';' after declaration", lex.TokPunc, ";")
	return ast.Decl(name, init)
}

func (p *parser) Stmt() ast.Stmt {
	if p.Match(lex.TokKW, "print") {
		return p.PrintStmt()
	}
	if p.Match(lex.TokPunc, "{") {
		return p.Block()
	}
	return p.ExprStmt()
}

func (p *parser) Block() ast.Stmt {
	sts := []ast.Stmt{}
	for !p.Check(lex.TokPunc, "}") && !p.Eof() {
		sts = append(sts, p.Decl())
	}
	p.Consume("block must close with '}'", lex.TokPunc, "}")
	return ast.BlockStmt(sts...)
}

func (p *parser) PrintStmt() ast.Stmt {
	e := p.Expr()
	p.Consume("';' expected after value", lex.TokPunc, ";")
	return ast.PrintStmt(e)
}

func (p *parser) ExprStmt() ast.Stmt {
	e := p.Expr()
	p.Consume("';' expected after value", lex.TokPunc, ";")
	return ast.ExprStmt(e)
}

func (p *parser) Expr() ast.Expr {
	return p.Assign()
}

func (p *parser) Assign() ast.Expr {
	lhs := p.Equality()
	if p.Match(lex.TokOp, "=") {
		rhs := p.Assign()
		switch lhs := lhs.(type) {
		case ast.Var:
			return ast.Assignment(lhs, rhs)
		}
		panic(p.Error("assignment must have variable on the LHS"))
	}
	return lhs
}

func (p *parser) Equality() ast.Expr {
	e := p.Comparison()

	if p.Match(lex.TokOp, "==", "!=") {
		op := p.Previous()
		e = ast.Bin(e, op.Lexeme, p.Comparison())
	}

	return e
}

func (p *parser) Comparison() ast.Expr {
	e := p.Term()

	if p.Match(lex.TokOp, "<", "<=", ">", ">=") {
		op := p.Previous()
		e = ast.Bin(e, op.Lexeme, p.Term())
	}

	return e
}

func (p *parser) Term() ast.Expr {
	e := p.Factor()

	for p.Match(lex.TokOp, "+", "-") {
		op := p.Previous()
		r := p.Factor()
		e = ast.Bin(e, op.Lexeme, r)
	}

	return e
}

func (p *parser) Factor() ast.Expr {
	e := p.Unary()

	for p.Match(lex.TokOp, "*", "/") {
		op := p.Previous()
		r := p.Unary()
		e = ast.Bin(e, op.Lexeme, r)
	}

	return e
}

func (p *parser) Unary() ast.Expr {
	if p.Match(lex.TokOp, "-", "!") {
		op := p.Previous()
		return ast.Un(op.Lexeme, p.Unary())
	}

	return p.Primary()
}

func (p *parser) Primary() ast.Expr {
	if p.Match(lex.TokStr) {
		return ast.Str(p.Previous().Lexeme)
	}
	if p.Match(lex.TokNum) {
		n := p.Previous()
		v, err := strconv.ParseFloat(n.Lexeme, 64)
		if err == nil {
			return ast.Num(v)
		}
		panic(p.Error("Can't parse numeric value %s: %s", n.Lexeme, err))
	}
	if p.Match(lex.TokKW, "nil") {
		return ast.Nil
	}
	if p.Match(lex.TokPunc, "(") {
		e := p.Expr()
		p.Consume("expect ')' after expression", lex.TokPunc, ")")
		return e
	}
	if p.Match(lex.TokId) {
		return ast.Id(p.Previous().Lexeme)
	}
	panic(p.Error("expected: Primary"))
}
