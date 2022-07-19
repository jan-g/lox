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
	if p.Match(lex.TokKW, "fun") {
		return p.FunDef()
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

func (p *parser) FunDef() ast.Stmt {
	name := p.Consume("function requires a name", lex.TokId).Lexeme
	fName := ast.Id(name)
	p.Consume("function definition expects '('", lex.TokPunc, "(")
	var params []ast.Var
	if !p.Check(lex.TokPunc, ")") {
		for {
			formal := p.Consume("formal parameter must be an identifier", lex.TokId)
			params = append(params, ast.Id(formal.Lexeme))
			if !p.Match(lex.TokPunc, ",") {
				break
			}
		}
	}
	p.Consume("formal parameters must end with ')'", lex.TokPunc, ")")
	p.Consume("function body must be a block", lex.TokPunc, "{")
	body := p.Block()
	return ast.FunStmt(fName, params, body)
}

func (p *parser) Stmt() ast.Stmt {
	if p.Match(lex.TokKW, "if") {
		return p.IfStmt()
	}
	if p.Match(lex.TokKW, "while") {
		return p.WhileStmt()
	}
	if p.Match(lex.TokKW, "for") {
		return p.ForStmt()
	}
	if p.Match(lex.TokKW, "print") {
		return p.PrintStmt()
	}
	if p.Match(lex.TokKW, "return") {
		return p.ReturnStmt()
	}
	if p.Match(lex.TokPunc, "{") {
		return p.Block()
	}
	return p.ExprStmt()
}

func (p *parser) IfStmt() ast.Stmt {
	p.Consume("if condition must be preceded by '('", lex.TokPunc, "(")
	cond := p.Expr()
	p.Consume("if condition must be followed by ')'", lex.TokPunc, ")")
	th := p.Stmt()
	var el ast.Stmt
	if p.Match(lex.TokKW, "else") {
		el = p.Stmt()
	}
	return ast.IfStmt(cond, th, el)
}

func (p *parser) WhileStmt() ast.Stmt {
	p.Consume("while condition must be preceded by '('", lex.TokPunc, "(")
	cond := p.Expr()
	p.Consume("while condition must be followed by ')'", lex.TokPunc, ")")
	body := p.Stmt()
	return ast.WhileStmt(cond, body)
}

func (p *parser) ForStmt() ast.Stmt {
	p.Consume("for condition must be followed by '('", lex.TokPunc, "(")

	var init ast.Stmt
	if p.Match(lex.TokPunc, ";") {
		// Nothing to do
	} else if p.Match(lex.TokKW, "var") {
		init = p.DeclStmt()
	} else {
		init = p.ExprStmt()
	}
	var cond ast.Expr
	if p.Check(lex.TokPunc, ";") {
		// Nothing to do
	} else {
		cond = p.Expr()
	}
	p.Consume("for condition must be followed by ';'", lex.TokPunc, ";")
	var incr ast.Expr
	if p.Check(lex.TokPunc, ")") {
		// Nothing to do
	} else {
		incr = p.Expr()
	}
	p.Consume("for increment must be followed by ')'", lex.TokPunc, ")")
	body := p.Stmt()

	// Desugar
	if incr != nil {
		body = ast.BlockStmt(body, ast.ExprStmt(incr))
	}
	if cond == nil {
		cond = ast.True
	}
	body = ast.WhileStmt(cond, body)
	if init != nil {
		body = ast.BlockStmt(init, body)
	}
	return body
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

func (p *parser) ReturnStmt() ast.Stmt {
	if p.Match(lex.TokPunc, ";") {
		return ast.ReturnStmt(nil)
	}
	e := p.Expr()
	p.Consume("return requires ';'", lex.TokPunc, ";")
	return ast.ReturnStmt(e)
}

func (p *parser) Expr() ast.Expr {
	return p.Assign()
}

func (p *parser) Assign() ast.Expr {
	lhs := p.LogOr()
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

func (p *parser) LogOr() ast.Expr {
	cond := p.LogAnd()
	for p.Match(lex.TokKW, "or") {
		c2 := p.LogOr()
		cond = ast.Log(cond, "or", c2)
	}
	return cond
}

func (p *parser) LogAnd() ast.Expr {
	cond := p.Equality()
	for p.Match(lex.TokKW, "and") {
		c2 := p.LogAnd()
		cond = ast.Log(cond, "and", c2)
	}
	return cond
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

	return p.Call()
}

func (p *parser) Call() ast.Expr {
	c := p.Primary()
	for p.Match(lex.TokPunc, "(") {
		if p.Match(lex.TokPunc, ")") {
			// Nothing to do
			c = ast.CallExpr(c)
		} else {
			c = ast.CallExpr(c, p.Arguments()...)
		}
	}
	return c
}

func (p *parser) Arguments() []ast.Expr {
	var as []ast.Expr
	for {
		a := p.Expr()
		as = append(as, a)
		if p.Match(lex.TokPunc, ")") {
			return as
		}
		if !p.Match(lex.TokPunc, ",") {
			panic(p.Error("unclosed argument list"))
		}
	}
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
	if p.Match(lex.TokKW, "true") {
		return ast.True
	}
	if p.Match(lex.TokKW, "false") {
		return ast.False
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
