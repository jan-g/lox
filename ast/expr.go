package ast

import (
	"fmt"
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

type Var string

func (v Var) String() string {
	return string(v)
}

func (v Var) VarName() string {
	return string(v)
}

func Id(name string) Expr {
	return Var(name)
}

type Assign struct {
	Lhs Var
	Rhs Expr
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
