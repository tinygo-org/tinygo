package riscv

// This file lists constants for CSR operations and defines methods on CSRs that
// are implemented as compiler intrinsics.

// CSR constants are used for use in CSR (Control and Status Register) compiler
// intrinsics.
type CSR int16

// Get returns the value of the given CSR.
func (csr CSR) Get() uintptr

// Set stores a new value in the given CSR.
func (csr CSR) Set(uintptr)

// SetBits atomically sets the given bits in this ISR and returns the old value.
func (csr CSR) SetBits(uintptr) uintptr

// ClearBits atomically clears the given bits in this ISR and returns the old
// value.
func (csr CSR) ClearBits(uintptr) uintptr

// CSR values defined in the RISC-V privileged specification. Not all values may
// be available on any given chip.
//
// Source: https://github.com/riscv/riscv-isa-manual/blob/riscv-priv-1.10/src/priv-csrs.tex
const (
	// User Trap Setup
	USTATUS CSR = 0x000 // User status register.
	UIE     CSR = 0x004 // User interrupt-enable register.
	UTVEC   CSR = 0x005 // User trap handler base address.

	// User Trap Handling
	USCRATCH CSR = 0x040 // Scratch register for user trap handlers.
	UEPC     CSR = 0x041 // User exception program counter.
	UCAUSE   CSR = 0x042 // User trap cause.
	UTVAL    CSR = 0x043 // User bad address or instruction.
	UIP      CSR = 0x044 // User interrupt pending.

	// User Floating-Point CSRs
	FFLAGS CSR = 0x001 // Floating-Point Accrued Exceptions.
	FRM    CSR = 0x002 // Floating-Point Dynamic Rounding Mode.
	FCSR   CSR = 0x003 // Floating-Point Control and Status

	// User Counter/Timers
	CYCLE         CSR = 0xC00 // Cycle counter for RDCYCLE instruction.
	TIME          CSR = 0xC01 // Timer for RDTIME instruction.
	INSTRET       CSR = 0xC02 // Instructions-retired counter for RDINSTRET instruction.
	HPMCOUNTER3   CSR = 0xC03 // Performance-monitoring counter 3.
	HPMCOUNTER4   CSR = 0xC04 // Performance-monitoring counter 4.
	HPMCOUNTER5   CSR = 0xC05 // Performance-monitoring counter 5.
	HPMCOUNTER6   CSR = 0xC06 // Performance-monitoring counter 6.
	HPMCOUNTER7   CSR = 0xC07 // Performance-monitoring counter 7.
	HPMCOUNTER8   CSR = 0xC08 // Performance-monitoring counter 8.
	HPMCOUNTER9   CSR = 0xC09 // Performance-monitoring counter 9.
	HPMCOUNTER10  CSR = 0xC0A // Performance-monitoring counter 10.
	HPMCOUNTER11  CSR = 0xC0B // Performance-monitoring counter 11.
	HPMCOUNTER12  CSR = 0xC0C // Performance-monitoring counter 12.
	HPMCOUNTER13  CSR = 0xC0D // Performance-monitoring counter 13.
	HPMCOUNTER14  CSR = 0xC0E // Performance-monitoring counter 14.
	HPMCOUNTER15  CSR = 0xC0F // Performance-monitoring counter 15.
	HPMCOUNTER16  CSR = 0xC10 // Performance-monitoring counter 16.
	HPMCOUNTER17  CSR = 0xC11 // Performance-monitoring counter 17.
	HPMCOUNTER18  CSR = 0xC12 // Performance-monitoring counter 18.
	HPMCOUNTER19  CSR = 0xC13 // Performance-monitoring counter 19.
	HPMCOUNTER20  CSR = 0xC14 // Performance-monitoring counter 20.
	HPMCOUNTER21  CSR = 0xC15 // Performance-monitoring counter 21.
	HPMCOUNTER22  CSR = 0xC16 // Performance-monitoring counter 22.
	HPMCOUNTER23  CSR = 0xC17 // Performance-monitoring counter 23.
	HPMCOUNTER24  CSR = 0xC18 // Performance-monitoring counter 24.
	HPMCOUNTER25  CSR = 0xC19 // Performance-monitoring counter 25.
	HPMCOUNTER26  CSR = 0xC1A // Performance-monitoring counter 26.
	HPMCOUNTER27  CSR = 0xC1B // Performance-monitoring counter 27.
	HPMCOUNTER28  CSR = 0xC1C // Performance-monitoring counter 28.
	HPMCOUNTER29  CSR = 0xC1D // Performance-monitoring counter 29.
	HPMCOUNTER30  CSR = 0xC1E // Performance-monitoring counter 30.
	HPMCOUNTER31  CSR = 0xC1F // Performance-monitoring counter 31.
	CYCLEH        CSR = 0xC80 // Upper 32 bits of CYCLE, RV32I only.
	TIMEH         CSR = 0xC81 // Upper 32 bits of TIME, RV32I only.
	INSTRETH      CSR = 0xC82 // Upper 32 bits of INSTRET, RV32I only.
	HPMCOUNTER3H  CSR = 0xC83 // Upper 32 bits of HPMCOUNTER3, RV32I only.
	HPMCOUNTER4H  CSR = 0xC84 // Upper 32 bits of HPMCOUNTER4, RV32I only.
	HPMCOUNTER5H  CSR = 0xC85 // Upper 32 bits of HPMCOUNTER5, RV32I only.
	HPMCOUNTER6H  CSR = 0xC86 // Upper 32 bits of HPMCOUNTER6, RV32I only.
	HPMCOUNTER7H  CSR = 0xC87 // Upper 32 bits of HPMCOUNTER7, RV32I only.
	HPMCOUNTER8H  CSR = 0xC88 // Upper 32 bits of HPMCOUNTER8, RV32I only.
	HPMCOUNTER9H  CSR = 0xC89 // Upper 32 bits of HPMCOUNTER9, RV32I only.
	HPMCOUNTER10H CSR = 0xC8A // Upper 32 bits of HPMCOUNTER10, RV32I only.
	HPMCOUNTER11H CSR = 0xC8B // Upper 32 bits of HPMCOUNTER11, RV32I only.
	HPMCOUNTER12H CSR = 0xC8C // Upper 32 bits of HPMCOUNTER12, RV32I only.
	HPMCOUNTER13H CSR = 0xC8D // Upper 32 bits of HPMCOUNTER13, RV32I only.
	HPMCOUNTER14H CSR = 0xC8E // Upper 32 bits of HPMCOUNTER14, RV32I only.
	HPMCOUNTER15H CSR = 0xC8F // Upper 32 bits of HPMCOUNTER15, RV32I only.
	HPMCOUNTER16H CSR = 0xC90 // Upper 32 bits of HPMCOUNTER16, RV32I only.
	HPMCOUNTER17H CSR = 0xC91 // Upper 32 bits of HPMCOUNTER17, RV32I only.
	HPMCOUNTER18H CSR = 0xC92 // Upper 32 bits of HPMCOUNTER18, RV32I only.
	HPMCOUNTER19H CSR = 0xC93 // Upper 32 bits of HPMCOUNTER19, RV32I only.
	HPMCOUNTER20H CSR = 0xC94 // Upper 32 bits of HPMCOUNTER20, RV32I only.
	HPMCOUNTER21H CSR = 0xC95 // Upper 32 bits of HPMCOUNTER21, RV32I only.
	HPMCOUNTER22H CSR = 0xC96 // Upper 32 bits of HPMCOUNTER22, RV32I only.
	HPMCOUNTER23H CSR = 0xC97 // Upper 32 bits of HPMCOUNTER23, RV32I only.
	HPMCOUNTER24H CSR = 0xC98 // Upper 32 bits of HPMCOUNTER24, RV32I only.
	HPMCOUNTER25H CSR = 0xC99 // Upper 32 bits of HPMCOUNTER25, RV32I only.
	HPMCOUNTER26H CSR = 0xC9A // Upper 32 bits of HPMCOUNTER26, RV32I only.
	HPMCOUNTER27H CSR = 0xC9B // Upper 32 bits of HPMCOUNTER27, RV32I only.
	HPMCOUNTER28H CSR = 0xC9C // Upper 32 bits of HPMCOUNTER28, RV32I only.
	HPMCOUNTER29H CSR = 0xC9D // Upper 32 bits of HPMCOUNTER29, RV32I only.
	HPMCOUNTER30H CSR = 0xC9E // Upper 32 bits of HPMCOUNTER30, RV32I only.
	HPMCOUNTER31H CSR = 0xC9F // Upper 32 bits of HPMCOUNTER31, RV32I only.

	// Supervisor Trap Setup
	SSTATUS    CSR = 0x100 // Supervisor status register.
	SEDELEG    CSR = 0x102 // Supervisor exception delegation register.
	SIDELEG    CSR = 0x103 // Supervisor interrupt delegation register.
	SIE        CSR = 0x104 // Supervisor interrupt-enable register.
	STVEC      CSR = 0x105 // Supervisor trap handler base address.
	SCOUNTEREN CSR = 0x106 // Supervisor counter enable.

	// Supervisor Trap Handling
	SSCRATCH CSR = 0x140 // Scratch register for supervisor trap handlers.
	SEPC     CSR = 0x141 // Supervisor exception program counter.
	SCAUSE   CSR = 0x142 // Supervisor trap cause.
	STVAL    CSR = 0x143 // Supervisor bad address or instruction.
	SIP      CSR = 0x144 // Supervisor interrupt pending.

	// Supervisor Protection and Translation
	SATP CSR = 0x180 // Supervisor address translation and protection.

	// Machine Information Registers
	MVENDORID CSR = 0xF11 // Vendor ID.
	MARCHID   CSR = 0xF12 // Architecture ID.
	MIMPID    CSR = 0xF13 // Implementation ID.
	MHARTID   CSR = 0xF14 // Hardware thread ID.

	// Machine Trap Setup
	MSTATUS    CSR = 0x300 // Machine status register.
	MISA       CSR = 0x301 // ISA and extensions
	MEDELEG    CSR = 0x302 // Machine exception delegation register.
	MIDELEG    CSR = 0x303 // Machine interrupt delegation register.
	MIE        CSR = 0x304 // Machine interrupt-enable register.
	MTVEC      CSR = 0x305 // Machine trap-handler base address.
	MCOUNTEREN CSR = 0x306 // Machine counter enable.

	// Machine Trap Handling
	MSCRATCH CSR = 0x340 // Scratch register for machine trap handlers.
	MEPC     CSR = 0x341 // Machine exception program counter.
	MCAUSE   CSR = 0x342 // Machine trap cause.
	MTVAL    CSR = 0x343 // Machine bad address or instruction.
	MIP      CSR = 0x344 // Machine interrupt pending.

	// Machine Protection and Translation
	PMPCFG0   CSR = 0x3A0 // Physical memory protection configuration.
	PMPCFG1   CSR = 0x3A1 // Physical memory protection configuration, RV32 only.
	PMPCFG2   CSR = 0x3A2 // Physical memory protection configuration.
	PMPCFG3   CSR = 0x3A3 // Physical memory protection configuration, RV32 only.
	PMPADDR0  CSR = 0x3B0 // Physical memory protection address register 0.
	PMPADDR1  CSR = 0x3B1 // Physical memory protection address register 1.
	PMPADDR2  CSR = 0x3B2 // Physical memory protection address register 2.
	PMPADDR3  CSR = 0x3B3 // Physical memory protection address register 3.
	PMPADDR4  CSR = 0x3B4 // Physical memory protection address register 4.
	PMPADDR5  CSR = 0x3B5 // Physical memory protection address register 5.
	PMPADDR6  CSR = 0x3B6 // Physical memory protection address register 6.
	PMPADDR7  CSR = 0x3B7 // Physical memory protection address register 7.
	PMPADDR8  CSR = 0x3B8 // Physical memory protection address register 8.
	PMPADDR9  CSR = 0x3B9 // Physical memory protection address register 9.
	PMPADDR10 CSR = 0x3BA // Physical memory protection address register 10.
	PMPADDR11 CSR = 0x3BB // Physical memory protection address register 11.
	PMPADDR12 CSR = 0x3BC // Physical memory protection address register 12.
	PMPADDR13 CSR = 0x3BD // Physical memory protection address register 13.
	PMPADDR14 CSR = 0x3BE // Physical memory protection address register 14.
	PMPADDR15 CSR = 0x3BF // Physical memory protection address register 15.

	// Machine Counter/Timers
	mcycle         CSR = 0xB00 // Machine cycle counter.
	minstret       CSR = 0xB02 // Machine instructions-retired counter.
	MHPMCOUNTER3   CSR = 0xB03 // Machine performance-monitoring counter 3.
	MHPMCOUNTER4   CSR = 0xB04 // Machine performance-monitoring counter 4.
	MHPMCOUNTER5   CSR = 0xB05 // Machine performance-monitoring counter 5.
	MHPMCOUNTER6   CSR = 0xB06 // Machine performance-monitoring counter 6.
	MHPMCOUNTER7   CSR = 0xB07 // Machine performance-monitoring counter 7.
	MHPMCOUNTER8   CSR = 0xB08 // Machine performance-monitoring counter 8.
	MHPMCOUNTER9   CSR = 0xB09 // Machine performance-monitoring counter 9.
	MHPMCOUNTER10  CSR = 0xB0A // Machine performance-monitoring counter 10.
	MHPMCOUNTER11  CSR = 0xB0B // Machine performance-monitoring counter 11.
	MHPMCOUNTER12  CSR = 0xB0C // Machine performance-monitoring counter 12.
	MHPMCOUNTER13  CSR = 0xB0D // Machine performance-monitoring counter 13.
	MHPMCOUNTER14  CSR = 0xB0E // Machine performance-monitoring counter 14.
	MHPMCOUNTER15  CSR = 0xB0F // Machine performance-monitoring counter 15.
	MHPMCOUNTER16  CSR = 0xB10 // Machine performance-monitoring counter 16.
	MHPMCOUNTER17  CSR = 0xB11 // Machine performance-monitoring counter 17.
	MHPMCOUNTER18  CSR = 0xB12 // Machine performance-monitoring counter 18.
	MHPMCOUNTER19  CSR = 0xB13 // Machine performance-monitoring counter 19.
	MHPMCOUNTER20  CSR = 0xB14 // Machine performance-monitoring counter 20.
	MHPMCOUNTER21  CSR = 0xB15 // Machine performance-monitoring counter 21.
	MHPMCOUNTER22  CSR = 0xB16 // Machine performance-monitoring counter 22.
	MHPMCOUNTER23  CSR = 0xB17 // Machine performance-monitoring counter 23.
	MHPMCOUNTER24  CSR = 0xB18 // Machine performance-monitoring counter 24.
	MHPMCOUNTER25  CSR = 0xB19 // Machine performance-monitoring counter 25.
	MHPMCOUNTER26  CSR = 0xB1A // Machine performance-monitoring counter 26.
	MHPMCOUNTER27  CSR = 0xB1B // Machine performance-monitoring counter 27.
	MHPMCOUNTER28  CSR = 0xB1C // Machine performance-monitoring counter 28.
	MHPMCOUNTER29  CSR = 0xB1D // Machine performance-monitoring counter 29.
	MHPMCOUNTER30  CSR = 0xB1E // Machine performance-monitoring counter 30.
	MHPMCOUNTER31  CSR = 0xB1F // Machine performance-monitoring counter 31.
	MCYCLEH        CSR = 0xB80 // Upper 32 bits of MCYCLE, RV32I only.
	MINSTRETH      CSR = 0xB82 // Upper 32 bits of MINSTRET, RV32I only.
	MHPMCOUNTER3H  CSR = 0xB83 // Upper 32 bits of MHPMCOUNTER3, RV32I only.
	MHPMCOUNTER4H  CSR = 0xB84 // Upper 32 bits of MHPMCOUNTER4, RV32I only.
	MHPMCOUNTER5H  CSR = 0xB85 // Upper 32 bits of MHPMCOUNTER5, RV32I only.
	MHPMCOUNTER6H  CSR = 0xB86 // Upper 32 bits of MHPMCOUNTER6, RV32I only.
	MHPMCOUNTER7H  CSR = 0xB87 // Upper 32 bits of MHPMCOUNTER7, RV32I only.
	MHPMCOUNTER8H  CSR = 0xB88 // Upper 32 bits of MHPMCOUNTER8, RV32I only.
	MHPMCOUNTER9H  CSR = 0xB89 // Upper 32 bits of MHPMCOUNTER9, RV32I only.
	MHPMCOUNTER10H CSR = 0xB8A // Upper 32 bits of MHPMCOUNTER10, RV32I only.
	MHPMCOUNTER11H CSR = 0xB8B // Upper 32 bits of MHPMCOUNTER11, RV32I only.
	MHPMCOUNTER12H CSR = 0xB8C // Upper 32 bits of MHPMCOUNTER12, RV32I only.
	MHPMCOUNTER13H CSR = 0xB8D // Upper 32 bits of MHPMCOUNTER13, RV32I only.
	MHPMCOUNTER14H CSR = 0xB8E // Upper 32 bits of MHPMCOUNTER14, RV32I only.
	MHPMCOUNTER15H CSR = 0xB8F // Upper 32 bits of MHPMCOUNTER15, RV32I only.
	MHPMCOUNTER16H CSR = 0xB90 // Upper 32 bits of MHPMCOUNTER16, RV32I only.
	MHPMCOUNTER17H CSR = 0xB91 // Upper 32 bits of MHPMCOUNTER17, RV32I only.
	MHPMCOUNTER18H CSR = 0xB92 // Upper 32 bits of MHPMCOUNTER18, RV32I only.
	MHPMCOUNTER19H CSR = 0xB93 // Upper 32 bits of MHPMCOUNTER19, RV32I only.
	MHPMCOUNTER20H CSR = 0xB94 // Upper 32 bits of MHPMCOUNTER20, RV32I only.
	MHPMCOUNTER21H CSR = 0xB95 // Upper 32 bits of MHPMCOUNTER21, RV32I only.
	MHPMCOUNTER22H CSR = 0xB96 // Upper 32 bits of MHPMCOUNTER22, RV32I only.
	MHPMCOUNTER23H CSR = 0xB97 // Upper 32 bits of MHPMCOUNTER23, RV32I only.
	MHPMCOUNTER24H CSR = 0xB98 // Upper 32 bits of MHPMCOUNTER24, RV32I only.
	MHPMCOUNTER25H CSR = 0xB99 // Upper 32 bits of MHPMCOUNTER25, RV32I only.
	MHPMCOUNTER26H CSR = 0xB9A // Upper 32 bits of MHPMCOUNTER26, RV32I only.
	MHPMCOUNTER27H CSR = 0xB9B // Upper 32 bits of MHPMCOUNTER27, RV32I only.
	MHPMCOUNTER28H CSR = 0xB9C // Upper 32 bits of MHPMCOUNTER28, RV32I only.
	MHPMCOUNTER29H CSR = 0xB9D // Upper 32 bits of MHPMCOUNTER29, RV32I only.
	MHPMCOUNTER30H CSR = 0xB9E // Upper 32 bits of MHPMCOUNTER30, RV32I only.
	MHPMCOUNTER31H CSR = 0xB9F // Upper 32 bits of MHPMCOUNTER31, RV32I only.

	// Machine Counter Setup
	MHPMEVENT4  CSR = 0x324 // Machine performance-monitoring event selector 4.
	MHPMEVENT5  CSR = 0x325 // Machine performance-monitoring event selector 5.
	MHPMEVENT6  CSR = 0x326 // Machine performance-monitoring event selector 6.
	MHPMEVENT7  CSR = 0x327 // Machine performance-monitoring event selector 7.
	MHPMEVENT8  CSR = 0x328 // Machine performance-monitoring event selector 8.
	MHPMEVENT9  CSR = 0x329 // Machine performance-monitoring event selector 9.
	MHPMEVENT10 CSR = 0x32A // Machine performance-monitoring event selector 10.
	MHPMEVENT11 CSR = 0x32B // Machine performance-monitoring event selector 11.
	MHPMEVENT12 CSR = 0x32C // Machine performance-monitoring event selector 12.
	MHPMEVENT13 CSR = 0x32D // Machine performance-monitoring event selector 13.
	MHPMEVENT14 CSR = 0x32E // Machine performance-monitoring event selector 14.
	MHPMEVENT15 CSR = 0x32F // Machine performance-monitoring event selector 15.
	MHPMEVENT16 CSR = 0x330 // Machine performance-monitoring event selector 16.
	MHPMEVENT17 CSR = 0x331 // Machine performance-monitoring event selector 17.
	MHPMEVENT18 CSR = 0x332 // Machine performance-monitoring event selector 18.
	MHPMEVENT19 CSR = 0x333 // Machine performance-monitoring event selector 19.
	MHPMEVENT20 CSR = 0x334 // Machine performance-monitoring event selector 20.
	MHPMEVENT21 CSR = 0x335 // Machine performance-monitoring event selector 21.
	MHPMEVENT22 CSR = 0x336 // Machine performance-monitoring event selector 22.
	MHPMEVENT23 CSR = 0x337 // Machine performance-monitoring event selector 23.
	MHPMEVENT24 CSR = 0x338 // Machine performance-monitoring event selector 24.
	MHPMEVENT25 CSR = 0x339 // Machine performance-monitoring event selector 25.
	MHPMEVENT26 CSR = 0x33A // Machine performance-monitoring event selector 26.
	MHPMEVENT27 CSR = 0x33B // Machine performance-monitoring event selector 27.
	MHPMEVENT28 CSR = 0x33C // Machine performance-monitoring event selector 28.
	MHPMEVENT29 CSR = 0x33D // Machine performance-monitoring event selector 29.
	MHPMEVENT30 CSR = 0x33E // Machine performance-monitoring event selector 30.
	MHPMEVENT31 CSR = 0x33F // Machine performance-monitoring event selector 31.

	// Debug/Trace Registers (shared with Debug Mode)
	TSELECT CSR = 0x7A0 // Debug/Trace trigger register select.
	TDATA1  CSR = 0x7A1 // First Debug/Trace trigger data register.
	TDATA2  CSR = 0x7A2 // Second Debug/Trace trigger data register.
	TDATA3  CSR = 0x7A3 // Third Debug/Trace trigger data register.

	// Debug Mode Registers
	DCSR     CSR = 0x7B0 // Debug control and status register.
	DPC      CSR = 0x7B1 // Debug PC.
	DSCRATCH CSR = 0x7B2 // Debug scratch register.
)
