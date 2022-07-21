package analysis

import (
	"fmt"
	"github.com/jan-g/lox/ast"
)

type classScope int
type functionScope int

const (
	noClass classScope = iota
	inClass

	noFun functionScope = iota
	inFun
	inInit
)

type env struct {
	classScope
	functionScope
	parent *env
	vars   map[string]struct{}
}

func makeEnv(parent *env) *env {
	e := &env{
		classScope:    noClass,
		functionScope: noFun,
		parent:        parent,
		vars:          make(map[string]struct{}),
	}
	if parent != nil {
		e.classScope = parent.classScope
		e.functionScope = parent.functionScope
	}
	return e
}

func (e *env) depth(v string) int {
	d := 0
	for e != nil {
		if _, ok := e.vars[v]; ok {
			// we found it
			return d
		}
		// look deeper
		e = e.parent
		if e == nil {
			break
		}
		d += 1
	}
	return d
}

func (e *env) bind(v string) {
	e.vars[v] = struct{}{}
}

func Analyse(stmt ast.Stmt) error {
	// Walk down a set of statements and analyse them
	return visitStmt(nil, stmt)
}

func visitStmt(e *env, s ast.Stmt) error {
	switch s := s.(type) {
	case ast.Program:
		e2 := makeEnv(e)
		for _, i := range s {
			if err := visitStmt(e2, i); err != nil {
				return err
			}
		}
		return nil
	case ast.Block:
		e2 := makeEnv(e)
		for _, i := range s {
			if err := visitStmt(e2, i); err != nil {
				return err
			}
		}
		return nil
	case *ast.Print:
		return visitExpr(e, s.Expr)
	case *ast.Expression:
		return visitExpr(e, s.Expr)
	case *ast.VarDecl:
		// We resolve the expression first, then add the binding
		if err := visitExpr(e, s.Expr); err != nil {
			return err
		}
		e.bind(s.VarName)
		return nil
	case *ast.FunDef:
		e.bind(s.Name.VarName())
		e2 := makeEnv(e)
		if e.classScope == inClass && s.Name.VarName() == "init" {
			e2.functionScope = inInit
		} else {
			e2.functionScope = inFun
		}
		for _, i := range s.Params {
			e2.bind(i.VarName())
		}
		return visitStmt(e2, s.Body)
	case *ast.If:
		if err := visitExpr(e, s.Cond); err != nil {
			return err
		}
		if err := visitStmt(e, s.Then); err != nil {
			return err
		}
		return visitStmt(e, s.Else)
	case *ast.While:
		if err := visitExpr(e, s.Cond); err != nil {
			return err
		}
		return visitStmt(e, s.Body)
	case *ast.Return:
		switch e.functionScope {
		case inFun:
			if s.Expr == nil {
				return nil
			}
			return visitExpr(e, s.Expr)
		case noFun:
			return fmt.Errorf("return not enclosed by function")
		case inInit:
			if s.Expr == nil {
				return nil
			}
			return fmt.Errorf("nonempty return not permitted in initialiser")
		default:
			return fmt.Errorf("PANIC: unrecognised function scope %d", e.functionScope)
		}
	case ast.ClassDef:
		e.bind(s.Name.VarName())
		e2 := makeEnv(e)
		e2.bind("this")
		e2.classScope = inClass
		for _, m := range s.Methods {
			if err := visitStmt(e2, m); err != nil {
				return err
			}
		}
		return nil

	default:
		return fmt.Errorf("don't know how to visit stmt %s", s)
	}
}

func visitExpr(e *env, x ast.Expr) error {
	switch x := x.(type) {
	case ast.StrLit:
		return nil
	case ast.NLit:
		return nil
	case *ast.UnOp:
		return visitExpr(e, x.Arg)
	case *ast.BinOp:
		if err := visitExpr(e, x.Left); err != nil {
			return nil
		}
		return visitExpr(e, x.Right)
	case *ast.LogOp:
		if err := visitExpr(e, x.First); err != nil {
			return nil
		}
		return visitExpr(e, x.Second)
	case ast.NilT:
		return nil
	case ast.Bool:
		return nil
	case ast.Var:
		x.Depth = e.depth(x.VarName())
		return nil
	case ast.ThisT:
		switch e.classScope {
		case inClass:
			x.Depth = e.depth(x.VarName())
			return nil
		default:
			return fmt.Errorf("'this' keyword not in class scope")
		}
	case *ast.Assign:
		if err := visitExpr(e, x.Rhs); err != nil {
			return err
		}
		return visitExpr(e, x.Lhs)
	case *ast.Call:
		if err := visitExpr(e, x.Callee); err != nil {
			return err
		}
		for _, i := range x.Args {
			if err := visitExpr(e, i); err != nil {
				return err
			}
		}
		return nil
	case *ast.Get:
		return visitExpr(e, x.Object)
	case *ast.Set:
		if err := visitExpr(e, x.Object); err != nil {
			return err
		}
		return visitExpr(e, x.Rhs)

	default:
		return fmt.Errorf("don't know how to visit expr %s", x)
	}
}
