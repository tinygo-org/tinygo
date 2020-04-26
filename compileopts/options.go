package compileopts

import (
	"fmt"
	"strings"
)

<<<<<<< HEAD
<<<<<<< HEAD
<<<<<<< HEAD
<<<<<<< HEAD
=======
>>>>>>> 1c511b2... foxed lint complains and performed validation in main
=======
>>>>>>> a60b2c1... foxed lint complains and performed validation in main
var (
	validGCOptions        = []string{"none", "leaking", "extalloc", "conservative"}
	validSchedulerOptions = []string{"task", "coroutines"}
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

<<<<<<< HEAD
<<<<<<< HEAD
=======
>>>>>>> 9f24b37... moved verify method from Config to Options
=======
>>>>>>> 1c511b2... foxed lint complains and performed validation in main
=======
>>>>>>> a7f25a1... moved verify method from Config to Options
=======
>>>>>>> a60b2c1... foxed lint complains and performed validation in main
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

<<<<<<< HEAD
<<<<<<< HEAD
<<<<<<< HEAD
<<<<<<< HEAD
=======
>>>>>>> 1c511b2... foxed lint complains and performed validation in main
=======
>>>>>>> a60b2c1... foxed lint complains and performed validation in main
// Verify performs a validation on the given options, raising an error is options are not valid
// In particular:
// ErrGCInvalidOption will be reised if gc is not valid
// ErrSchedulerInvalidOption will be reised if scheduler is not valid
// ErrPrintSizeInvalidOption will be reised if size is not valid
<<<<<<< HEAD
<<<<<<< HEAD
=======
=======
>>>>>>> a7f25a1... moved verify method from Config to Options
var (
	validGCOptions        = []string{"none", "leaking", "extalloc", "conservative"}
	validSchedulerOptions = []string{"task", "coroutines"}
	validPrintSizeOptions = []string{"none", "short", "full"}
	GCInvalidOptionError  = fmt.Errorf(`invalid gc option: valid values are %s`,
		strings.Join(validGCOptions, ", "))
	SchedulerInvalidOptionError = fmt.Errorf(`invalid scheduler option: valid values are %s`,
		strings.Join(validSchedulerOptions, ", "))
	PrintSizeInvalidOptionError = fmt.Errorf(`invalid size option: valid value are %s`,
		strings.Join(validPrintSizeOptions, ", "))
)

<<<<<<< HEAD
>>>>>>> 9f24b37... moved verify method from Config to Options
=======
>>>>>>> 1c511b2... foxed lint complains and performed validation in main
=======
>>>>>>> a7f25a1... moved verify method from Config to Options
=======
>>>>>>> a60b2c1... foxed lint complains and performed validation in main
func (o *Options) Verify() error {
	if o.GC != "" {
		valid := isInArray(validGCOptions, o.GC)
		if !valid {
<<<<<<< HEAD
<<<<<<< HEAD
<<<<<<< HEAD
<<<<<<< HEAD
			return ErrGCInvalidOption
=======
			return GCInvalidOptionError
>>>>>>> 9f24b37... moved verify method from Config to Options
=======
			return ErrGCInvalidOption
>>>>>>> 1c511b2... foxed lint complains and performed validation in main
=======
			return GCInvalidOptionError
>>>>>>> a7f25a1... moved verify method from Config to Options
=======
			return ErrGCInvalidOption
>>>>>>> a60b2c1... foxed lint complains and performed validation in main
		}
	}

	if o.Scheduler != "" {
		valid := isInArray(validSchedulerOptions, o.Scheduler)
		if !valid {
<<<<<<< HEAD
<<<<<<< HEAD
<<<<<<< HEAD
<<<<<<< HEAD
			return ErrSchedulerInvalidOption
=======
			return SchedulerInvalidOptionError
>>>>>>> 9f24b37... moved verify method from Config to Options
=======
			return ErrSchedulerInvalidOption
>>>>>>> 1c511b2... foxed lint complains and performed validation in main
=======
			return SchedulerInvalidOptionError
>>>>>>> a7f25a1... moved verify method from Config to Options
=======
			return ErrSchedulerInvalidOption
>>>>>>> a60b2c1... foxed lint complains and performed validation in main
		}
	}

	if o.PrintSizes != "" {
		valid := isInArray(validPrintSizeOptions, o.PrintSizes)
		if !valid {
<<<<<<< HEAD
<<<<<<< HEAD
<<<<<<< HEAD
<<<<<<< HEAD
			return ErrPrintSizeInvalidOption
=======
			return PrintSizeInvalidOptionError
>>>>>>> 9f24b37... moved verify method from Config to Options
=======
			return ErrPrintSizeInvalidOption
>>>>>>> 1c511b2... foxed lint complains and performed validation in main
=======
			return PrintSizeInvalidOptionError
>>>>>>> a7f25a1... moved verify method from Config to Options
=======
			return ErrPrintSizeInvalidOption
>>>>>>> a60b2c1... foxed lint complains and performed validation in main
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
