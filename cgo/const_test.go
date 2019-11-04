package cgo

import (
	"bytes"
	"go/format"
	"go/token"
	"testing"
)

func TestParseConst(t *testing.T) {
	// Test converting a C constant to a Go constant.
	for _, tc := range []struct {
		C  string
		Go string
	}{
		{`5`, `5`},
		{`(5)`, `5`},
		{`(((5)))`, `5`},
		{`5.8f`, `5.8`},
		{`foo`, `<invalid>`}, // identifiers unimplemented
		{``, `<invalid>`},    // empty constants not allowed in Go
		{`"foo"`, `"foo"`},
		{`'a'`, `'a'`},
		{`0b10`, `<invalid>`}, // binary number literals unimplemented
	} {
		fset := token.NewFileSet()
		startPos := fset.AddFile("test.c", -1, 1000).Pos(0)
		expr := parseConst(startPos, tc.C)
		s := "<invalid>"
		if expr != nil {
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
