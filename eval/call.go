package eval

import (
	"fmt"
	"github.com/jan-g/lox/builtin"
	"github.com/jan-g/lox/value"
)

type WrappedReturn struct {
	value.Value
}

func (WrappedReturn) Error() string {
	return "return not from enclosing function"
}

func (e *Env) call(target value.Callable, args ...value.Value) value.Value {
	switch target := target.(type) {
	case *builtin.Builtin:
		return target.Builtin(e, args...)

	case *value.Closure:
		e2 := target.ParentEnv.Child()
		for i, f := range target.Formals {
			e2.Bind(f.VarName(), args[i])
		}
		err := e2.Run(target.Body)
		if v, ok := err.(WrappedReturn); ok {
			return v.Value
		} else if err != nil {
			panic(err)
		}
		return value.Nil

	case value.Class:
		inst, err := value.Instantiate(target, args...)
		if err != nil {
			panic(err)
		}
		return inst

	default:
		panic(fmt.Errorf("don't know how to call %s", target))
	}
}
