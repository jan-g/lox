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

type If struct {
	Cond Expr
	Then Stmt
	Else Stmt
}

func (i *If) String() string {
	buf := strings.Builder{}
	buf.WriteString(fmt.Sprintf("if (%s)\n\t", i.Cond))
	buf.WriteString(i.Then.String())
	if i.Else != nil {
		buf.WriteString("else\n\t")
		buf.WriteString(i.Else.String())
	}
	return buf.String()
}

func IfStmt(cond Expr, then Stmt, otherwise Stmt) Stmt {
	return &If{
		Cond: cond,
		Then: then,
		Else: otherwise,
	}
}

type While struct {
	Cond Expr
	Body Stmt
}

func (w *While) String() string {
	buf := strings.Builder{}
	buf.WriteString(fmt.Sprintf("while (%s)\n\t", w.Cond))
	buf.WriteString(w.Body.String())
	return buf.String()
}

func WhileStmt(cond Expr, body Stmt) Stmt {
	return &While{
		Cond: cond,
		Body: body,
	}
}

type Return struct {
	Expr
}

func (r *Return) String() string {
	if r.Expr == nil {
		return "return;"
	}
	return fmt.Sprintf("return %s;", r.Expr)
}

func ReturnStmt(e Expr) Stmt {
	return &Return{
		Expr: e,
	}
}

type _ClassDef struct {
	Name       Var
	Methods    []*FunDef
	Superclass Var
}

type ClassDef = *_ClassDef

func (c *_ClassDef) String() string {
	buf := strings.Builder{}
	buf.WriteString("class ")
	buf.WriteString(c.Name.String())
	if c.Superclass != nil {
		buf.WriteString(" < ")
		buf.WriteString(c.Superclass.String())
	}
	buf.WriteString(" {\n")
	for i, p := range c.Methods {
		if i > 0 {
			buf.WriteString("\n")
		}
		buf.WriteString(p._String(""))
	}
	buf.WriteString("}\n")
	return buf.String()
}

func ClassStmt(name Var, superclass Var, methods ...*FunDef) Stmt {
	return &_ClassDef{
		Name:       name,
		Methods:    methods,
		Superclass: superclass,
	}
}
