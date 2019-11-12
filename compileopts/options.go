package compileopts

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
