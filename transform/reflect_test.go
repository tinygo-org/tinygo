package transform_test

import (
	"testing"

	"github.com/tinygo-org/tinygo/transform"
	"tinygo.org/x/go-llvm"
)

type reflectAssert struct {
	call           llvm.Value
	name           string
	expectedNumber uint64
}

// Test reflect lowering. This code looks at IR like this:
//
//	call void @main.assertType(i32 ptrtoint (%runtime.typecodeID* @"reflect/types.type:basic:int" to i32), i8* inttoptr (i32 3 to i8*), i32 4, i8* undef, i8* undef)
//
// and verifies that the ptrtoint constant (the first parameter of
// @main.assertType) is replaced with the correct type code.  The expected
// output is this:
//
//	call void @main.assertType(i32 4, i8* inttoptr (i32 3 to i8*), i32 4, i8* undef, i8* undef)
//
// The first and third parameter are compared and must match, the second
// parameter is ignored.
func TestReflect(t *testing.T) {
	t.Parallel()

	mod := compileGoFileForTesting(t, "./testdata/reflect.go")

	// Run the instcombine pass, to clean up the IR a bit (especially
	// insertvalue/extractvalue instructions).
	pm := llvm.NewPassManager()
	defer pm.Dispose()
	pm.AddInstructionCombiningPass()
	pm.Run(mod)

	// Get a list of all the asserts in the source code.
	assertType := mod.NamedFunction("main.assertType")
	var asserts []reflectAssert
	for user := assertType.FirstUse(); !user.IsNil(); user = user.NextUse() {
		use := user.User()
		if use.IsACallInst().IsNil() {
			t.Fatal("expected call use of main.assertType")
		}
		global := use.Operand(0).Operand(0)
		expectedNumber := use.Operand(2).ZExtValue()
		asserts = append(asserts, reflectAssert{
			call:           use,
			name:           global.Name(),
			expectedNumber: expectedNumber,
		})
	}

	// Sanity check to show that the test is actually testing anything.
	if len(asserts) < 3 {
		t.Errorf("expected at least 3 test cases, got %d", len(asserts))
	}

	// Now lower the type codes.
	transform.LowerReflect(mod)

	// Check whether the values are as expected.
	for _, assert := range asserts {
		actualNumberValue := assert.call.Operand(0)
		if actualNumberValue.IsAConstantInt().IsNil() {
			t.Errorf("expected to see a constant for %s, got something else", assert.name)
			continue
		}
		actualNumber := actualNumberValue.ZExtValue()
		if actualNumber != assert.expectedNumber {
			t.Errorf("%s: expected number 0b%b, got 0b%b", assert.name, assert.expectedNumber, actualNumber)
		}
	}
}
