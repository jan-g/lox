package examples

import (
	"bytes"
	"fmt"
	"github.com/jan-g/lox/analysis"
	"github.com/jan-g/lox/builtin"
	"github.com/jan-g/lox/eval"
	"github.com/jan-g/lox/parse"
	"github.com/stretchr/testify/assert"
	"io"
	"io/fs"
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
	var files []string
	if err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			return nil
		}
		if filepath.Ext(d.Name()) != ".lox" {
			return nil
		}
		fmt.Println(path, d)
		files = append(files, path)
		return nil
	}); err != nil {
		panic(err)
	}
	return files
}

func TestExamples(t *testing.T) {
	d := curDir()
	for _, f := range loxFiles(d) {
		dir, fn := filepath.Split(f)
		t.Run(strings.TrimPrefix(f, d+"/"), func(t *testing.T) {
			if err := run1(t, dir, fn); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func loadFile(dir string, fn string, ext string) (string, error) {
	fn2 := strings.TrimSuffix(fn, ".lox") + ext
	f, err := os.Open(filepath.Join(dir, fn2))
	if err != nil {
		return "", err
	}
	defer f.Close()
	expected, err := io.ReadAll(f)
	if err != nil {
		return "", err
	}
	return string(expected), nil
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
			expected, e2 := loadFile(dir, fn, ".err")
			if e2 != nil {
				return
			}
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

	expected, e2 := loadFile(dir, fn, ".out")
	if e2 != nil {
		t.Fatal(e2)
	}
	if expected == buf.String() {
		return nil
	}
	assert.Equal(t, expected, buf.String())
	return nil
}
