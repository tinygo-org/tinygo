package pragmadeps

//go:inline
//go:noheap
//go:noruntime
func Definition() {}

type Register struct{}

//go:inline
//go:nobounds
func (*Register) Get() {}

//go:section .text.pragmadeps
func Section() {}

//go:interrupt
func Interrupt() {}

//go:inline
//go:noheap
//go:noruntime
func Generic[T any]() {}

//go:noescape
func Noescape(p *int)

//go:wasmimport test imported
func WasmImport()

//go:wasmexport exported
func WasmExport() {}

//go:noescape
func NoescapeDefinition(p *int) {}

//go:noescape
func NoescapeGeneric[T any]() {}

//go:inline
func InlineDeclaration()

//go:noheap
func NoheapDeclaration()

//go:noruntime
func NoruntimeDeclaration()

//go:nobounds
func NoboundsDeclaration()

//go:section .text.pragmadeps
func SectionDeclaration()

//go:interrupt
func InterruptDeclaration()

//go:wasmimport test invalid
func WasmImportDefinition() {}

//go:wasmexport invalid
func WasmExportDeclaration()

// ERROR: can only use //go:noescape on declarations
// ERROR: can only use //go:noescape on declarations
// ERROR: can only use //go:inline on definitions
// ERROR: can only use //go:noheap on definitions
// ERROR: can only use //go:noruntime on definitions
// ERROR: can only use //go:nobounds on definitions
// ERROR: can only use //go:section on definitions
// ERROR: can only use //go:interrupt on definitions
// ERROR: can only use //go:wasmimport on declarations
// ERROR: can only use //go:wasmexport on definitions
