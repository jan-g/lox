package ast

import "strings"

type FunDef struct {
	Name   Var
	Params []Var
	Body   Stmt
}

func (f *FunDef) String() string {
	return f._String("fun ")
}

func (f *FunDef) _String(prefix string) string {
	buf := strings.Builder{}
	buf.WriteString(prefix)
	if f.Name != nil {
		buf.WriteString(f.Name.String())
	}
	buf.WriteString("(")
	for i, p := range f.Params {
		if i > 0 {
			buf.WriteString(", ")
		}
		buf.WriteString(p.String())
	}
	buf.WriteString(") ")
	buf.WriteString(f.Body.String())
	return buf.String()
}

func FunStmt(name Var, params []Var, body Stmt) Stmt {
	return &FunDef{
		Name:   name,
		Params: params,
		Body:   body,
	}
}

type FunLit FunDef

func (f *FunLit) String() string {
	return (*FunDef)(f).String()
}

func FunExpr(name Var, params []Var, body Stmt) *FunLit {
	f := FunLit(FunDef{
		Name:   name,
		Params: params,
		Body:   body,
	})
	return &f
}
