// Package fuzz is a shim to allow compilation against Go 1.18.
// It only defines a single type to work with testing/internal/testdeps,
// and hide all the other new dependencies from this package.
package fuzz

import (
	"context"
	"errors"
	"io"
	"reflect"
	"time"
)

// CorpusEntry represents an individual input for fuzzing.
//
// We must use an equivalent type in the testing and testing/internal/testdeps
// packages, but testing can't import this package directly, and we don't want
// to export this type from testing. Instead, we use the same struct type and
// use a type alias (not a defined type) for convenience.
type CorpusEntry = struct {
	Parent string

	// Path is the path of the corpus file, if the entry was loaded from disk.
	// For other entries, including seed values provided by f.Add, Path is the
	// name of the test, e.g. seed#0 or its hash.
	Path string

	// Data is the raw input data. Data should only be populated for seed
	// values. For on-disk corpus files, Data will be nil, as it will be loaded
	// from disk using Path.
	Data []byte

	// Values is the unmarshaled values from a corpus file.
	Values []any

	Generation int

	// IsSeed indicates whether this entry is part of the seed corpus.
	IsSeed bool
}

// CoordinateFuzzingOpts is a set of arguments for CoordinateFuzzing.
// The zero value is valid for each field unless specified otherwise.
type CoordinateFuzzingOpts struct {
	// Log is a writer for logging progress messages and warnings.
	// If nil, io.Discard will be used instead.
	Log io.Writer

	// Timeout is the amount of wall clock time to spend fuzzing after the corpus
	// has loaded. If zero, there will be no time limit.
	Timeout time.Duration

	// Limit is the number of random values to generate and test. If zero,
	// there will be no limit on the number of generated values.
	Limit int64

	// MinimizeTimeout is the amount of wall clock time to spend minimizing
	// after discovering a crasher. If zero, there will be no time limit. If
	// MinimizeTimeout and MinimizeLimit are both zero, then minimization will
	// be disabled.
	MinimizeTimeout time.Duration

	// MinimizeLimit is the maximum number of calls to the fuzz function to be
	// made while minimizing after finding a crash. If zero, there will be no
	// limit. Calls to the fuzz function made when minimizing also count toward
	// Limit. If MinimizeTimeout and MinimizeLimit are both zero, then
	// minimization will be disabled.
	MinimizeLimit int64

	// parallel is the number of worker processes to run in parallel. If zero,
	// CoordinateFuzzing will run GOMAXPROCS workers.
	Parallel int

	// Seed is a list of seed values added by the fuzz target with testing.F.Add
	// and in testdata.
	Seed []CorpusEntry

	// Types is the list of types which make up a corpus entry.
	// Types must be set and must match values in Seed.
	Types []reflect.Type

	// CorpusDir is a directory where files containing values that crash the
	// code being tested may be written. CorpusDir must be set.
	CorpusDir string

	// CacheDir is a directory containing additional "interesting" values.
	// The fuzzer may derive new values from these, and may write new values here.
	CacheDir string
}

// CoordinateFuzzing creates several worker processes and communicates with
// them to test random inputs that could trigger crashes and expose bugs.
// The worker processes run the same binary in the same directory with the
// same environment variables as the coordinator process. Workers also run
// with the same arguments as the coordinator, except with the -test.fuzzworker
// flag prepended to the argument list.
//
// If a crash occurs, the function will return an error containing information
// about the crash, which can be reported to the user.
func CoordinateFuzzing(ctx context.Context, opts CoordinateFuzzingOpts) (err error) {
	return errors.New("not implemented")
}

// ReadCorpus reads the corpus from the provided dir. The returned corpus
// entries are guaranteed to match the given types. Any malformed files will
// be saved in a MalformedCorpusError and returned, along with the most recent
// error.
func ReadCorpus(dir string, types []reflect.Type) ([]CorpusEntry, error) {
	return nil, errors.New("not implemented")
}

// CheckCorpus verifies that the types in vals match the expected types
// provided.
func CheckCorpus(vals []any, types []reflect.Type) error {
	return errors.New("not implemented")
}

func ResetCoverage()    {}
func SnapshotCoverage() {}

// RunFuzzWorker is called in a worker process to communicate with the
// coordinator process in order to fuzz random inputs. RunFuzzWorker loops
// until the coordinator tells it to stop.
//
// fn is a wrapper on the fuzz function. It may return an error to indicate
// a given input "crashed". The coordinator will also record a crasher if
// the function times out or terminates the process.
//
// RunFuzzWorker returns an error if it could not communicate with the
// coordinator process.
func RunFuzzWorker(ctx context.Context, fn func(CorpusEntry) error) error {
	return errors.New("not implemented")
}
