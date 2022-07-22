package examples

import (
	"bytes"
	"github.com/jan-g/lox/analysis"
	"github.com/jan-g/lox/builtin"
	"github.com/jan-g/lox/eval"
	"github.com/jan-g/lox/parse"
	"github.com/stretchr/testify/assert"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func curDir() string {
	_, filename, _, _ := runtime.Caller(0)
	return filepath.Dir(filename)
}

func loxFiles(dir string) []string {
	m, _ := filepath.Glob(filepath.Join(dir, "*.lox"))
	return m
}

func TestExamples(t *testing.T) {
	dir := curDir()
	for _, f := range loxFiles(dir) {
		_, fn := filepath.Split(f)
		t.Run(fn, func(t *testing.T) {
			if err := run1(t, dir, fn); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func loadFile(t *testing.T, dir string, fn string, ext string) string {
	fn2 := strings.TrimSuffix(fn, ".lox") + ext
	f, err := os.Open(filepath.Join(dir, fn2))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	expected, err := io.ReadAll(f)
	if err != nil {
		t.Fatal(err)
	}
	return string(expected)
}

func run1(t *testing.T, dir string, fn string) (err error) {
	buf := &bytes.Buffer{}
	env := builtin.InitEnv(eval.New(buf))
	f, err := os.Open(filepath.Join(dir, fn))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	defer func() {
		if err != nil {
			expected := loadFile(t, dir, fn, ".err")
			expected = strings.Trim(expected, "\n")
			if expected == err.Error() {
				err = nil
			} else {
				assert.Equal(t, expected, err.Error())
			}
		}
	}()

	p := parse.New(f)
	ast, err := p.Parse()
	if err != nil {
		return err
	}
	if err := analysis.Analyse(ast); err != nil {
		return err
	}

	if err := env.Run(ast); err != nil {
		return err
	}

	expected := loadFile(t, dir, fn, ".out")
	if expected == buf.String() {
		return nil
	}
	assert.Equal(t, expected, buf.String())
	return nil
}
