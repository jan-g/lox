package lex

import (
	"bufio"
	"fmt"
	"io"
	"strings"
	"unicode"
	"unicode/utf8"
)

type Pos struct {
	Line int
	Col  int
}

type TokenType int

const (
	TokEof TokenType = iota
	TokErr
	TokKW
	TokId
	TokOp
	TokPunc
	TokStr
	TokNum
)

func (t TokenType) String() string {
	switch t {
	case TokEof:
		return "EOF"
	case TokKW:
		return "KW"
	case TokErr:
		return "ERR"
	case TokId:
		return "ID"
	case TokOp:
		return "OP"
	case TokPunc:
		return "PUNC"
	case TokStr:
		return "STR"
	case TokNum:
		return "NUM"
	default:
		return fmt.Sprintf("?%d", t)
	}
}

func (p Pos) String() string {
	return fmt.Sprintf("[%d,%d]", p.Line, p.Col)
}

func (t T) String() string {
	switch t.Token {
	case TokEof:
		return "EOF"
	case TokErr:
		return fmt.Sprintf("ERR{%s; %s-%s}",
			t.Lexeme, t.Start, t.End)
	case TokOp, TokPunc:
		return t.Lexeme
	case TokStr:
		return fmt.Sprintf("%q", t.Lexeme)
	default:
		if len(t.Lexeme) > 10 {
			return fmt.Sprintf("%s{%10q...}", t.Token, t.Lexeme)
		}
		return fmt.Sprintf("%s{%s}", t.Token, t.Lexeme)
	}
}

type T struct {
	Token  TokenType
	Start  Pos
	End    Pos
	Lexeme string
}

type Lexer struct {
	r       *bufio.Reader
	start   Pos
	pos     Pos
	lastPos Pos
	tokens  chan T
	rs      []rune
	state   StateFunc
	hitEof  bool
	started bool
	current T
}

type StateFunc func(l *Lexer) StateFunc

func New(r io.Reader, state StateFunc) *Lexer {
	return &Lexer{
		r:      bufio.NewReader(r),
		tokens: make(chan T, 2),
		state:  state,
	}
}

func (l *Lexer) Current() T {
	if !l.started {
		l.Scan()
	}
	return l.current
}

func (l *Lexer) Scan() T {
	if l.state != nil {
	loop:
		for {
			select {
			case t, ok := <-l.tokens:
				l.started = true
				if ok {
					l.current = t
					return t
				}
				break loop
			default:
				l.state = l.state(l)
			}
		}
	}

	return T{
		Token: TokEof,
		Start: l.pos,
		End:   l.pos,
	}
}

func (l *Lexer) Runes() string {
	return string(l.rs)
}

func (l *Lexer) Emit(t TokenType) {
	l.tokens <- T{
		Token:  t,
		Start:  l.start,
		End:    l.pos,
		Lexeme: l.Runes(),
	}
	l.Drop()
}

func (l *Lexer) Drop() {
	l.start = l.pos
	l.rs = nil
}

const eof = -1

func (l *Lexer) Next() rune {
	if l.hitEof {
		return eof
	}
	// Do we need to read more stuff in?
	r, _, err := l.r.ReadRune()
	l.lastPos = l.pos
	if err == io.EOF {
		l.hitEof = true
		l.pos.Col += 1
		return eof
	}
	l.rs = append(l.rs, r)
	if r == '\n' {
		l.pos.Col = 0
		l.pos.Line += 1
	} else {
		l.pos.Col += 1
	}
	return r
}

func (l *Lexer) Backup() {
	if l.hitEof {
		l.hitEof = false
	} else {
		if err := l.r.UnreadRune(); err != nil {
			panic(err)
		}
		if len(l.rs) > 0 {
			l.rs = l.rs[:len(l.rs)-1]
		}
	}
	l.pos = l.lastPos
}

func (l *Lexer) Peek() rune {
	r := l.Next()
	l.Backup()
	return r
}

func (l *Lexer) Peek2() (rune, rune) {
	buf, _ := l.r.Peek(16)
	r1, w := utf8.DecodeRune(buf)
	r2, _ := utf8.DecodeRune(buf[w:])
	return r1, r2
}

func Lex(l *Lexer) StateFunc {
	for l.Next() != eof {
	}
	l.Emit(TokId)
	close(l.tokens)
	return Lex
}

