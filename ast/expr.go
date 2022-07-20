package ast

import (
	"fmt"
	"strings"
)

type Expr interface {
	String() string
}

type StrLit string

func (s StrLit) String() string {
	return fmt.Sprintf("%q", string(s))
}

func Str(s string) Expr {
	return StrLit(s)
}

type NLit float64

func (n NLit) String() string {
	return fmt.Sprintf("%g", n)
}

func Num(n float64) Expr {
	return NLit(n)
}

type BinOp struct {
	Left  Expr
	Op    string
	Right Expr
}

func (b BinOp) String() string {
	return fmt.Sprintf("(%s %s %s)", b.Left, b.Op, b.Right)
}

func Bin(l Expr, op string, r Expr) Expr {
	return &BinOp{
		Left:  l,
		Op:    op,
		Right: r,
	}
}

type UnOp struct {
	Op  string
	Arg Expr
}

func (u UnOp) String() string {
	return fmt.Sprintf("%s%s", u.Op, u.Arg)
}

func Un(op string, arg Expr) Expr {
	return &UnOp{
		Op:  op,
		Arg: arg,
	}
}

type NilT struct{}

var Nil = NilT{}

func (NilT) String() string {
	return "nil"
}

type Bool bool

func (b Bool) String() string {
	if b {
		return "true"
	} else {
		return "false"
	}
}

var True = Bool(true)
var False = Bool(false)

type _Var struct {
	Name  string
	Depth int
}

func (v *_Var) String() string {
	return fmt.Sprintf("%s@%d", v.Name, v.Depth)
}

func (v *_Var) VarName() string {
	return v.Name
}

func Id(name string) *_Var {
	return &_Var{
		Name: name,
	}
}

type Var = *_Var

type Assign struct {
	Lhs   Var
	Rhs   Expr
	Depth int // For closures
}

func (a *Assign) String() string {
	return fmt.Sprintf("(%s = %s)", a.Lhs, a.Rhs)
}

func Assignment(lhs Var, rhs Expr) Expr {
	return &Assign{
		Lhs: lhs,
		Rhs: rhs,
	}
}

type LogOp struct {
	First  Expr
	Op     string
	Second Expr
}

func (b LogOp) String() string {
	return fmt.Sprintf("(%s %s %s)", b.First, b.Op, b.Second)
}

func Log(a Expr, op string, b Expr) Expr {
	return &LogOp{
		First:  a,
		Op:     op,
		Second: b,
	}
}

type Call struct {
	Callee Expr
	Args   []Expr
}

func (c *Call) String() string {
	buf := strings.Builder{}
	buf.WriteString(c.Callee.String())
	buf.WriteRune('(')
	for i, a := range c.Args {
		if i > 0 {
			buf.WriteString(", ")
		}
		buf.WriteString(a.String())
	}
	buf.WriteRune(')')
	return buf.String()
}

func CallExpr(c Expr, as ...Expr) Expr {
	return &Call{
		Callee: c,
		Args:   as,
	}
}

type Get struct {
	Object    Expr
	Attribute string
}

func (g *Get) String() string {
	return fmt.Sprintf("%s.%s", g.Object, g.Attribute)
}

func GetAttr(obj Expr, attr string) Expr {
	return &Get{
		Object:    obj,
		Attribute: attr,
	}
}

type Set struct {
	Object    Expr
	Attribute string
	Rhs       Expr
}

func (s *Set) String() string {
	return fmt.Sprintf("%s.%s = %s", s.Object, s.Attribute, s.Rhs)
}

func SetAttr(obj Expr, attr string, expr Expr) Expr {
	return &Set{
		Object:    obj,
		Attribute: attr,
		Rhs:       expr,
	}
}
