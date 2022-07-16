package builtin

import (
	"github.com/jan-g/lox/value"
	"time"
)

type Builtin struct {
	Name    string
	NArgs   int
	Builtin func(env value.Env, ps ...value.Value) value.Value
}

var _ value.Callable = &Builtin{}

func (b *Builtin) Arity() int {
	return b.NArgs
}

func (b *Builtin) Call(env value.Env, ps ...value.Value) value.Value {
	return b.Builtin(env, ps...)
}

func (b *Builtin) String() string {
	return b.Name
}

var builtins = []*Builtin{
	{
		Name:  "clock",
		NArgs: 0,
		Builtin: func(env value.Env, ps ...value.Value) value.Value {
			sec := time.Now().Unix()
			return value.Num(sec)
		},
	},
}

func InitEnv(e value.Env) value.Env {
	for _, b := range builtins {
		e.Bind(b.Name, b)
	}
	return e
}
