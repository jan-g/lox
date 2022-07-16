package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"github.com/jan-g/lox/builtin"
	"github.com/jan-g/lox/eval"
	"github.com/jan-g/lox/parse"
	"github.com/jan-g/lox/value"
	"io"
	"os"
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
		p := parse.New(bytes.NewReader(l[:len(l)-1]))
		ast, err := p.Parse()
		if err != nil {
			_, _ = fmt.Fprintln(os.Stderr, err.Error())
			continue
		}
		if ast == nil {
			break
		}
		_, _ = fmt.Fprintln(os.Stderr, ast)

		if err := env.Run(ast); err != nil {
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
		if err := run1(env, f); err != nil {
			_, _ = fmt.Fprintln(os.Stderr, err.Error())
		}
		_ = f.Close()
	}
}

func run1(env value.Env, in io.Reader) error {
	p := parse.New(in)
	ast, err := p.Parse()
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err.Error())
		return err
	}
	if ast == nil {
		return nil
	}

	return env.Run(ast)
}
