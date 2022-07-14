package lex

import (
    "bytes"
    "github.com/stretchr/testify/assert"
    "testing"
)

func TestBasicRead(t *testing.T) {
    r := bytes.NewReader([]byte("hello\nworld"))
    s := New(r, Lex)
    assert.Equal(t, s.Scan().Lexeme, "hello\nworld")
    assert.Equal(t, s.Scan().Token, TokEof)
}

func TestBasicTokenise(t *testing.T) {
    r := bytes.NewReader([]byte("hello\nworld\n\nHow are you?"))
    s := New(r, Alpha)
    ws := []string{}
    for {
        n := s.Scan()
        if n.Token == TokEof {
            break
        }
        t.Log("next token", n)
        assert.Equal(t, n.Token, TokId)
        ws = append(ws, n.Lexeme)
    }
    assert.Equal(t, ws, []string{"hello", "world", "How", "are", "you"})
}

func TestTokenise(t *testing.T) {
    type test struct{in string; res []string}
    var ts = []test{
        {"if\nhello+ <= < /* dwjqdwqj */ {!!=",
            []string{"KW{if}", "ID{hello}", "+", "<=", "<", "{", "!", "!="}},
        {`"one two"`, []string{`"one two"`}},
        {"1 1. 1.1", []string{"NUM{1}", "NUM{1}", ".", "NUM{1.1}"}},
        {`"abhab`, []string{"ERR{unterminated string; [0,0]-[0,6]}"}},
        {"\"abhab\n", []string{"ERR{unterminated string; [0,0]-[0,6]}"}},
        {"\"abhab\"", []string{`"abhab"`}},
        {`"ab\nhab"`, []string{`"ab\nhab"`}},
    }
    for _, tt := range ts {
        t.Run(tt.in, func(t *testing.T) {
            r := bytes.NewReader([]byte(tt.in))
            s := New(r, MakeSwitch(MakeId(Kws...), WS, Op, Num, Str))
            ws := []string{}
            for {
                n := s.Scan()
                if n.Token == TokEof {
                    break
                }
                t.Log("next token", n)
                ws = append(ws, n.String())
            }
            assert.Equal(t, ws, tt.res)
        })
    }
}