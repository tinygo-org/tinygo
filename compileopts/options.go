package compileopts

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	validGCOptions            = []string{"none", "leaking", "conservative"}
	validSchedulerOptions     = []string{"none", "tasks", "asyncify"}
	validSerialOptions        = []string{"none", "uart", "usb"}
	validPrintSizeOptions     = []string{"none", "short", "full"}
	validPanicStrategyOptions = []string{"print", "trap"}
	validOptOptions           = []string{"none", "0", "1", "2", "s", "z"}
)

// Options contains extra options to give to the compiler. These options are
// usually passed from the command line, but can also be passed in environment
// variables for example.
type Options struct {
	GOOS            string // environment variable
	GOARCH          string // environment variable
	GOARM           string // environment variable (only used with GOARCH=arm)
	Target          string
	Opt             string
	GC              string
	PanicStrategy   string
	Scheduler       string
	Serial          string
	PrintIR         bool
	DumpSSA         bool
	VerifyIR        bool
	PrintCommands   func(cmd string, args ...string)
	Semaphore       chan struct{} // -p flag controls cap
	Debug           bool
	PrintSizes      string
	PrintAllocs     *regexp.Regexp // regexp string
	PrintStacks     bool
	Tags            string
	WasmAbi         string
	GlobalValues    map[string]map[string]string // map[pkgpath]map[varname]value
	TestConfig      TestConfig
	Programmer      string
	OpenOCDCommands []string
	LLVMFeatures    string
	Directory       string
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

	if o.Serial != "" {
		valid := isInArray(validSerialOptions, o.Serial)
		if !valid {
			return fmt.Errorf(`invalid serial option '%s': valid values are %s`,
				o.Serial,
				strings.Join(validSerialOptions, ", "))
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

	if o.Opt != "" {
		if !isInArray(validOptOptions, o.Opt) {
			return fmt.Errorf("invalid -opt=%s: valid values are %s", o.Opt, strings.Join(validOptOptions, ", "))
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
