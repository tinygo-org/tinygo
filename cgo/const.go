package cgo

// This file implements a parser of a subset of the C language, just enough to
// parse common #define statements to Go constant expressions.

import (
	"fmt"
	"go/ast"
	"go/scanner"
	"go/token"
	"strings"
)

// parseConst parses the given string as a C constant.
func parseConst(pos token.Pos, fset *token.FileSet, value string) (ast.Expr, *scanner.Error) {
	t := newTokenizer(pos, fset, value)
	expr, err := parseConstExpr(t)
	if t.token != token.EOF {
		return nil, &scanner.Error{
			Pos: t.fset.Position(t.pos),
			Msg: "unexpected token " + t.token.String(),
		}
	}
	return expr, err
}

// parseConstExpr parses a stream of C tokens to a Go expression.
func parseConstExpr(t *tokenizer) (ast.Expr, *scanner.Error) {
	switch t.token {
	case token.LPAREN:
		lparen := t.pos
		t.Next()
		x, err := parseConstExpr(t)
		if err != nil {
			return nil, err
		}
		if t.token != token.RPAREN {
			return nil, unexpectedToken(t, token.RPAREN)
		}
		expr := &ast.ParenExpr{
			Lparen: lparen,
			X:      x,
			Rparen: t.pos,
		}
		t.Next()
		return expr, nil
	case token.INT, token.FLOAT, token.STRING, token.CHAR:
		expr := &ast.BasicLit{
			ValuePos: t.pos,
			Kind:     t.token,
			Value:    t.value,
		}
		t.Next()
		return expr, nil
	case token.IDENT:
		expr := &ast.Ident{
			NamePos: t.pos,
			Name:    "C." + t.value,
		}
		t.Next()
		return expr, nil
	case token.EOF:
		return nil, &scanner.Error{
			Pos: t.fset.Position(t.pos),
			Msg: "empty constant",
		}
	default:
		return nil, &scanner.Error{
			Pos: t.fset.Position(t.pos),
			Msg: fmt.Sprintf("unexpected token %s", t.token),
		}
	}
}

// unexpectedToken returns an error of the form "unexpected token FOO, expected
// BAR".
func unexpectedToken(t *tokenizer, expected token.Token) *scanner.Error {
	return &scanner.Error{
		Pos: t.fset.Position(t.pos),
		Msg: fmt.Sprintf("unexpected token %s, expected %s", t.token, expected),
	}
}

// tokenizer reads C source code and converts it to Go tokens.
type tokenizer struct {
	pos   token.Pos
	fset  *token.FileSet
	token token.Token
	value string
	buf   string
}

// newTokenizer initializes a new tokenizer, positioned at the first token in
// the string.
func newTokenizer(start token.Pos, fset *token.FileSet, buf string) *tokenizer {
	t := &tokenizer{
		pos:   start,
		fset:  fset,
		buf:   buf,
		token: token.ILLEGAL,
	}
	t.Next() // Parse the first token.
	return t
}

// Next consumes the next token in the stream. There is no return value, read
// the next token from the pos, token and value properties.
func (t *tokenizer) Next() {
	t.pos += token.Pos(len(t.value))
	for {
		if len(t.buf) == 0 {
			t.token = token.EOF
			return
		}
		c := t.buf[0]
		switch {
		case c == ' ' || c == '\f' || c == '\n' || c == '\r' || c == '\t' || c == '\v':
			// Skip whitespace.
			// Based on this source, not sure whether it represents C whitespace:
			// https://en.cppreference.com/w/cpp/string/byte/isspace
			t.pos++
			t.buf = t.buf[1:]
		case c == '(' || c == ')':
			// Single-character tokens.
			switch c {
			case '(':
				t.token = token.LPAREN
			case ')':
				t.token = token.RPAREN
			}
			t.value = t.buf[:1]
			t.buf = t.buf[1:]
			return
		case (c >= '0' && c <= '9') || c == '-':
			// Numeric constant (int, float, etc.).
			// Find the last non-numeric character.
			tokenLen := len(t.buf)
			hasDot := false
			for i, c := range t.buf {
				if c == '.' {
					hasDot = true
				}
				if c >= '0' && c <= '9' || c == '-' || c == '.' || c == '_' || c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' {
					tokenLen = i + 1
				} else {
					break
				}
			}
			t.value = t.buf[:tokenLen]
			t.buf = t.buf[tokenLen:]
			if hasDot {
				// Integer constants are more complicated than this but this is
				// a close approximation.
				// https://en.cppreference.com/w/cpp/language/integer_literal
				t.token = token.FLOAT
				t.value = strings.TrimRight(t.value, "f")
			} else {
				t.token = token.INT
				t.value = strings.TrimRight(t.value, "uUlL")
			}
			return
		case c >= 'A' && c <= 'Z' || c >= 'a' && c <= 'z' || c == '_':
			// Identifier. Find all remaining tokens that are part of this
			// identifier.
			tokenLen := len(t.buf)
			for i, c := range t.buf {
				if c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c >= 'a' && c <= 'z' || c == '_' {
					tokenLen = i + 1
				} else {
					break
				}
			}
			t.value = t.buf[:tokenLen]
			t.buf = t.buf[tokenLen:]
			t.token = token.IDENT
			return
		case c == '"':
			// String constant. Find the first '"' character that is not
			// preceded by a backslash.
			escape := false
			tokenLen := len(t.buf)
			for i, c := range t.buf {
				if i != 0 && c == '"' && !escape {
					tokenLen = i + 1
					break
				}
				if !escape {
					escape = c == '\\'
				}
			}
			t.token = token.STRING
			t.value = t.buf[:tokenLen]
			t.buf = t.buf[tokenLen:]
			return
		case c == '\'':
			// Char (rune) constant. Find the first '\'' character that is not
			// preceded by a backslash.
			escape := false
			tokenLen := len(t.buf)
			for i, c := range t.buf {
				if i != 0 && c == '\'' && !escape {
					tokenLen = i + 1
					break
				}
				if !escape {
					escape = c == '\\'
				}
			}
			t.token = token.CHAR
			t.value = t.buf[:tokenLen]
			t.buf = t.buf[tokenLen:]
			return
		default:
			t.token = token.ILLEGAL
			return
		}
	}
}
