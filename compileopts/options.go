package compileopts

import (
	"fmt"
	"strings"
)

var (
	validGCOptions            = []string{"none", "leaking", "extalloc", "conservative"}
	validSchedulerOptions     = []string{"none", "tasks", "coroutines"}
	validPrintSizeOptions     = []string{"none", "short", "full"}
	validPanicStrategyOptions = []string{"print", "trap"}
)

var (
	// ErrGCInvalidOption is an error returned if gc option is not valid.
	ErrGCInvalidOption = fmt.Errorf(`invalid gc option: valid values are %s`,
		strings.Join(validGCOptions, ", "))
	// ErrSchedulerInvalidOption is an error returned if scheduler option is not valid.
	ErrSchedulerInvalidOption = fmt.Errorf(`invalid scheduler option: valid values are %s`,
		strings.Join(validSchedulerOptions, ", "))
	// ErrPrintSizeInvalidOption is an error returned if size option is not valid.
	ErrPrintSizeInvalidOption = fmt.Errorf(`invalid size option: valid values are %s`,
		strings.Join(validPrintSizeOptions, ", "))
	// ErrPanicStrategyInvalidOption is an error returned if panic strategy option is not valid.
	ErrPanicStrategyInvalidOption = fmt.Errorf(`invalid panic option: valid values are %s`,
		strings.Join(validPanicStrategyOptions, ", "))
)

// Options contains extra options to give to the compiler. These options are
// usually passed from the command line.
type Options struct {
	Target        string
	Opt           string
	GC            string
	PanicStrategy string
	Scheduler     string
	PrintIR       bool
	DumpSSA       bool
	VerifyIR      bool
	Debug         bool
	PrintSizes    string
	CFlags        []string
	LDFlags       []string
	Tags          string
	WasmAbi       string
	HeapSize      int64
	TestConfig    TestConfig
	Programmer    string
}

// Verify performs a validation on the given options, raising an error is options are not valid.
// In particular:
// ErrGCInvalidOption will be returned if gc option is not valid.
// ErrSchedulerInvalidOption will be returned if scheduler option is not valid.
// ErrPrintSizeInvalidOption will be rereturned if size option is not valid.
// ErrPanicStrategyInvalidOption will be returned if panic strategy option is not valid
func (o *Options) Verify() error {
	if o.GC != "" {
		valid := isInArray(validGCOptions, o.GC)
		if !valid {
			return ErrGCInvalidOption
		}
	}

	if o.Scheduler != "" {
		valid := isInArray(validSchedulerOptions, o.Scheduler)
		if !valid {
			return ErrSchedulerInvalidOption
		}
	}

	if o.PrintSizes != "" {
		valid := isInArray(validPrintSizeOptions, o.PrintSizes)
		if !valid {
			return ErrPrintSizeInvalidOption
		}
	}

	if o.PanicStrategy != "" {
		valid := isInArray(validPanicStrategyOptions, o.PanicStrategy)
		if !valid {
			return ErrPanicStrategyInvalidOption
		}
	}

	return nil
}

func isInArray(arr []string, item string) bool {
	for _, i := range arr {
		if i == item {
			return true
		}
	}
	return false
}
