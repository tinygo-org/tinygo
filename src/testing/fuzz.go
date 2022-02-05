package testing

import (
	"errors"
	"fmt"
	"reflect"
	"time"
)

// InternalFuzzTarget is an internal type but exported because it is
// cross-package; it is part of the implementation of the "go test" command.
type InternalFuzzTarget struct {
	Name string
	Fn   func(f *F)
}

// F is a type passed to fuzz tests.
//
// Fuzz tests run generated inputs against a provided fuzz target, which can
// find and report potential bugs in the code being tested.
//
// A fuzz test runs the seed corpus by default, which includes entries provided
// by (*F).Add and entries in the testdata/fuzz/<FuzzTestName> directory. After
// any necessary setup and calls to (*F).Add, the fuzz test must then call
// (*F).Fuzz to provide the fuzz target. See the testing package documentation
// for an example, and see the F.Fuzz and F.Add method documentation for
// details.
//
// *F methods can only be called before (*F).Fuzz. Once the test is
// executing the fuzz target, only (*T) methods can be used. The only *F methods
// that are allowed in the (*F).Fuzz function are (*F).Failed and (*F).Name.
type F struct {
	common
	fuzzContext *fuzzContext
	testContext *testContext

	// inFuzzFn is true when the fuzz function is running. Most F methods cannot
	// be called when inFuzzFn is true.
	inFuzzFn bool

	// corpus is a set of seed corpus entries, added with F.Add and loaded
	// from testdata.
	corpus []corpusEntry

	result     fuzzResult
	fuzzCalled bool
}

// corpusEntry is an alias to the same type as internal/fuzz.CorpusEntry.
// We use a type alias because we don't want to export this type, and we can't
// import internal/fuzz from testing.
type corpusEntry = struct {
	Parent     string
	Path       string
	Data       []byte
	Values     []interface{}
	Generation int
	IsSeed     bool
}

// Add will add the arguments to the seed corpus for the fuzz test. This will be
// a no-op if called after or within the fuzz target, and args must match the
// arguments for the fuzz target.
func (f *F) Add(args ...interface{}) {
	var values []interface{}
	for i := range args {
		if t := reflect.TypeOf(args[i]); !supportedTypes[t] {
			panic(fmt.Sprintf("testing: unsupported type to Add %v", t))
		}
		values = append(values, args[i])
	}
	f.corpus = append(f.corpus, corpusEntry{Values: values, IsSeed: true, Path: fmt.Sprintf("seed#%d", len(f.corpus))})
}

// supportedTypes represents all of the supported types which can be fuzzed.
var supportedTypes = map[reflect.Type]bool{
	reflect.TypeOf(([]byte)("")):  true,
	reflect.TypeOf((string)("")):  true,
	reflect.TypeOf((bool)(false)): true,
	reflect.TypeOf((byte)(0)):     true,
	reflect.TypeOf((rune)(0)):     true,
	reflect.TypeOf((float32)(0)):  true,
	reflect.TypeOf((float64)(0)):  true,
	reflect.TypeOf((int)(0)):      true,
	reflect.TypeOf((int8)(0)):     true,
	reflect.TypeOf((int16)(0)):    true,
	reflect.TypeOf((int32)(0)):    true,
	reflect.TypeOf((int64)(0)):    true,
	reflect.TypeOf((uint)(0)):     true,
	reflect.TypeOf((uint8)(0)):    true,
	reflect.TypeOf((uint16)(0)):   true,
	reflect.TypeOf((uint32)(0)):   true,
	reflect.TypeOf((uint64)(0)):   true,
}

// Fuzz runs the fuzz function, ff, for fuzz testing. If ff fails for a set of
// arguments, those arguments will be added to the seed corpus.
//
// ff must be a function with no return value whose first argument is *T and
// whose remaining arguments are the types to be fuzzed.
// For example:
//
//     f.Fuzz(func(t *testing.T, b []byte, i int) { ... })
//
// The following types are allowed: []byte, string, bool, byte, rune, float32,
// float64, int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64.
// More types may be supported in the future.
//
// ff must not call any *F methods, e.g. (*F).Log, (*F).Error, (*F).Skip. Use
// the corresponding *T method instead. The only *F methods that are allowed in
// the (*F).Fuzz function are (*F).Failed and (*F).Name.
//
// This function should be fast and deterministic, and its behavior should not
// depend on shared state. No mutatable input arguments, or pointers to them,
// should be retained between executions of the fuzz function, as the memory
// backing them may be mutated during a subsequent invocation. ff must not
// modify the underlying data of the arguments provided by the fuzzing engine.
//
// When fuzzing, F.Fuzz does not return until a problem is found, time runs out
// (set with -fuzztime), or the test process is interrupted by a signal. F.Fuzz
// should be called exactly once, unless F.Skip or F.Fail is called beforehand.
func (f *F) Fuzz(ff interface{}) {
	f.failed = true
	f.result.N = 0
	f.result.T = 0
	f.result.Error = errors.New("operation not implemented")
	return
}

// fuzzContext holds fields common to all fuzz tests.
type fuzzContext struct {
	deps testDeps
	mode fuzzMode
}

type fuzzMode uint8

// fuzzResult contains the results of a fuzz run.
type fuzzResult struct {
	N     int           // The number of iterations.
	T     time.Duration // The total time taken.
	Error error         // Error is the error from the failing input
}
