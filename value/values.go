package value

import (
	"fmt"
	"github.com/jan-g/lox/ast"
	"strconv"
)

type Value interface {
	String() string
}

type Str string

func (s Str) String() string {
	return string(s)
}

type Num float64

func (n Num) String() string {
	return strconv.FormatFloat(float64(n), 'g', -1, 64)
}

type Bool bool

func (b Bool) String() string {
	if b {
		return "true"
	}
	return "false"
}

type NilT struct{}

var Nil = NilT{}

func (NilT) String() string {
	return "nil"
}

func Truthful(v Value) bool {
	if v == Nil || v == Bool(false) {
		return false
	}
	return true
}

type Env interface {
	Child() Env

	Bind(name string, v Value)
	Lookup(name string) Value
	Assign(name string, v Value)

	Run(stmt ast.Stmt) error
}

type Callable interface {
	Value
	Arity() int
	Call(e Env, ps ...Value) Value
}

type Closure struct {
	ParentEnv Env
	Formals   []string
	Body      ast.Stmt
}

func (c *Closure) String() string {
	return fmt.Sprintf("<closure of arity %d>", len(c.Formals))
}

func (c *Closure) Arity() int {
	return len(c.Formals)
}

type WrappedReturn struct {
	Value
}

func (WrappedReturn) Error() string {
	return "return not from enclosing function"
}

func (c *Closure) Call(e Env, ps ...Value) Value {
	e2 := c.ParentEnv.Child()
	for i, f := range c.Formals {
		e2.Bind(f, ps[i])
	}
	err := e2.Run(c.Body)
	if v, ok := err.(WrappedReturn); ok {
		return v.Value
	} else if err != nil {
		panic(err)
	}
	return Nil
}

var _ Callable = &Closure{}

func MakeClosure(parentEnv Env, formals []string, body ast.Stmt) Value {
	return &Closure{
		ParentEnv: parentEnv,
		Formals:   formals,
		Body:      body,
	}
}
