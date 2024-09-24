package compileopts

// Scheduler options
const (
	SchedulerNone     = "none"
	SchedulerTasks    = "tasks"
	SchedulerAsyncify = "asyncify"

	// GC options
	GCNone         = "none"
	GCLeaking      = "leaking"
	GCConservative = "conservative"
	GCCustom       = "custom"
	GCPrecise      = "precise"

	// Serial options
	SerialNone = "none"
	SerialUART = "uart"
	SerialUSB  = "usb"
	SerialRTT  = "rtt"

	// Size options
	SizeNone  = "none"
	SizeShort = "short"
	SizeFull  = "full"

	// Panic strategies
	PanicPrint = "print"
	PanicTrap  = "trap"

	// Optimization Options
	OptNone = "none"
	Opt1    = "1"
	Opt2    = "2"
	Opt3    = "3"
	Opts    = "s"
	Optz    = "z"

	// Go builder Options
	TinyGoRoot  = "TINYGOROOT"
	GolangCache = "GOCACHE"

	// Well defined Golang architectures
	ArchArm64 = "armd64"
	ArchArm   = "arm"

	// Binary extension formats
	BinExtNone = ""
	BinExtWasm = "wasm"
	BinExtExe  = "exe"
	BinExtElf  = "elf"

	// Non executable binary formats
	BinFormatBin = "bin"
	BinFormatGba = "gba"
	BinFormatNro = "nro"
	BinFormatImg = "img"
	BinFormatHex = "hex"
	BinFormatUf2 = "uf2"
	BinFormatZip = "zip"

	// Operating System identifiers
	OsWindows = "windows"
	OsLinux   = "linux"

	// Programmers
	ProgOpenOCD = "openocd"
	ProgMSD     = "msd"
	ProgCommand = "command"
	ProgBMP     = "bmp"
)
