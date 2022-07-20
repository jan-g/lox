package value

import (
	"fmt"
	"github.com/jan-g/lox/ast"
)

type Class = *_Class
type _Class struct {
	Name    string
	Methods map[string]*ast.FunDef
	Env     Env
}

func (c *_Class) String() string {
	return fmt.Sprintf("<class %s>", c.Name)
}

func (_ *_Class) Arity() int {
	return 0
}

var _ Callable = &_Class{}

func MakeClass(env Env, name string, defs ...*ast.FunDef) Class {
	methods := make(map[string]*ast.FunDef)
	for _, d := range defs {
		methods[d.Name.VarName()] = d
	}
	return &_Class{
		Name:    name,
		Env:     env,
		Methods: methods,
	}
}

type Instance = *_Instance
type _Instance struct {
	Class  Class
	Fields map[string]Value
}

var _ Value = &_Instance{}

func (i *_Instance) String() string {
	return fmt.Sprintf("<instance %s>", i.Class.Name)
}

func Instantiate(c Class, ps ...Value) (Value, error) {
	return &_Instance{
		Class:  c,
		Fields: make(map[string]Value),
	}, nil
}
