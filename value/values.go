package value

import "strconv"

type Value interface {
	String() string
}

type Str string

func (s Str) String() string {
	return string(s)
}

type Num float64

func (n Num) String() string {
	return strconv.FormatFloat(float64(n), 'g', -1, 64)
}

type Bool bool

func (b Bool) String() string {
	if b {
		return "true"
	}
	return "false"
}

type NilT struct{}

var Nil = NilT{}

func (NilT) String() string {
	return "nil"
}

func Truthful(v Value) bool {
	if v == Nil || v == Bool(false) {
		return false
	}
	return true
}
