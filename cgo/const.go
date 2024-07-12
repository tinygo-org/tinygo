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

var (
	prefixParseFns map[token.Token]func(*tokenizer) (ast.Expr, *scanner.Error)
	precedences    = map[token.Token]int{
		token.OR:  precedenceOr,
		token.XOR: precedenceXor,
		token.AND: precedenceAnd,
		token.SHL: precedenceShift,
		token.SHR: precedenceShift,
		token.ADD: precedenceAdd,
		token.SUB: precedenceAdd,
		token.MUL: precedenceMul,
		token.QUO: precedenceMul,
		token.REM: precedenceMul,
	}
)

// See: https://en.cppreference.com/w/c/language/operator_precedence
const (
	precedenceLowest = iota + 1
	precedenceOr
	precedenceXor
	precedenceAnd
	precedenceShift
	precedenceAdd
	precedenceMul
	precedencePrefix
)

func init() {
	// This must be done in an init function to avoid an initialization order
	// failure.
	prefixParseFns = map[token.Token]func(*tokenizer) (ast.Expr, *scanner.Error){
		token.IDENT:  parseIdent,
		token.INT:    parseBasicLit,
		token.FLOAT:  parseBasicLit,
		token.STRING: parseBasicLit,
		token.CHAR:   parseBasicLit,
		token.LPAREN: parseParenExpr,
		token.SUB:    parseUnaryExpr,
	}
}

// parseConst parses the given string as a C constant.
func parseConst(pos token.Pos, fset *token.FileSet, value string) (ast.Expr, *scanner.Error) {
	t := newTokenizer(pos, fset, value)
	expr, err := parseConstExpr(t, precedenceLowest)
	t.Next()
	if t.curToken != token.EOF {
		return nil, &scanner.Error{
			Pos: t.fset.Position(t.curPos),
			Msg: "unexpected token " + t.curToken.String() + ", expected end of expression",
		}
	}
	return expr, err
}

// parseConstExpr parses a stream of C tokens to a Go expression.
func parseConstExpr(t *tokenizer, precedence int) (ast.Expr, *scanner.Error) {
	if t.curToken == token.EOF {
		return nil, &scanner.Error{
			Pos: t.fset.Position(t.curPos),
			Msg: "empty constant",
		}
	}
	prefix := prefixParseFns[t.curToken]
	if prefix == nil {
		return nil, &scanner.Error{
			Pos: t.fset.Position(t.curPos),
			Msg: fmt.Sprintf("unexpected token %s", t.curToken),
		}
	}
	leftExpr, err := prefix(t)

	for t.peekToken != token.EOF && precedence < precedences[t.peekToken] {
		switch t.peekToken {
		case token.OR, token.XOR, token.AND, token.SHL, token.SHR, token.ADD, token.SUB, token.MUL, token.QUO, token.REM:
			t.Next()
			leftExpr, err = parseBinaryExpr(t, leftExpr)
		}
	}

	return leftExpr, err
}

func parseIdent(t *tokenizer) (ast.Expr, *scanner.Error) {
	return &ast.Ident{
		NamePos: t.curPos,
		Name:    "C." + t.curValue,
	}, nil
}

func parseBasicLit(t *tokenizer) (ast.Expr, *scanner.Error) {
	return &ast.BasicLit{
		ValuePos: t.curPos,
		Kind:     t.curToken,
		Value:    t.curValue,
	}, nil
}

func parseParenExpr(t *tokenizer) (ast.Expr, *scanner.Error) {
	lparen := t.curPos
	t.Next()
	x, err := parseConstExpr(t, precedenceLowest)
	if err != nil {
		return nil, err
	}
	t.Next()
	if t.curToken != token.RPAREN {
		return nil, unexpectedToken(t, token.RPAREN)
	}
	expr := &ast.ParenExpr{
		Lparen: lparen,
		X:      x,
		Rparen: t.curPos,
	}
	return expr, nil
}

func parseBinaryExpr(t *tokenizer, left ast.Expr) (ast.Expr, *scanner.Error) {
	expression := &ast.BinaryExpr{
		X:     left,
		Op:    t.curToken,
		OpPos: t.curPos,
	}
	precedence := precedences[t.curToken]
	t.Next()
	right, err := parseConstExpr(t, precedence)
	expression.Y = right
	return expression, err
}

func parseUnaryExpr(t *tokenizer) (ast.Expr, *scanner.Error) {
	expression := &ast.UnaryExpr{
		OpPos: t.curPos,
		Op:    t.curToken,
	}
	t.Next()
	x, err := parseConstExpr(t, precedencePrefix)
	expression.X = x
	return expression, err
}

