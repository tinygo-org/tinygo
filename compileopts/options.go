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
	validOutputOptions        = []string{"none", "default", "custom", "usbcdc", "uart"}
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
	PrintStacks   bool
	CFlags        []string
	LDFlags       []string
	Tags          string
	WasmAbi       string
	HeapSize      int64
	TestConfig    TestConfig
	Programmer    string
	Output        string
}

// Verify performs a validation on the given options, raising an error if options are not valid.
func (o *Options) Verify() error {
	if o.GC != "" {
		valid := isInArray(validGCOptions, o.GC)
		if !valid {
			return fmt.Errorf(`invalid gc option '%s': valid values are %s`,
				o.GC,
				strings.Join(validGCOptions, ", "))
		}
	}

	if o.Scheduler != "" {
		valid := isInArray(validSchedulerOptions, o.Scheduler)
		if !valid {
			return fmt.Errorf(`invalid scheduler option '%s': valid values are %s`,
				o.Scheduler,
				strings.Join(validSchedulerOptions, ", "))
		}
	}

	if o.PrintSizes != "" {
		valid := isInArray(validPrintSizeOptions, o.PrintSizes)
		if !valid {
			return fmt.Errorf(`invalid size option '%s': valid values are %s`,
				o.PrintSizes,
				strings.Join(validPrintSizeOptions, ", "))
		}
	}

	if o.PanicStrategy != "" {
		valid := isInArray(validPanicStrategyOptions, o.PanicStrategy)
		if !valid {
			return fmt.Errorf(`invalid panic option '%s': valid values are %s`,
				o.PanicStrategy,
				strings.Join(validPanicStrategyOptions, ", "))
		}
	}

	if o.Output != "" {
		valid := isInArray(validOutputOptions, o.Output)
		if !valid {
			return fmt.Errorf(`invalid output option '%s': valid values are %s`,
				o.Output,
				strings.Join(validOutputOptions, ", "))
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
