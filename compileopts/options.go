package compileopts

import (
	"fmt"
	"strings"
)

var (
	validGCOptions        = []string{"none", "leaking", "extalloc", "conservative"}
	validSchedulerOptions = []string{"tasks", "coroutines"}
	validPrintSizeOptions = []string{"none", "short", "full"}
	// ErrGCInvalidOption is an error raised if gc option is not valid
	ErrGCInvalidOption = fmt.Errorf(`invalid gc option: valid values are %s`,
		strings.Join(validGCOptions, ", "))
	// ErrSchedulerInvalidOption is an error raised if scheduler option is not valid
	ErrSchedulerInvalidOption = fmt.Errorf(`invalid scheduler option: valid values are %s`,
		strings.Join(validSchedulerOptions, ", "))
	//ErrPrintSizeInvalidOption is an error raised if size option is not valid
	ErrPrintSizeInvalidOption = fmt.Errorf(`invalid size option: valid values are %s`,
		strings.Join(validPrintSizeOptions, ", "))
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

// Verify performs a validation on the given options, raising an error is options are not valid
// In particular:
// ErrGCInvalidOption will be reised if gc is not valid
// ErrSchedulerInvalidOption will be reised if scheduler is not valid
// ErrPrintSizeInvalidOption will be reised if size is not valid
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
