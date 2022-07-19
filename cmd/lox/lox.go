package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"github.com/jan-g/lox/analysis"
	"github.com/jan-g/lox/builtin"
	"github.com/jan-g/lox/eval"
	"github.com/jan-g/lox/parse"
	"github.com/jan-g/lox/value"
	"io"
	"os"
)

var (
	listAst = flag.Bool("list", false, "show syntax")
)

func main() {
	flag.Parse()

	if len(flag.Args()) == 0 {
		repl()
	} else {
		run(flag.Args()...)
	}
}

func repl() {
	r := bufio.NewReader(os.Stdin)
	env := builtin.InitEnv(eval.New())

	for {
		l, err := r.ReadBytes('\n')
		if err == io.EOF {
			break
		}
		if err != nil {
			panic(err)
		}
		if err := run1(env, bytes.NewReader(l[:len(l)-1]), true); err != nil {
			_, _ = fmt.Fprintln(os.Stderr, err.Error())
		} /* else {
			fmt.Println(v)
		}*/
	}
}

func run(in ...string) {
	env := builtin.InitEnv(eval.New())
	for _, fn := range in {
		f, err := os.Open(fn)
		if err != nil {
			panic(err)
		}
		if err := run1(env, f, false); err != nil {
			_, _ = fmt.Fprintln(os.Stderr, err.Error())
		}
		_ = f.Close()
	}
}

func run1(env value.Env, in io.Reader, printAst bool) error {
	p := parse.New(in)
	ast, err := p.Parse()
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err.Error())
		return err
	}
	if ast == nil {
		return nil
	}
	if err := analysis.Analyse(ast); err != nil {
		return err
	}
	if printAst || *listAst {
		_, _ = fmt.Fprintf(os.Stderr, "%s\n", ast)
		if *listAst {
			return nil
		}
	}

	return env.Run(ast)
}
