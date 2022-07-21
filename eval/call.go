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

func (e *Env) call(target value.Callable, initialising bool, args ...value.Value) value.Value {
	if target.Arity() != len(args) {
		panic(fmt.Errorf("%s required %d args, %d given", target, target.Arity(), len(args)))
	}

	switch target := target.(type) {
	case *builtin.Builtin:
		return target.Builtin(e, args...)

	case *value.Closure:
		if !initialising && target.IsInitialiser {
			return target.ParentEnv.Lookup(0, "this")
		}
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
		inst, err := value.Instantiate(target)
		if err != nil {
			panic(err)
		}
		if init, err := target.FindMethod("init"); err == nil {
			e.call(value.Bind(inst, init), true, args...)
		}
		return inst

	default:
		panic(fmt.Errorf("don't know how to call %s", target))
	}
}