// unexpectedToken returns an error of the form "unexpected token FOO, expected
// BAR".
func unexpectedToken(t *tokenizer, expected token.Token) *scanner.Error {
	return &scanner.Error{
		Pos: t.fset.Position(t.curPos),
		Msg: fmt.Sprintf("unexpected token %s, expected %s", t.curToken, expected),
	}
}

// tokenizer reads C source code and converts it to Go tokens.
type tokenizer struct {
	curPos, peekPos     token.Pos
	fset                *token.FileSet
	curToken, peekToken token.Token
	curValue, peekValue string
	buf                 string
}

// newTokenizer initializes a new tokenizer, positioned at the first token in
// the string.
func newTokenizer(start token.Pos, fset *token.FileSet, buf string) *tokenizer {
	t := &tokenizer{
		peekPos:   start,
		fset:      fset,
		buf:       buf,
		peekToken: token.ILLEGAL,
	}
	// Parse the first two tokens (cur and peek).
	t.Next()
	t.Next()
	return t
}

// Next consumes the next token in the stream. There is no return value, read
// the next token from the pos, token and value properties.
func (t *tokenizer) Next() {
	// The previous peek is now the current token.
	t.curPos = t.peekPos
	t.curToken = t.peekToken
	t.curValue = t.peekValue

	// Parse the next peek token.
	if t.peekPos != token.NoPos {
		t.peekPos += token.Pos(len(t.curValue))
	}
	for {
		if len(t.buf) == 0 {
			t.peekToken = token.EOF
			return
		}
		c := t.buf[0]
		switch {
		case c == ' ' || c == '\f' || c == '\n' || c == '\r' || c == '\t' || c == '\v':
			// Skip whitespace.
			// Based on this source, not sure whether it represents C whitespace:
			// https://en.cppreference.com/w/cpp/string/byte/isspace
			if t.peekPos != token.NoPos {
				t.peekPos++
			}
			t.buf = t.buf[1:]
		case len(t.buf) >= 2 && (string(t.buf[:2]) == "||" || string(t.buf[:2]) == "&&" || string(t.buf[:2]) == "<<" || string(t.buf[:2]) == ">>"):
			// Two-character tokens.
			switch c {
			case '&':
				t.peekToken = token.LAND
			case '|':
				t.peekToken = token.LOR
			case '<':
				t.peekToken = token.SHL
			case '>':
				t.peekToken = token.SHR
			default:
				panic("unreachable")
			}
			t.peekValue = t.buf[:2]
			t.buf = t.buf[2:]
			return
		case c == '(' || c == ')' || c == '+' || c == '-' || c == '*' || c == '/' || c == '%' || c == '&' || c == '|' || c == '^':
			// Single-character tokens.
			// TODO: ++ (increment) and -- (decrement) operators.
			switch c {
			case '(':
				t.peekToken = token.LPAREN
			case ')':
				t.peekToken = token.RPAREN
			case '+':
				t.peekToken = token.ADD
			case '-':
				t.peekToken = token.SUB
			case '*':
				t.peekToken = token.MUL
			case '/':
				t.peekToken = token.QUO
			case '%':
				t.peekToken = token.REM
			case '&':
				t.peekToken = token.AND
			case '|':
				t.peekToken = token.OR
			case '^':
				t.peekToken = token.XOR
			}
			t.peekValue = t.buf[:1]
			t.buf = t.buf[1:]
			return
		case c >= '0' && c <= '9':
			// Numeric constant (int, float, etc.).
			// Find the last non-numeric character.
			tokenLen := len(t.buf)
			hasDot := false
			for i, c := range t.buf {
				if c == '.' {
					hasDot = true
				}
				if c >= '0' && c <= '9' || c == '.' || c == '_' || c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' {
					tokenLen = i + 1
				} else {
					break
				}
			}
			t.peekValue = t.buf[:tokenLen]
			t.buf = t.buf[tokenLen:]
			if hasDot {
				// Integer constants are more complicated than this but this is
				// a close approximation.
				// https://en.cppreference.com/w/cpp/language/integer_literal
				t.peekToken = token.FLOAT
				t.peekValue = strings.TrimRight(t.peekValue, "f")
			} else {
				t.peekToken = token.INT
				t.peekValue = strings.TrimRight(t.peekValue, "uUlL")
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
			t.peekValue = t.buf[:tokenLen]
			t.buf = t.buf[tokenLen:]
			t.peekToken = token.IDENT
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
			t.peekToken = token.STRING
			t.peekValue = t.buf[:tokenLen]
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
			t.peekToken = token.CHAR
			t.peekValue = t.buf[:tokenLen]
			t.buf = t.buf[tokenLen:]
			return
		default:
			t.peekToken = token.ILLEGAL
			return
		}
	}
}
