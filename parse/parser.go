package parse

import (
	"fmt"
	"github.com/jan-g/lox/ast"
	"github.com/jan-g/lox/lex"
	"io"
)

type Parser interface {
	Parse() (ast.Stmt, error)
	Expr() ast.Expr
}

type parser struct {
	l    *lex.Lexer
	prev lex.T
}

func New(r io.Reader) Parser {
	return &parser{
		l: lex.New(r, lex.MakeSwitch(lex.MakeId(lex.Kws...), lex.WS, lex.Op, lex.Num, lex.Str)),
	}
}

func (p *parser) Peek() lex.T {
	return p.l.Current()
}

func (p *parser) Next() lex.T {
	p.prev = p.Peek()
	return p.l.Scan()
}

func (p *parser) Previous() lex.T {
	return p.prev
}

func (p *parser) Check(t lex.TokenType, ls ...string) bool {
	c := p.Peek()
	if c.Token != t {
		return false
	}
	if len(ls) == 0 {
		return true
	}
	for _, m := range ls {
		if m == c.Lexeme {
			return true
		}
	}
	return false
}

func (p *parser) Match(t lex.TokenType, ls ...string) bool {
	if p.Check(t, ls...) {
		p.Next()
		return true
	}
	return false
}

func (p *parser) Advance() lex.T {
	p.Next()
	return p.Previous()
}

func (p *parser) Eof() bool {
	return p.Check(lex.TokEof)
}

func (p *parser) Consume(msg string, t lex.TokenType, lexeme ...string) lex.T {
	if p.Check(t, lexeme...) {
		curr := p.Peek()
		p.Next()
		return curr
	}
	panic(p.Error(msg))
}

func (p *parser) Error(msg string, xs ...interface{}) error {
	if p.Eof() {
		return fmt.Errorf("%s AT EOF", fmt.Sprintf(msg, xs...))
	}
	return fmt.Errorf("%s %s", fmt.Sprintf(msg, xs...), p.Peek().Start)
}

func (p *parser) accept(t lex.TokenType) (lex.T, bool) {
	c := p.l.Current()
	if c.Token == t {
		p.l.Scan()
		return c, true
	}
	return lex.T{}, false
}

func (p *parser) match(t lex.TokenType) bool {
	return p.l.Current().Token == t
}

func (p *parser) match2(t lex.TokenType, lexeme string) bool {
	return p.l.Current().Token == t && p.l.Current().Lexeme == lexeme
}

func (p *parser) accept2(t lex.TokenType, lexeme string) bool {
	c := p.l.Current()
	if c.Token == t && c.Lexeme == lexeme {
		p.l.Scan()
		return true
	}
	return false
}
