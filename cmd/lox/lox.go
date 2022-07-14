package main

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/jan-g/lox/eval"
	"github.com/jan-g/lox/parse"
	"io"
	"os"
)

func main() {
	r := bufio.NewReader(os.Stdin)
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
		fmt.Println(ast)
		v, err := eval.Run(ast)
		if err != nil {
			_, _ = fmt.Fprintln(os.Stderr, err.Error())
		} else {
			fmt.Println(v)
		}
	}
}