func Alpha(l *Lexer) StateFunc {
	for unicode.IsLetter(l.Peek()) {
		l.Next()
	}
	l.Emit(TokId)
	for !unicode.IsLetter(l.Peek()) {
		if l.Next() == eof {
			return nil
		}
	}
	l.Drop()
	return Alpha
}

type scanFunc func(l *Lexer) bool

func MakeSwitch(sfs ...scanFunc) StateFunc {
	var sw StateFunc
	sw = func(l *Lexer) StateFunc {
		for _, f := range sfs {
			if f(l) {
				return sw
			}
		}
		if l.Next() == eof {
			l.Emit(TokEof)
			return nil
		}
		l.Emit(TokErr)
		return nil
	}
	return sw
}

var (
	alphaNum = []*unicode.RangeTable{unicode.Letter, unicode.Number}
	Kws      = strings.Split("and class else false fun for if nil or print return super this true var while", " ")
)

func MakeId(kws ...string) scanFunc {
	m := make(map[string]TokenType)
	for _, k := range kws {
		m[k] = TokKW
	}
	return func(l *Lexer) bool {
		if unicode.IsLetter(l.Peek()) {
			l.Next()
			for unicode.IsOneOf(alphaNum, l.Peek()) || l.Peek() == '_' {
				l.Next()
			}
			if t, ok := m[l.Runes()]; ok {
				l.Emit(t)
			} else {
				l.Emit(TokId)
			}
			return true
		}
		return false
	}
}

func WS(l *Lexer) bool {
	if unicode.IsSpace(l.Peek()) {
		for unicode.IsSpace(l.Peek()) {
			l.Next()
		}
		l.Drop()
		return true
	}
	return false
}

func Op(l *Lexer) bool {
	c := l.Next()
	switch c {
	case '>':
		if l.Peek() == '=' {
			l.Next()
		}
		l.Emit(TokOp)
		return true
	case '<':
		if l.Peek() == '=' {
			l.Next()
		}
		l.Emit(TokOp)
		return true
	case '=':
		if l.Peek() == '=' {
			l.Next()
		}
		l.Emit(TokOp)
		return true
	case '+', '-', '*', '%':
		l.Emit(TokOp)
		return true
	case '!':
		if l.Peek() == '=' {
			l.Next()
		}
		l.Emit(TokOp)
		return true
	case '/':
		if l.Peek() == '/' {
			// Consume to \n
			for {
				c := l.Next()
				if c == '\n' || c == eof {
					break
				}
			}
			l.Drop()
			return true
		}
		if l.Peek() == '*' {
			l.Next()
			// Consume to */
			for {
				c := l.Next()
				if c == eof {
					l.rs = []rune("no closing comment")
					l.Emit(TokErr)
					return true
				}
				if c != '*' {
					continue
				}
				if l.Peek() != '/' {
					continue
				}
				l.Next()
				l.Drop()
				return true
			}
		}
		l.Emit(TokOp)
		return true
	case '{', '}', ';', '(', ')', '.', ',':
		l.Emit(TokPunc)
		return true

	}
	l.Backup()
	return false
}

func Str(l *Lexer) bool {
	if l.Peek() != '"' {
		return false
	}
	l.Next()
	for {
		c := l.Next()
		switch c {
		case eof, '\n':
			l.Backup()
			l.rs = []rune("unterminated string")
			l.Emit(TokErr)
			return true
		case '\\':
			c := l.Next()
			switch c {
			case eof, '\n':
				l.Backup()
				l.rs = []rune("unterminated string")
				l.Emit(TokErr)
				return true
			case 'n':
				c = '\n'
				fallthrough
			default:
				l.rs = l.rs[:len(l.rs)-1]
				l.rs[len(l.rs)-1] = c
			}
		case '"':
			// End of string, tidy up
			l.rs = l.rs[1 : len(l.rs)-1]
			l.Emit(TokStr)
			return true
		}
	}
}

func Num(l *Lexer) bool {
	c := l.Peek()
	if !unicode.IsDigit(c) {
		return false
	}
	l.Next()
	for {
		c = l.Peek()
		if unicode.IsDigit(c) {
			l.Next()
			continue
		}
		if c == '.' {
			_, c2 := l.Peek2()
			if unicode.IsDigit(c2) {
				l.Next()
				continue
			}
		}
		l.Emit(TokNum)
		return true
	}
}
