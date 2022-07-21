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
	Lookup(depth int, name string) Value
	Assign(depth int, name string, v Value)

	Run(stmt ast.Stmt) error
}

type Callable interface {
	Value
	Arity() int
}

type Closure struct {
	ParentEnv     Env
	Formals       []ast.Var
	Body          ast.Stmt
	IsInitialiser bool
}

func (c *Closure) String() string {
	return fmt.Sprintf("<closure of arity %d>", len(c.Formals))
}

func (c *Closure) Arity() int {
	return len(c.Formals)
}

var _ Callable = &Closure{}

func MakeClosure(parentEnv Env, formals []ast.Var, body ast.Stmt) *Closure {
	return &Closure{
		ParentEnv: parentEnv,
		Formals:   formals,
		Body:      body,
	}
}
