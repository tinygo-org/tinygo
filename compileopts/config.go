// Package compileopts contains the configuration for a single to-be-built
// binary.
package compileopts

// Config keeps all configuration affecting the build in a single struct.
type Config struct {
	Triple        string   // LLVM target triple, e.g. x86_64-unknown-linux-gnu (empty string means default)
	CPU           string   // LLVM CPU name, e.g. atmega328p (empty string means default)
	Features      []string // LLVM CPU features
	GOOS          string   //
	GOARCH        string   //
	GC            string   // garbage collection strategy
	Scheduler     string   // scheduler implementation ("coroutines" or "tasks")
	PanicStrategy string   // panic strategy ("print" or "trap")
	CFlags        []string // cflags to pass to cgo
	LDFlags       []string // ldflags to pass to cgo
	ClangHeaders  string   // Clang built-in header include path
	DumpSSA       bool     // dump Go SSA, for compiler debugging
	VerifyIR      bool     // run extra checks on the IR
	Debug         bool     // add debug symbols for gdb
	GOROOT        string   // GOROOT
	TINYGOROOT    string   // GOROOT for TinyGo
	GOPATH        string   // GOPATH, like `go env GOPATH`
	BuildTags     []string // build tags for TinyGo (empty means {Config.GOOS/Config.GOARCH})
	TestConfig    TestConfig
}

type TestConfig struct {
	CompileTestBinary bool
	// TODO: Filter the test functions to run, include verbose flag, etc
}
