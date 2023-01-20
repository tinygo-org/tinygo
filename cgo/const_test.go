package cgo

import (
	"bytes"
	"go/format"
	"go/token"
	"strings"
	"testing"
)

func TestParseConst(t *testing.T) {
	// Test converting a C constant to a Go constant.
	for _, tc := range []struct {
		C  string
		Go string
	}{
		{`5`, `5`},
		{`(5)`, `(5)`},
		{`(((5)))`, `(5)`},
		{`)`, `error: 1:1: unexpected token )`},
		{`5)`, `error: 1:2: unexpected token ), expected end of expression`},
		{"  \t)", `error: 1:4: unexpected token )`},
		{`5.8f`, `5.8`},
		{`foo`, `C.foo`},
		{``, `error: 1:1: empty constant`}, // empty constants not allowed in Go
		{`"foo"`, `"foo"`},
		{`"a\\n"`, `"a\\n"`},
		{`"a\n"`, `"a\n"`},
		{`"a\""`, `"a\""`},
		{`'a'`, `'a'`},
		{`0b10`, `0b10`},
		{`0x1234_5678`, `0x1234_5678`},
		{`5 5`, `error: 1:3: unexpected token INT, expected end of expression`}, // test for a bugfix
		// Binary operators.
		{`5+5`, `5 + 5`},
		{`5-5`, `5 - 5`},
		{`5*5`, `5 * 5`},
		{`5/5`, `5 / 5`},
		{`5%5`, `5 % 5`},
		{`5&5`, `5 & 5`},
		{`5|5`, `5 | 5`},
		{`5^5`, `5 ^ 5`},
		{`5||5`, `error: 1:2: unexpected token ||, expected end of expression`}, // logical binops aren't supported yet
		{`(5/5)`, `(5 / 5)`},
		{`1 - 2`, `1 - 2`},
		{`1 - 2 + 3`, `1 - 2 + 3`},
		{`1 - 2 * 3`, `1 - 2*3`},
		{`(1 - 2) * 3`, `(1 - 2) * 3`},
		{`1 * 2 - 3`, `1*2 - 3`},
		{`1 * (2 - 3)`, `1 * (2 - 3)`},
		// Unary operators.
		{`-5`, `-5`},
		{`-5-2`, `-5 - 2`},
		{`5 - - 2`, `5 - -2`},
	} {
		fset := token.NewFileSet()
		startPos := fset.AddFile("", -1, 1000).Pos(0)
		expr, err := parseConst(startPos, fset, tc.C)
		s := "<invalid>"
		if err != nil {
			if !strings.HasPrefix(tc.Go, "error: ") {
				t.Errorf("expected value %#v for C constant %#v but got error %#v", tc.Go, tc.C, err.Error())
				continue
			}
			s = "error: " + err.Error()
		} else if expr != nil {
			// Serialize the Go constant to a string, for more readable test
			// cases.
			buf := &bytes.Buffer{}
			err := format.Node(buf, fset, expr)
			if err != nil {
				t.Errorf("could not format expr from C constant %#v: %v", tc.C, err)
				continue
			}
			s = buf.String()
		}
		if s != tc.Go {
			t.Errorf("C constant %#v was parsed to %#v while expecting %#v", tc.C, s, tc.Go)
		}
	}
}
