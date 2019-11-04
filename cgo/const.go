package cgo

// This file implements a parser of a subset of the C language, just enough to
// parse common #define statements to Go constant expressions.

import (
	"go/ast"
	"go/token"
	"strings"
)

// parseConst parses the given string as a C constant.
func parseConst(pos token.Pos, value string) *ast.BasicLit {
	for len(value) != 0 && value[0] == '(' && value[len(value)-1] == ')' {
		value = strings.TrimSpace(value[1 : len(value)-1])
	}
	if len(value) == 0 {
		// Pretend it doesn't exist at all.
		return nil
	}
	// For information about integer literals:
	// https://en.cppreference.com/w/cpp/language/integer_literal
	if value[0] == '"' {
		// string constant
		return &ast.BasicLit{ValuePos: pos, Kind: token.STRING, Value: value}
	}
	if value[0] == '\'' {
		// char constant
		return &ast.BasicLit{ValuePos: pos, Kind: token.CHAR, Value: value}
	}
	// assume it's a number (int or float)
	value = strings.Replace(value, "'", "", -1) // remove ' chars
	value = strings.TrimRight(value, "lu")      // remove llu suffixes etc.
	// find the first non-number
	nonnum := byte(0)
	for i := 0; i < len(value); i++ {
		if value[i] < '0' || value[i] > '9' {
			nonnum = value[i]
			break
		}
	}
	// determine number type based on the first non-number
	switch nonnum {
	case 0:
		// no non-number found, must be an integer
		return &ast.BasicLit{ValuePos: pos, Kind: token.INT, Value: value}
	case 'x', 'X':
		// hex integer constant
		// TODO: may also be a floating point number per C++17.
		return &ast.BasicLit{ValuePos: pos, Kind: token.INT, Value: value}
	case '.', 'e':
		// float constant
		value = strings.TrimRight(value, "fFlL")
		return &ast.BasicLit{ValuePos: pos, Kind: token.FLOAT, Value: value}
	default:
		// unknown type, ignore
	}
	return nil
}
