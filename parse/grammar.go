package parse

import (
	"fmt"
	"github.com/jan-g/lox/ast"
	"github.com/jan-g/lox/lex"
	"strconv"
)

func (p *parser) Parse() (e ast.Expr, err error) {
	defer func() {
		if e := recover(); e != nil {
			err = e.(error)
		}
	}()
	e = p.Expr()
	if p.Eof() {
		return e, nil
	}
	return nil, fmt.Errorf("unexpected token: %s %s", p.Peek().Lexeme, p.Peek().Start)
}

func (p *parser) Expr() ast.Expr {
	return p.Equality()
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
	if p.Match(lex.TokPunc, "(") {
		e := p.Expr()
		p.Consume(lex.TokPunc, ")", "expect ')' after expression")
		return e
	}
	panic(p.Error("expected: Primary"))
}
