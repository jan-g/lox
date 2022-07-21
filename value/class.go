package value

import (
	"fmt"
	"github.com/jan-g/lox/ast"
)

type Class = *_Class
type _Class struct {
	Name    string
	Methods map[string]*Closure
	Env     Env
}

func (c *_Class) String() string {
	return fmt.Sprintf("<class %s>", c.Name)
}

func (c *_Class) Arity() int {
	if m, err := c.FindMethod("init"); err == nil {
		return m.Arity()
	} else {
		return 0
	}
}

var _ Callable = &_Class{}

func MakeClass(env Env, name string, defs ...*ast.FunDef) Class {
	methods := make(map[string]*Closure)
	for _, d := range defs {
		m := MakeClosure(env, d.Params, d.Body)
		if d.Name.VarName() == "init" {
			m.IsInitialiser = true
		}
		methods[d.Name.VarName()] = m
	}
	return &_Class{
		Name:    name,
		Env:     env,
		Methods: methods,
	}
}

func (c *_Class) FindMethod(name string) (*Closure, error) {
	if m, ok := c.Methods[name]; ok {
		return m, nil
	}
	return nil, fmt.Errorf("cannot find method %s on %s", name, c)
}

func Bind(i Instance, m *Closure) *Closure {
	e2 := m.ParentEnv.Child()
	e2.Bind("this", i)
	m2 := MakeClosure(e2, m.Formals, m.Body)
	m2.IsInitialiser = m.IsInitialiser
	return m2
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

func Instantiate(c Class) (Instance, error) {
	return &_Instance{
		Class:  c,
		Fields: make(map[string]Value),
	}, nil
}

func (i *_Instance) Get(attr string) (Value, error) {
	if v, ok := i.Fields[attr]; ok {
		return v, nil
	}
	if m, err := i.Class.FindMethod(attr); err == nil {
		return Bind(i, m), nil
	} else {
		return nil, fmt.Errorf("Undefined property '%s' on %s", attr, i)
	}
}
