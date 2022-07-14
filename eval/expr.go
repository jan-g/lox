package eval

import (
	"fmt"
	"github.com/jan-g/lox/ast"
	"github.com/jan-g/lox/value"
)

func Run(e ast.Expr) (_ value.Value, err error) {
	defer func() {
		e := recover()
		if e == nil {
			return
		}
		err = e.(error)
	}()
	return Eval(e), nil
}

func Eval(e ast.Expr) value.Value {
	switch e := e.(type) {
	case ast.StrLit:
		return value.Str(e)
	case ast.NLit:
		return value.Num(e)
	case *ast.UnOp:
		return UnOp(e)
	case *ast.BinOp:
		return BinOp(e)
	}
	panic(fmt.Errorf("unhandled expr %s", e))
}

func UnOp(e *ast.UnOp) value.Value {
	switch e.Op {
	case "-":
		n := Eval(e.Arg).(value.Num)
		return -n
	}
	panic(fmt.Errorf("unhandled unary op %s", e))
}

func BinOp(e *ast.BinOp) value.Value {
	l := Eval(e.Left)
	r := Eval(e.Right)
	switch e.Op {
	case "+":
		{
			l, okl := l.(value.Num)
			r, okr := r.(value.Num)
			if okl && okr {
				return l + r
			}
		}
		{
			l, okl := l.(value.Str)
			r, okr := r.(value.Str)
			if okl && okr {
				return l + r
			}
		}
	case "*":
		return l.(value.Num) * r.(value.Num)
	case "-":
		return l.(value.Num) - r.(value.Num)
	case "/":
		return l.(value.Num) / r.(value.Num)
	case "<":
		return value.Bool(l.(value.Num) < r.(value.Num))
	case "<=":
		return value.Bool(l.(value.Num) <= r.(value.Num))
	case ">":
		return value.Bool(l.(value.Num) > r.(value.Num))
	case ">=":
		return value.Bool(l.(value.Num) >= r.(value.Num))
	case "==":
		return value.Bool(l == r)
	case "!=":
		return value.Bool(l != r)
	}
	panic(fmt.Errorf("unhandled binary op %s", e))
}
