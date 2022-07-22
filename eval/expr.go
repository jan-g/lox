package eval

import (
	"fmt"
	"github.com/jan-g/lox/ast"
	"github.com/jan-g/lox/value"
	"io"
)

type Env struct {
	Out      io.Writer
	Parent   *Env
	Bindings map[string]value.Value
}

var _ value.Env = &Env{}

func (e *Env) Child() value.Env {
	return New(e.Out, e)
}

func (env *Env) Bind(name string, v value.Value) {
	env.Bindings[name] = v
}

func (env *Env) Lookup(depth int, name string) value.Value {
	for depth > 0 {
		env = env.Parent
		depth--
	}

	v, ok := env.Bindings[name]
	if ok {
		return v
	}
	panic(fmt.Errorf("unbound variable: %s", name))
}

func (env *Env) Assign(depth int, name string, v value.Value) {
	for depth > 0 {
		env = env.Parent
		depth--
	}

	_, ok := env.Bindings[name]
	if ok {
		env.Bindings[name] = v
		return
	}

	panic(fmt.Errorf("cannot update unbound variable: %s", name))
}

func New(out io.Writer, parent ...*Env) *Env {
	if len(parent) > 1 {
		panic("can only call New with 0 or 1 items")
	}
	var p *Env
	if len(parent) == 1 {
		p = parent[0]
	}
	return &Env{
		Out:      out,
		Parent:   p,
		Bindings: make(map[string]value.Value),
	}
}

func (env *Env) Run(e ast.Stmt) (err error) {
	defer func() {
		e := recover()
		if e == nil {
			return
		}
		err = e.(error)
	}()
	return env.Exec(e)
}

func (env *Env) Exec(s ast.Stmt) error {
	switch s := s.(type) {
	case ast.Program:
		for _, ss := range s {
			if err := env.Exec(ss); err != nil {
				return err
			}
		}
		return nil
	case *ast.Print:
		e := env.Eval(s.Expr)
		_, _ = fmt.Fprintln(env.Out, e)
		return nil
	case *ast.Expression:
		_ = env.Eval(s.Expr)
		return nil
	case *ast.VarDecl:
		v := env.Eval(s.Expr)
		env.Bind(s.VarName, v)
		return nil
	case *ast.FunDef:
		env.Bind(s.Name.VarName(), value.MakeClosure(env, s.Params, s.Body))
		return nil
	case ast.ClassDef:
		var sc value.Class
		var e2 value.Env = env
		if s.Superclass != nil {
			var ok bool
			sup := env.Eval(s.Superclass)
			sc, ok = sup.(value.Class)
			if !ok {
				return fmt.Errorf("%s is not a class", sup)
			}
			e2 = e2.Child()
			e2.Bind("super", sc)
		}
		env.Bind(s.Name.VarName(), value.MakeClass(e2, s.Name.VarName(), sc, s.Methods...))
		return nil
	case ast.Block:
		env2 := env.Child()
		for _, ss := range s {
			if err := env2.Exec(ss); err != nil {
				return err
			}
		}
		return nil
	case *ast.If:
		cond := env.Eval(s.Cond)
		if value.Truthful(cond) {
			return env.Exec(s.Then)
		} else if s.Else != nil {
			return env.Exec(s.Else)
		}
	case *ast.While:
		for {
			cond := env.Eval(s.Cond)
			if !value.Truthful(cond) {
				return nil
			}
			if err := env.Exec(s.Body); err != nil {
				return err
			}
		}
	case *ast.Return:
		var v value.Value = value.Nil
		if s.Expr != nil {
			v = env.Eval(s.Expr)
		}
		panic(WrappedReturn{Value: v})
	}
	return fmt.Errorf("unknown statement type %s", s)
}

func (env *Env) Eval(e ast.Expr) value.Value {
	switch e := e.(type) {
	case ast.StrLit:
		return value.Str(e)
	case ast.NLit:
		return value.Num(e)
	case *ast.UnOp:
		return env.UnOp(e)
	case *ast.BinOp:
		return env.BinOp(e)
	case *ast.LogOp:
		return env.LogOp(e)
	case ast.NilT:
		return value.Nil
	case ast.Bool:
		return value.Bool(e)
	case ast.Var:
		return env.Lookup(e.Depth, e.VarName())
	case ast.ThisT:
		return env.Lookup(e.Depth, e.VarName())
	case *ast.Assign:
		rhs := env.Eval(e.Rhs)
		env.Assign(e.Lhs.Depth, e.Lhs.VarName(), rhs)
		return rhs
	case *ast.Call:
		t := env.Eval(e.Callee)
		target, ok := t.(value.Callable)
		if !ok {
			panic(fmt.Errorf("target %s is not callable", t))
		}
		ps := make([]value.Value, len(e.Args))
		for i, a := range e.Args {
			ps[i] = env.Eval(a)
		}
		return env.call(target, false, ps...)

	case *ast.Get:
		t := env.Eval(e.Object)
		target, ok := t.(value.Instance)
		if !ok {
			panic(fmt.Errorf("target %s has no attributes", t))
		}
		if v, err := target.Get(e.Attribute); err != nil {
			panic(err)
		} else {
			return v
		}

	case *ast.Set:
		t := env.Eval(e.Object)
		target, ok := t.(value.Instance)
		if !ok {
			panic(fmt.Errorf("target %s has no attributes", t))
		}
		v := env.Eval(e.Rhs)
		target.Fields[e.Attribute] = v
		return v

	case *ast.Super:
		sc := env.Lookup(e.S.Depth, "super").(value.Class)
		this := env.Lookup(e.S.Depth-1, "this").(value.Instance)
		m, err := sc.FindMethod(e.Attribute)
		if err != nil {
			panic(err)
		}
		return value.Bind(this, m)
	}
	panic(fmt.Errorf("unhandled expr %s", e))
}

func (env *Env) UnOp(e *ast.UnOp) value.Value {
	switch e.Op {
	case "-":
		n := env.Eval(e.Arg).(value.Num)
		return -n
	case "!":
		return value.Bool(!value.Truthful(env.Eval(e.Arg)))
	}
	panic(fmt.Errorf("unhandled unary op %s", e))
}

func (env *Env) BinOp(e *ast.BinOp) value.Value {
	l := env.Eval(e.Left)
	r := env.Eval(e.Right)
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

func (env *Env) LogOp(e *ast.LogOp) value.Value {
	a := env.Eval(e.First)
	switch e.Op {
	case "or":
		if value.Truthful(a) {
			return a
		}
		return env.Eval(e.Second)
	case "and":
		if !value.Truthful(a) {
			return a
		}
		return env.Eval(e.Second)
	}
	panic(fmt.Errorf("unhandled binary op %s", e))
}
