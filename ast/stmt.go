package ast

import (
	"fmt"
	"strings"
)

type Stmt interface {
	String() string
}

type Expression struct {
	Expr
}

func (e *Expression) String() string {
	return fmt.Sprintf("%s;\n", e.Expr)
}

func ExprStmt(e Expr) Stmt {
	return &Expression{e}
}

type Print struct {
	Expr
}

func (p *Print) String() string {
	return fmt.Sprintf("print %s;\n", p.Expr)
}

func PrintStmt(e Expr) Stmt {
	return &Print{e}
}

type Program []Stmt

func (p Program) String() string {
	buf := strings.Builder{}
	for _, s := range p {
		_, _ = buf.WriteString(s.String())
	}
	return buf.String()
}

func ProgStmt(stmts ...Stmt) Stmt {
	return Program(stmts)
}

type VarDecl struct {
	VarName string
	Expr
}

func (d *VarDecl) String() string {
	return fmt.Sprintf("var %s = %s;\n", d.VarName, d.Expr)
}

func Decl(id string, e Expr) Stmt {
	return &VarDecl{
		VarName: id,
		Expr:    e,
	}
}

type Block []Stmt

func (b Block) String() string {
	buf := strings.Builder{}
	buf.WriteString("{\n")
	for _, s := range b {
		_, _ = buf.WriteString(s.String())
	}
	buf.WriteString("}\n")
	return buf.String()
}

func BlockStmt(sts ...Stmt) Stmt {
	return Block(sts)
}
