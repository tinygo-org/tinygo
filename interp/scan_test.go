package interp

import (
	"os"
	"sort"
	"testing"

	"tinygo.org/x/go-llvm"
)

var scanTestTable = []struct {
	name            string
	severity        sideEffectSeverity
	mentionsGlobals []string
}{
	{"returnsConst", sideEffectNone, nil},
	{"returnsArg", sideEffectNone, nil},
	{"externalCallOnly", sideEffectNone, nil},
	{"externalCallAndReturn", sideEffectLimited, nil},
	{"externalCallBranch", sideEffectLimited, nil},
	{"readCleanGlobal", sideEffectNone, []string{"cleanGlobalInt"}},
	{"readDirtyGlobal", sideEffectLimited, []string{"dirtyGlobalInt"}},
	{"callFunctionPointer", sideEffectAll, []string{"functionPointer"}},
	{"getDirtyPointer", sideEffectLimited, nil},
	{"storeToPointer", sideEffectLimited, nil},
}

func TestScan(t *testing.T) {
	t.Parallel()

	// Read the input IR.
	path := "testdata/scan.ll"
	ctx := llvm.NewContext()
	buf, err := llvm.NewMemoryBufferFromFile(path)
	os.Stat(path) // make sure this file is tracked by `go test` caching
	if err != nil {
		t.Fatalf("could not read file %s: %v", path, err)
	}
	mod, err := ctx.ParseIR(buf)
	if err != nil {
		t.Fatalf("could not load module:\n%v", err)
	}

	// Check all to-be-tested functions.
	for _, tc := range scanTestTable {
		// Create an eval object, for testing.
		e := &Eval{
			Mod:          mod,
			TargetData:   llvm.NewTargetData(mod.DataLayout()),
			dirtyGlobals: map[llvm.Value]struct{}{},
		}

		// Mark some globals dirty, for testing.
		e.markDirty(mod.NamedGlobal("dirtyGlobalInt"))

		// Scan for side effects.
		fn := mod.NamedFunction(tc.name)
		if fn.IsNil() {
			t.Errorf("scan test: could not find tested function %s in the IR", tc.name)
			continue
		}
		result := e.hasSideEffects(fn)

		// Check whether the result is what we expect.
		if result.severity != tc.severity {
			t.Errorf("scan test: function %s should have severity %s but it has %s", tc.name, tc.severity, result.severity)
		}

		// Check whether the mentioned globals match with what we'd expect.
		mentionsGlobalNames := make([]string, 0, len(result.mentionsGlobals))
		for global := range result.mentionsGlobals {
			mentionsGlobalNames = append(mentionsGlobalNames, global.Name())
		}
		sort.Strings(mentionsGlobalNames)
		globalsMismatch := false
		if len(result.mentionsGlobals) != len(tc.mentionsGlobals) {
			globalsMismatch = true
		} else {
			for i, globalName := range mentionsGlobalNames {
				if tc.mentionsGlobals[i] != globalName {
					globalsMismatch = true
				}
			}
		}
		if globalsMismatch {
			t.Errorf("scan test: expected %s to mention globals %v, but it mentions globals %v", tc.name, tc.mentionsGlobals, mentionsGlobalNames)
		}
	}
}
